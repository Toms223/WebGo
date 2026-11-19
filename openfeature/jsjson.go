package openfeature

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

var errCycle = errors.New("cyclic value")

type visit struct {
	p uintptr
	n int
	t reflect.Type
}

func stringify(v any) (string, error) {
	var b strings.Builder
	err := writeJS(&b, v, map[visit]bool{})
	return b.String(), err
}

func writeJS(b *strings.Builder, v any, seen map[visit]bool) error {
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
		return nil
	case time.Time:
		b.WriteString(jsQuote(x.UTC().Format("2006-01-02T15:04:05.000Z")))
		return nil
	case *time.Time:
		if x == nil {
			b.WriteString("null")
			return nil
		}
		return writeJS(b, *x, seen)
	case Object:
		if x == nil {
			b.WriteString("null")
			return nil
		}
		k := visit{p: reflect.ValueOf(x).Pointer(), n: len(x), t: reflect.TypeFor[Object]()}
		if seen[k] {
			return errCycle
		}
		seen[k] = true
		defer delete(seen, k)
		ms := make([]member, 0, len(x))
		at := map[string]int{}
		for _, f := range x {
			if i, ok := at[f.Key]; ok {
				ms[i].val = f.Value
				continue
			}
			at[f.Key] = len(ms)
			ms = append(ms, member{f.Key, f.Value})
		}
		return writeObject(b, ms, seen)
	case json.Marshaler:
		if isNull(v) {
			b.WriteString("null")
			return nil
		}
		raw, err := x.MarshalJSON()
		if err != nil {
			return err
		}
		return reemit(b, raw)
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
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice:
		if rv.IsNil() {
			b.WriteString("null")
			return nil
		}
		if rv.Kind() == reflect.Interface {
			return writeJS(b, rv.Elem().Interface(), seen)
		}
		k := visit{p: rv.Pointer(), t: rv.Type()}
		if rv.Kind() == reflect.Slice {
			k.n = rv.Len()
		}
		if seen[k] {
			return errCycle
		}
		seen[k] = true
		defer delete(seen, k)
		switch rv.Kind() {
		case reflect.Pointer:
			return writeJS(b, rv.Elem().Interface(), seen)
		case reflect.Map:
			return writeMap(b, rv, seen)
		}
		return writeList(b, rv, seen)
	case reflect.Array:
		return writeList(b, rv, seen)
	case reflect.Struct:
		return writeStruct(b, rv, seen)
	default:
		return marshalJS(b, v)
	}
	return nil
}

func writeList(b *strings.Builder, rv reflect.Value, seen map[visit]bool) error {
	b.WriteByte('[')
	for i := range rv.Len() {
		if i > 0 {
			b.WriteByte(',')
		}
		if err := writeJS(b, rv.Index(i).Interface(), seen); err != nil {
			return err
		}
	}
	b.WriteByte(']')
	return nil
}

type member struct {
	key string
	val any
}

func writeMap(b *strings.Builder, rv reflect.Value, seen map[visit]bool) error {
	if rv.Type().Key().Kind() != reflect.String {
		return marshalJS(b, rv.Interface())
	}
	ms := make([]member, 0, rv.Len())
	for it := rv.MapRange(); it.Next(); {
		ms = append(ms, member{it.Key().String(), it.Value().Interface()})
	}
	slices.SortFunc(ms, func(x, y member) int { return strings.Compare(x.key, y.key) })
	return writeObject(b, ms, seen)
}

func writeStruct(b *strings.Builder, rv reflect.Value, seen map[visit]bool) error {
	ms, err := structMembers(rv, nil)
	if err != nil {
		return err
	}
	return writeObject(b, ms, seen)
}

func structMembers(rv reflect.Value, ms []member) ([]member, error) {
	t := rv.Type()
	for i := range t.NumField() {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name, opts, _ := strings.Cut(tag, ",")
		fv := rv.Field(i)
		if f.Anonymous && name == "" {
			ft := f.Type
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				if fv.Kind() == reflect.Pointer {
					if fv.IsNil() {
						continue
					}
					fv = fv.Elem()
				}
				var err error
				if ms, err = structMembers(fv, ms); err != nil {
					return nil, err
				}
				continue
			}
		}
		if !f.IsExported() || !fv.CanInterface() {
			continue
		}
		if name == "" {
			name = f.Name
		}
		if slices.Contains(strings.Split(opts, ","), "omitempty") && emptyValue(fv) {
			continue
		}
		ms = append(ms, member{name, fv.Interface()})
	}
	return ms, nil
}

func emptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Interface, reflect.Pointer:
		return v.IsZero()
	}
	return false
}

func arrayIndex(k string) (uint64, bool) {
	if k == "" || len(k) > 10 || (len(k) > 1 && k[0] == '0') {
		return 0, false
	}
	n, err := strconv.ParseUint(k, 10, 64)
	return n, err == nil && n <= 1<<32-2
}

func writeObject(b *strings.Builder, ms []member, seen map[visit]bool) error {
	slices.SortStableFunc(ms, func(x, y member) int {
		xi, xok := arrayIndex(x.key)
		yi, yok := arrayIndex(y.key)
		switch {
		case xok && yok:
			return cmp.Compare(xi, yi)
		case xok:
			return -1
		case yok:
			return 1
		}
		return 0
	})
	b.WriteByte('{')
	for i, m := range ms {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(jsQuote(m.key))
		b.WriteByte(':')
		if err := writeJS(b, m.val, seen); err != nil {
			return err
		}
	}
	b.WriteByte('}')
	return nil
}

func marshalJS(b *strings.Builder, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return reemit(b, raw)
}

func reemit(b *strings.Builder, raw []byte) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := reemitValue(b, dec); err != nil {
		return err
	}
	if dec.More() {
		return errors.New("trailing data after JSON value")
	}
	return nil
}

func reemitValue(b *strings.Builder, dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	switch t := tok.(type) {
	case json.Delim:
		closer := byte(']')
		if t == '{' {
			closer = '}'
		}
		b.WriteByte(byte(t))
		for i := 0; dec.More(); i++ {
			if i > 0 {
				b.WriteByte(',')
			}
			if t == '{' {
				k, err := dec.Token()
				if err != nil {
					return err
				}
				b.WriteString(jsQuote(k.(string)))
				b.WriteByte(':')
			}
			if err := reemitValue(b, dec); err != nil {
				return err
			}
		}
		if _, err := dec.Token(); err != nil {
			return err
		}
		b.WriteByte(closer)
	case json.Number:
		f, _ := strconv.ParseFloat(string(t), 64)
		b.WriteString(jsNumber(f))
	case string:
		b.WriteString(jsQuote(t))
	case bool:
		b.WriteString(strconv.FormatBool(t))
	case nil:
		b.WriteString("null")
	}
	return nil
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
	mant, exp, _ := strings.Cut(strconv.FormatFloat(f, 'e', -1, 64), "e")
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
	e := n - 1
	if e < 0 {
		e = -e
	}
	return sign + m + "e" + es + strconv.Itoa(e)
}
