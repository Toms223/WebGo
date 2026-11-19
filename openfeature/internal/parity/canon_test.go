package parity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/Toms223/WebGo/openfeature"
)

func TestCanonicalJSON(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		{"brief map", map[string]any{"b": 1.0, "a": []any{int64(2), math.Copysign(0, -1), 1.5}}, `{"a":[2,0,1.5],"b":1}`},
		{"time", time.Date(2026, 1, 2, 3, 4, 5, 6e6, time.UTC), `"2026-01-02T03:04:05.006Z"`},
		{"time non-utc", time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("x", 3600)), `"2026-01-02T02:04:05.000Z"`},
		{"1e21", 1e21, `1e+21`},
		{"1e20", 1e20, `100000000000000000000`},
		{"1e-7", 1e-7, `1e-7`},
		{"1.5e-7", 1.5e-7, `1.5e-7`},
		{"1e-6", 0.000001, `0.000001`},
		{"frac", 0.30000000000000004, `0.30000000000000004`},
		{"neg", -0.0025, `-0.0025`},
		{"large", 123456789012.0, `123456789012`},
		{"max", math.MaxFloat64, `1.7976931348623157e+308`},
		{"min", 5e-324, `5e-324`},
		{"neg exp", -1.25e-10, `-1.25e-10`},
		{"neg zero", math.Copysign(0, -1), `0`},
		{"nan", math.NaN(), `null`},
		{"inf", math.Inf(1), `null`},
		{"int", 3, `3`},
		{"uint8", uint8(7), `7`},
		{"float32", float32(0.5), `0.5`},
		{"nil map field omitted", map[string]any{"a": map[string]any(nil), "b": nil, "c": []string(nil)}, `{"b":null}`},
		{"empty map kept", map[string]any{"a": map[string]any{}, "b": []any{}}, `{"a":{},"b":[]}`},
		{"utf16 key order", map[string]any{"ﬁ": 1, "\U0001F600": 2, "a": 3}, `{"a":3,"😀":2,"ﬁ":1}`},
		{"no html escape", "<&>", `"<&>"`},
		{"u2028 raw", "a b c", "\"a b c\""},
		{"control escapes", "\"\\\b\f\n\r\t\x01\x1f\x7f", `"\"\\\b\f\n\r\t\u0001\u001f` + "\x7f\""},
		{"unicode raw", "ü", `"ü"`},
		{"string kind", openfeature.StatusReady, `"READY"`},
		{"ordered object sorted", object{{"z", 1.0}, {"a", 2.0}}, `{"a":2,"z":1}`},
		{"openfeature object sorted", openfeature.Object{{Key: "z", Value: openfeature.Object{{Key: "y", Value: 1}, {Key: "b", Value: nil}}}, {Key: "a", Value: []any{openfeature.Object{}}}}, `{"a":[{}],"z":{"b":null,"y":1}}`},
		{"context", openfeature.Context{TargetingKey: "k", Attributes: map[string]any{"d": time.Date(1999, 12, 31, 23, 59, 59, 999e6, time.UTC), "n": nil}}, `{"d":"1999-12-31T23:59:59.999Z","n":null,"targetingKey":"k"}`},
		{"empty context", openfeature.Context{}, `{}`},
		{"error details keep empty message", details(openfeature.EvaluationDetails[bool]{FlagKey: "f", ResolutionDetails: openfeature.ResolutionDetails[bool]{Value: false, Reason: openfeature.ReasonError, ErrorCode: openfeature.CodeParseError, FlagMetadata: openfeature.FlagMetadata{}}}), `{"errorCode":"PARSE_ERROR","errorMessage":"","flagKey":"f","flagMetadata":{},"reason":"ERROR","value":false}`},
		{"success details", details(openfeature.EvaluationDetails[any]{FlagKey: "f", ResolutionDetails: openfeature.ResolutionDetails[any]{Value: nil, FlagMetadata: openfeature.FlagMetadata{"a": 1.0}}}), `{"flagKey":"f","flagMetadata":{"a":1},"value":null}`},
		{"event details", openfeature.EventDetails{ProviderName: "p", FlagsChanged: []string{"b", "a"}}, `{"flagsChanged":["b","a"],"providerName":"p"}`},
		{"provider metadata", openfeature.Metadata{Name: "p", Attributes: map[string]string{"v": "1"}}, `{"name":"p","v":"1"}`},
		{"client metadata", openfeature.ClientMetadata{Paradigm: "client", SDK: "webgo", ProviderMetadata: openfeature.Metadata{Name: "n"}}, `{"paradigm":"client","providerMetadata":{"name":"n"}}`},
		{"tracking", openfeature.TrackingDetails{Value: new(0.0), Attributes: map[string]any{"k": "v"}}, `{"k":"v","value":0}`},
		{"tracking empty", openfeature.TrackingDetails{}, `{}`},
		{"telemetry", openfeature.EvaluationEvent{Name: "n", Attributes: map[string]any{"x": 0.0}}, `{"attributes":{"x":0},"name":"n"}`},
	}
	for _, c := range cases {
		if got := string(canonical(c.in)); got != c.want {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}

func TestOrderedDecode(t *testing.T) {
	v, err := decodeJSON([]byte(`{"zeta":1,"alpha":[true,null],"2":0,"10":0,"01":0,"mid":{"b":2,"a":1e-7},"zeta":-0,"<&>":" "}`))
	if err != nil {
		t.Fatal(err)
	}
	b, err := v.(object).MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	want := "{\"2\":0,\"10\":0,\"zeta\":0,\"alpha\":[true,null],\"01\":0,\"mid\":{\"b\":2,\"a\":1e-7},\"<&>\":\" \"}"
	if string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}
	if _, err := decodeJSON([]byte(`{} x`)); err == nil {
		t.Error("trailing data accepted")
	}
	var x struct{ A, B val }
	if err := json.Unmarshal([]byte(`{"A":null}`), &x); err != nil || !x.A.set || x.A.v != nil || x.B.set {
		t.Errorf("val presence: %+v %v", x, err)
	}
}

