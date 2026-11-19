package openfeature

// Context is the evaluation context. Attributes may hold nil, bool, string,
// numbers, time.Time, []any, map[string]any and Object.
type Context struct {
	TargetingKey string
	Attributes   map[string]any
}

func (c Context) copy() Context {
	out := Context{TargetingKey: c.TargetingKey}
	if c.Attributes != nil {
		out.Attributes = copyMap(c.Attributes)
	}
	return out
}

func copyMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = copyValue(v)
	}
	return out
}

func copyValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		return copyMap(x)
	case Object:
		if x == nil {
			return x
		}
		out := make(Object, len(x))
		for i, f := range x {
			out[i] = Field{f.Key, copyValue(f.Value)}
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = copyValue(e)
		}
		return out
	}
	return v
}

// FlagMetadata holds string, float64 or bool values.
type FlagMetadata map[string]any

// ResolutionDetails is what a provider returns for one flag.
type ResolutionDetails[T any] struct {
	Value        T
	Variant      string
	Reason       Reason
	ErrorCode    ErrorCode
	ErrorMessage string
	FlagMetadata FlagMetadata
}

// EvaluationDetails is what a client returns for one flag.
type EvaluationDetails[T any] struct {
	FlagKey string
	ResolutionDetails[T]
}

// Metadata describes a provider.
type Metadata struct {
	Name       string
	Attributes map[string]string
}

// ClientMetadata describes a client.
type ClientMetadata struct {
	Domain           string
	Name             string
	Version          string
	SDK              string
	Paradigm         string
	ProviderMetadata Metadata
	HasDomain        bool
}

// TrackingDetails carries the optional value and attributes of a Track call.
type TrackingDetails struct {
	Value      *float64
	Attributes map[string]any
}
