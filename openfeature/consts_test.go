package openfeature

import "testing"

func TestConstantStrings(t *testing.T) {
	cases := []struct{ got, want string }{
		{string(EventReady), "PROVIDER_READY"},
		{string(EventError), "PROVIDER_ERROR"},
		{string(EventConfigurationChanged), "PROVIDER_CONFIGURATION_CHANGED"},
		{string(EventContextChanged), "PROVIDER_CONTEXT_CHANGED"},
		{string(EventReconciling), "PROVIDER_RECONCILING"},
		{string(EventStale), "PROVIDER_STALE"},
		{string(StatusNotReady), "NOT_READY"},
		{string(StatusReady), "READY"},
		{string(StatusError), "ERROR"},
		{string(StatusStale), "STALE"},
		{string(StatusFatal), "FATAL"},
		{string(StatusReconciling), "RECONCILING"},
		{string(CodeProviderNotReady), "PROVIDER_NOT_READY"},
		{string(CodeProviderFatal), "PROVIDER_FATAL"},
		{string(CodeFlagNotFound), "FLAG_NOT_FOUND"},
		{string(CodeParseError), "PARSE_ERROR"},
		{string(CodeTypeMismatch), "TYPE_MISMATCH"},
		{string(CodeTargetingKeyMissing), "TARGETING_KEY_MISSING"},
		{string(CodeInvalidContext), "INVALID_CONTEXT"},
		{string(CodeGeneral), "GENERAL"},
		{string(ReasonStatic), "STATIC"},
		{string(ReasonDefault), "DEFAULT"},
		{string(ReasonTargetingMatch), "TARGETING_MATCH"},
		{string(ReasonSplit), "SPLIT"},
		{string(ReasonCached), "CACHED"},
		{string(ReasonDisabled), "DISABLED"},
		{string(ReasonUnknown), "UNKNOWN"},
		{string(ReasonStale), "STALE"},
		{string(ReasonError), "ERROR"},
		{string(TypeBoolean), "boolean"},
		{string(TypeString), "string"},
		{string(TypeNumber), "number"},
		{string(TypeObject), "object"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("got %q want %q", c.got, c.want)
		}
	}
}

func TestContextCopyIsDeep(t *testing.T) {
	in := Context{TargetingKey: "k", Attributes: map[string]any{
		"m": map[string]any{"a": []any{1.0, map[string]any{"b": "c"}}},
	}}
	out := in.copy()
	in.Attributes["m"].(map[string]any)["a"].([]any)[1].(map[string]any)["b"] = "changed"
	got := out.Attributes["m"].(map[string]any)["a"].([]any)[1].(map[string]any)["b"]
	if got != "c" || out.TargetingKey != "k" {
		t.Fatalf("copy not deep: %v", got)
	}
	if (Context{}).copy().Attributes != nil {
		t.Fatal("nil attributes must stay nil")
	}
}