type skip struct{}

type object []member

type member struct {
	Key   string
	Value any
}

type val struct {
	v   any
	set bool
}

func (x *val) UnmarshalJSON(b []byte) error {
	v, err := decodeJSON(b)
	x.v, x.set = v, true
	return err
}

func (o object) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	stringify(&b, o)
	return b.Bytes(), nil
}

func decodeJSON(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	v, err := decodeValue(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("trailing data after JSON value")
	}
	return v, nil
}

func decodeValue(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch t := tok.(type) {
	case json.Number:
		return strconv.ParseFloat(string(t), 64)
	case json.Delim:
		if t == '[' {
			arr := []any{}
			for dec.More() {
				v, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				arr = append(arr, v)
			}
			_, err := dec.Token()
			return arr, err
		}
		var obj object
		for dec.More() {
			k, err := dec.Token()
			if err != nil {
				return nil, err
			}
			v, err := decodeValue(dec)
			if err != nil {
				return nil, err
			}
			obj = obj.put(k.(string), v)
		}
		if _, err := dec.Token(); err != nil {
			return nil, err
		}
		return jsOrder(obj), nil
	}
	return tok, nil
}

func (o object) put(k string, v any) object {
	for i := range o {
		if o[i].Key == k {
			o[i].Value = v
			return o
		}
	}
	return append(o, member{k, v})
}

func arrayIndex(k string) (uint64, bool) {
	n, err := strconv.ParseUint(k, 10, 32)
	return n, err == nil && n < math.MaxUint32 && strconv.FormatUint(n, 10) == k
}

func jsOrder(o object) object {
	out := object{}
	for _, m := range o {
		if _, ok := arrayIndex(m.Key); ok {
			out = append(out, m)
		}
	}
	slices.SortStableFunc(out, func(a, b member) int {
		x, _ := arrayIndex(a.Key)
		y, _ := arrayIndex(b.Key)
		return int(x) - int(y)
	})
	for _, m := range o {
		if _, ok := arrayIndex(m.Key); !ok {
			out = append(out, m)
		}
	}
	return out
}

func plain(v any) any {
	switch x := v.(type) {
	case object:
		m := make(map[string]any, len(x))
		for _, e := range x {
			m[e.Key] = plain(e.Value)
		}
		return m
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = plain(e)
		}
		return out
	}
	return v
}

func stringify(b *bytes.Buffer, v any) {
	switch x := v.(type) {
	case object:
		b.WriteByte('{')
		for i, m := range x {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(jsQuote(m.Key))
			b.WriteByte(':')
			stringify(b, m.Value)
		}
		b.WriteByte('}')
	case []any:
		b.WriteByte('[')
		for i, e := range x {
			if i > 0 {
				b.WriteByte(',')
			}
			stringify(b, e)
		}
		b.WriteByte(']')
	default:
		writeCanon(b, v)
	}
}

func canonical(v any) []byte {
	var b bytes.Buffer
	writeCanon(&b, v)
	return b.Bytes()
}

func details[T any](d openfeature.EvaluationDetails[T]) map[string]any {
	m := map[string]any{"flagKey": d.FlagKey, "value": any(d.Value), "flagMetadata": map[string]any(d.FlagMetadata)}
	put(m, "variant", d.Variant)
	put(m, "reason", string(d.Reason))
	put(m, "errorCode", string(d.ErrorCode))
	if d.ErrorCode != "" || d.ErrorMessage != "" {
		m["errorMessage"] = d.ErrorMessage
	}
	return m
}

func putIf(m map[string]any, k, v string, present bool) {
	if present {
		m[k] = v
		return
	}
	put(m, k, v)
}

func put(m map[string]any, k, v string) {
	if v != "" {
		m[k] = v
	}
}

func contextMap(c openfeature.Context) map[string]any {
	m := make(map[string]any, len(c.Attributes)+1)
	for k, v := range c.Attributes {
		m[k] = v
	}
	put(m, "targetingKey", c.TargetingKey)
	return m
}

func metadataMap(md openfeature.Metadata) map[string]any {
	m := map[string]any{}
	for k, v := range md.Attributes {
		m[k] = v
	}
	m["name"] = md.Name
	return m
}

