package openfeature

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
	"time"
)

type rawJSON string

func (r rawJSON) MarshalJSON() ([]byte, error) { return []byte(r), nil }

type pair struct {
	B float64 `json:"b"`
	A string  `json:"a"`
}

type inner struct {
	In int `json:"in"`
}

type outer struct {
	inner
	Z time.Time
	N float64 `json:"n"`
	T string  `json:"tag,omitempty"`
	E string  `json:"e,omitempty"`
	P *pair   `json:"p"`
	X int     `json:"-"`
	y int
}

type node struct{ Next *node }

func TestCreateEvaluationEvent(t *testing.T) {
	hc := HookContext{FlagKey: "f", ProviderMetadata: Metadata{Name: "p"}, Context: Context{TargetingKey: "tk"}}
	base := map[string]any{"feature_flag.key": "f", "feature_flag.provider.name": "p", "feature_flag.context.id": "tk"}
	with := func(kv ...any) map[string]any {
		m := map[string]any{}
		for k, v := range base {
			m[k] = v
		}
		for i := 0; i < len(kv); i += 2 {
			if kv[i+1] == nil {
				delete(m, kv[i].(string))
			} else {
				m[kv[i].(string)] = kv[i+1]
			}
		}
		return m
	}
	det := func(v any, r Reason, md FlagMetadata) EvaluationDetails[any] {
		return EvaluationDetails[any]{FlagKey: "f", ResolutionDetails: ResolutionDetails[any]{Value: v, Reason: r, FlagMetadata: md}}
	}
	shared := []any{1.0, nil}
	shared[1] = shared[:1]
	loop := &node{}
	loop.Next = loop
	cyclic := map[string]any{}
	cyclic["self"] = cyclic
	tm := time.Date(2024, 1, 2, 3, 4, 5, 6e6, time.FixedZone("x", 3600))
	cases := []struct {
		name string
		d    EvaluationDetails[any]
		want map[string]any
	}{
		{"no reason is unknown", det("on", "", nil), with("feature_flag.result.reason", "unknown", "feature_flag.result.value", "on")},
		{"reason lowered, variant", func() EvaluationDetails[any] {
			d := det(false, ReasonTargetingMatch, FlagMetadata{})
			d.Variant = "v1"
			return d
		}(), with("feature_flag.result.reason", "targeting_match", "feature_flag.result.value", false, "feature_flag.result.variant", "v1")},
		{"zero number kept", det(0.0, ReasonStatic, nil), with("feature_flag.result.reason", "static", "feature_flag.result.value", 0.0)},
		{"int kept raw", det(int64(4), ReasonStatic, nil), with("feature_flag.result.reason", "static", "feature_flag.result.value", int64(4))},
		{"nil value omitted", det(nil, ReasonStatic, nil), with("feature_flag.result.reason", "static")},
		{"typed nil map omitted", det(map[string]any(nil), ReasonStatic, nil), with("feature_flag.result.reason", "static")},
		{"metadata ids", det("x", ReasonStatic, FlagMetadata{"contextId": "c", "flagSetId": "s", "version": 2.0}),
			with("feature_flag.result.reason", "static", "feature_flag.result.value", "x", "feature_flag.context.id", "c", "feature_flag.set.id", "s", "feature_flag.version", 2.0)},
		{"falsy metadata omitted, empty contextId does not fall back", det("x", ReasonStatic, FlagMetadata{"contextId": "", "flagSetId": 0.0, "version": false}),
			with("feature_flag.result.reason", "static", "feature_flag.result.value", "x", "feature_flag.context.id", nil)},
		{"nil contextId falls back", det("x", ReasonStatic, FlagMetadata{"contextId": nil}),
			with("feature_flag.result.reason", "static", "feature_flag.result.value", "x")},
		{"error with message", func() EvaluationDetails[any] {
			d := det(true, ReasonError, FlagMetadata{})
			d.ErrorCode, d.ErrorMessage = CodeFlagNotFound, "missing"
			return d
		}(), with("feature_flag.result.reason", "error", "feature_flag.result.value", true, "error.type", "flag_not_found", "error.message", "missing")},
		{"error without code or message", det(true, ReasonError, nil),
			with("feature_flag.result.reason", "error", "feature_flag.result.value", true, "error.type", "general")},
		{"error code ignored without ERROR reason", func() EvaluationDetails[any] {
			d := det(true, ReasonStatic, nil)
			d.ErrorCode, d.ErrorMessage = CodeFlagNotFound, "missing"
			return d
		}(), with("feature_flag.result.reason", "static", "feature_flag.result.value", true)},
		{"map stringified sorted", det(map[string]any{"z": 1.0, "a": []any{true, nil, 1e21, 1e-7, math.NaN()}, "m": "<\u2028>\n"}, ReasonStatic, nil),
			with("feature_flag.result.reason", "static", "feature_flag.result.value", `{"a":[true,null,1e+21,1e-7,null],"m":"<`+"\u2028"+`>\n","z":1}`)},
		{"marshaler order honoured and normalised", det(rawJSON(`{"z":1.0,"a":"\u003c","n":1E-07}`), ReasonStatic, nil),
			with("feature_flag.result.reason", "static", "feature_flag.result.value", `{"z":1,"a":"<","n":1e-7}`)},
		{"struct via json", det(pair{B: 0.5, A: "&"}, ReasonStatic, nil),
			with("feature_flag.result.reason", "static", "feature_flag.result.value", `{"b":0.5,"a":"&"}`)},
		{"time as ISO millis", det([]any{tm}, ReasonStatic, nil),
			with("feature_flag.result.reason", "static", "feature_flag.result.value", `["2024-01-02T02:04:05.006Z"]`)},
		{"array-index keys first", det(map[string]any{"b": nil, "10": 1.0, "-1": 3.0, "2": 2.0, "01": 4.0, "4294967295": 5.0, "4294967294": 6.0}, ReasonStatic, nil),
			with("feature_flag.result.reason", "static", "feature_flag.result.value", `{"2":2,"10":1,"4294967294":6,"-1":3,"01":4,"4294967295":5,"b":null}`)},
		{"struct walked with JS rules", det(outer{inner: inner{In: 1}, Z: tm, N: math.NaN(), T: "t", P: &pair{B: 1, A: "a"}, X: 2}, ReasonStatic, nil),
			with("feature_flag.result.reason", "static", "feature_flag.result.value", `{"in":1,"Z":"2024-01-02T02:04:05.006Z","n":null,"tag":"t","p":{"b":1,"a":"a"}}`)},
		{"shorter view of same array is not a cycle", det(shared, ReasonStatic, nil),
			with("feature_flag.result.reason", "static", "feature_flag.result.value", `[1,[1]]`)},
		{"cyclic pointer omitted", det(loop, ReasonStatic, nil), with("feature_flag.result.reason", "static")},
		{"cyclic value omitted", det(cyclic, ReasonStatic, nil), with("feature_flag.result.reason", "static")},
	}
	for _, c := range cases {
		e := CreateEvaluationEvent(hc, c.d)
		if e.Name != "feature_flag.evaluation" || !reflect.DeepEqual(e.Attributes, c.want) {
			got, _ := json.Marshal(e.Attributes)
			want, _ := json.Marshal(c.want)
			t.Errorf("%s:\n got %s\nwant %s", c.name, got, want)
		}
	}
	e := CreateEvaluationEvent(HookContext{FlagKey: "f"}, det("x", ReasonStatic, nil))
	if _, ok := e.Attributes["feature_flag.context.id"]; ok {
		t.Error("empty targeting key set context id")
	}
}
