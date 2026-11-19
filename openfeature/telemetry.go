package openfeature

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
)

// Telemetry attribute keys of an evaluation event.
const (
	TelemetryAttributeKey          = "feature_flag.key"
	TelemetryAttributeErrorCode    = "error.type"
	TelemetryAttributeErrorMessage = "error.message"
	TelemetryAttributeVariant      = "feature_flag.result.variant"
	TelemetryAttributeValue        = "feature_flag.result.value"
	TelemetryAttributeContextID    = "feature_flag.context.id"
	TelemetryAttributeReason       = "feature_flag.result.reason"
	TelemetryAttributeProvider     = "feature_flag.provider.name"
	TelemetryAttributeFlagSetID    = "feature_flag.set.id"
	TelemetryAttributeVersion      = "feature_flag.version"
)

// Flag metadata keys read by CreateEvaluationEvent.
const (
	TelemetryFlagMetadataContextID = "contextId"
	TelemetryFlagMetadataFlagSetID = "flagSetId"
	TelemetryFlagMetadataVersion   = "version"
)

// EvaluationEvent is an OpenTelemetry-shaped evaluation event.
type EvaluationEvent struct {
	Name       string
	Attributes map[string]any
}

// CreateEvaluationEvent builds the telemetry event for one evaluation.
// Object values are stringified like JavaScript's JSON.stringify: array-index
// keys come first in numeric order, other map keys are sorted, Object keys
// keep slice order (a repeated key once, with its last value), struct fields
// keep declaration order, json.Marshaler values keep their own order, and
// float32 values are widened to float64.
func CreateEvaluationEvent(hc HookContext, d EvaluationDetails[any]) EvaluationEvent {
	reason := d.Reason
	if reason == "" {
		reason = ReasonUnknown
	}
	a := map[string]any{
		TelemetryAttributeKey:      hc.FlagKey,
		TelemetryAttributeProvider: hc.ProviderMetadata.Name,
		TelemetryAttributeReason:   strings.ToLower(string(reason)),
	}
	if d.Variant != "" {
		a[TelemetryAttributeVariant] = d.Variant
	}
	if !isNull(d.Value) {
		if primitive(d.Value) {
			a[TelemetryAttributeValue] = d.Value
		} else if s, err := stringify(d.Value); err == nil {
			a[TelemetryAttributeValue] = s
		}
	}
	contextID, ok := d.FlagMetadata[TelemetryFlagMetadataContextID]
	if !ok || contextID == nil {
		contextID = hc.Context.TargetingKey
	}
	if truthy(contextID) {
		a[TelemetryAttributeContextID] = contextID
	}
	if v := d.FlagMetadata[TelemetryFlagMetadataFlagSetID]; truthy(v) {
		a[TelemetryAttributeFlagSetID] = v
	}
	if v := d.FlagMetadata[TelemetryFlagMetadataVersion]; truthy(v) {
		a[TelemetryAttributeVersion] = v
	}
	if d.Reason == ReasonError {
		code := d.ErrorCode
		if code == "" {
			code = CodeGeneral
		}
		a[TelemetryAttributeErrorCode] = strings.ToLower(string(code))
		if d.ErrorMessage != "" {
			a[TelemetryAttributeErrorMessage] = d.ErrorMessage
		}
	}
	return EvaluationEvent{Name: "feature_flag.evaluation", Attributes: a}
}

func isNull(v any) bool {
	if v == nil {
		return true
	}
	switch rv := reflect.ValueOf(v); rv.Kind() {
	case reflect.Map, reflect.Slice, reflect.Pointer, reflect.Interface:
		return rv.IsNil()
	}
	return false
}

func primitive(v any) bool {
	if _, ok := v.(json.Marshaler); ok {
		return false
	}
	switch reflect.ValueOf(v).Kind() {
	case reflect.Bool, reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func truthy(v any) bool {
	if isNull(v) {
		return false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.String:
		return rv.String() != ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() != 0
	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		return f != 0 && !math.IsNaN(f)
	}
	return true
}