func structured(v any) (any, bool) {
	switch x := v.(type) {
	case openfeature.Context:
		return contextMap(x), true
	case openfeature.EvaluationDetails[any]:
		return details(x), true
	case openfeature.Metadata:
		return metadataMap(x), true
	case openfeature.ClientMetadata:
		m := map[string]any{"providerMetadata": metadataMap(x.ProviderMetadata)}
		putIf(m, "domain", x.Domain, x.HasDomain)
		putIf(m, "name", x.Name, x.HasDomain)
		put(m, "version", x.Version)
		put(m, "paradigm", x.Paradigm)
		return m, true
	case openfeature.EventDetails:
		m := map[string]any{"flagsChanged": x.FlagsChanged, "metadata": x.Metadata}
		put(m, "providerName", x.ProviderName)
		putIf(m, "domain", x.Domain, x.HasDomain)
		putIf(m, "clientName", x.ClientName, x.HasClientName)
		put(m, "message", x.Message)
		put(m, "errorCode", string(x.ErrorCode))
		return m, true
	case openfeature.TrackingDetails:
		m := map[string]any{}
		for k, v := range x.Attributes {
			m[k] = v
		}
		if x.Value != nil {
			m["value"] = *x.Value
		}
		return m, true
	case openfeature.EvaluationEvent:
		return map[string]any{"name": x.Name, "attributes": x.Attributes}, true
	case object:
		return plain(x), true
	case openfeature.Object:
		if x == nil {
			return nil, true
		}
		m := make(map[string]any, len(x))
		for _, f := range x {
			m[f.Key] = f.Value
		}
		return m, true
	}
	return nil, false
}

func absent(v any) bool {
	if _, ok := v.(skip); ok {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.IsValid() && (rv.Kind() == reflect.Map || rv.Kind() == reflect.Slice) && rv.IsNil()
}

func less16(a, b string) int {
	return slices.Compare(utf16.Encode([]rune(a)), utf16.Encode([]rune(b)))
}

func writeCanon(b *bytes.Buffer, v any) {
	if s, ok := structured(v); ok {
		writeCanon(b, s)
		return
	}
	switch x := v.(type) {
	case nil, skip:
		b.WriteString("null")
		return
	case time.Time:
		b.WriteString(jsQuote(x.UTC().Format("2006-01-02T15:04:05.000Z")))
		return
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool:
		b.WriteString(strconv.FormatBool(rv.Bool()))
	case reflect.String:
		b.WriteString(jsQuote(rv.String()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.WriteString(jsNumber(float64(rv.Int())))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		b.WriteString(jsNumber(float64(rv.Uint())))
	case reflect.Float32, reflect.Float64:
		b.WriteString(jsNumber(rv.Float()))
	case reflect.Pointer, reflect.Interface:
		if rv.IsNil() {
			b.WriteString("null")
		} else {
			writeCanon(b, rv.Elem().Interface())
		}
	case reflect.Slice, reflect.Array:
		if rv.Kind() == reflect.Slice && rv.IsNil() {
			b.WriteString("null")
			return
		}
		b.WriteByte('[')
		for i := range rv.Len() {
			if i > 0 {
				b.WriteByte(',')
			}
			writeCanon(b, rv.Index(i).Interface())
		}
		b.WriteByte(']')
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			panic(fmt.Sprintf("canonical: map key %s", rv.Type().Key()))
		}
		if rv.IsNil() {
			b.WriteString("null")
			return
		}
		var keys []string
		for _, k := range rv.MapKeys() {
			if !absent(rv.MapIndex(k).Interface()) {
				keys = append(keys, k.String())
			}
		}
		slices.SortFunc(keys, less16)
		b.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(jsQuote(k))
			b.WriteByte(':')
			writeCanon(b, rv.MapIndex(reflect.ValueOf(k).Convert(rv.Type().Key())).Interface())
		}
		b.WriteByte('}')
	default:
		panic(fmt.Sprintf("canonical: unsupported %T", v))
	}
}

func jsQuote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

func jsNumber(f float64) string {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return "null"
	}
	if f == 0 {
		return "0"
	}
	sign := ""
	if f < 0 {
		sign, f = "-", -f
	}
	e := strconv.FormatFloat(f, 'e', -1, 64)
	mant, exp, _ := strings.Cut(e, "e")
	digits := strings.Replace(mant, ".", "", 1)
	x, _ := strconv.Atoi(exp)
	k, n := len(digits), x+1
	switch {
	case k <= n && n <= 21:
		return sign + digits + strings.Repeat("0", n-k)
	case 0 < n && n <= 21:
		return sign + digits[:n] + "." + digits[n:]
	case -6 < n && n <= 0:
		return sign + "0." + strings.Repeat("0", -n) + digits
	}
	es := "+"
	if n-1 < 0 {
		es = "-"
	}
	m := digits[:1]
	if k > 1 {
		m += "." + digits[1:]
	}
	return sign + m + "e" + es + strconv.Itoa(abs(n-1))
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
