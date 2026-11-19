package openfeature

import (
	"math"
	"testing"
)

func TestObject(t *testing.T) {
	inner := Object{{"y", 1.5}, {"b", nil}}
	o := Object{{"z", "<&> "}, {"a", inner}, {"2", math.Copysign(0, -1)}, {"m", map[string]any{"k": 1e21}}}
	b, err := o.MarshalJSON()
	if want := "{\"2\":0,\"z\":\"<&> \",\"a\":{\"y\":1.5,\"b\":null},\"m\":{\"k\":1e+21}}"; err != nil || string(b) != want {
		t.Errorf("MarshalJSON = %s, %v; want %s", b, err, want)
	}
	if v, ok := o.Get("a"); !ok || len(v.(Object)) != 2 {
		t.Errorf("Get(a) = %v, %v", v, ok)
	}
	if _, ok := o.Get("q"); ok {
		t.Error("Get(q) found")
	}
	cyc := Object{{"self", nil}}
	cyc[0].Value = cyc
	if _, err := cyc.MarshalJSON(); err == nil {
		t.Error("cycle not detected")
	}
	e := CreateEvaluationEvent(HookContext{FlagKey: "f"}, EvaluationDetails[any]{FlagKey: "f", ResolutionDetails: ResolutionDetails[any]{Value: o}})
	if got := e.Attributes[TelemetryAttributeValue]; got != string(b) {
		t.Errorf("telemetry value = %v", got)
	}
	c := Context{Attributes: map[string]any{"o": o}}.copy()
	c.Attributes["o"].(Object)[1].Value.(Object)[0].Value = 9
	if inner[0].Value != 1.5 {
		t.Error("context copy shares Object")
	}
}

func TestObjectDuplicateKeys(t *testing.T) {
	o := Object{{"a", 1}, {"b", 2}, {"a", 3}, {"1", 4}, {"1", 5}}
	b, err := o.MarshalJSON()
	if want := `{"1":5,"a":3,"b":2}`; err != nil || string(b) != want {
		t.Errorf("MarshalJSON = %s, %v; want %s", b, err, want)
	}
	if v, ok := o.Get("a"); !ok || v != 3 {
		t.Errorf("Get(a) = %v, %v; want 3", v, ok)
	}
}
