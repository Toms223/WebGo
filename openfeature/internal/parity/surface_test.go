package parity

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"slices"
	"testing"

	of "github.com/Toms223/WebGo/openfeature"
	"github.com/Toms223/WebGo/openfeature/memprovider"
)

var pkg = reflect.TypeFor[of.Client]().PkgPath()

var goSymbols = map[string]any{
	"openfeature":           pkg,
	"package openfeature":   pkg,
	"SetProvider":           of.SetProvider,
	"SetProviderAndWait":    of.SetProviderAndWait,
	"GetProvider":           of.GetProvider,
	"ProviderMetadata":      of.ProviderMetadata,
	"ProviderStatus":        of.ProviderStatus,
	"SetContext":            of.SetContext,
	"GetContext":            of.GetContext,
	"ClearContext":          of.ClearContext,
	"ClearContexts":         of.ClearContexts,
	"AddHooks":              of.AddHooks,
	"Hooks":                 of.Hooks,
	"ClearHooks":            of.ClearHooks,
	"AddHandler":            of.AddHandler,
	"Handlers":              of.Handlers,
	"ClearHandlers":         of.ClearHandlers,
	"GetClient":             of.GetClient,
	"GetClientVersion":      of.GetClientVersion,
	"SetLogger":             of.SetLogger,
	"ClearProviders":        of.ClearProviders,
	"Close":                 of.Close,
	"DefaultLogger":         of.DefaultLogger,
	"NoopProvider":          of.NoopProvider,
	"CreateEvaluationEvent": of.CreateEvaluationEvent,
	"WithHooks":             of.WithHooks,
	"WithHookHints":         of.WithHookHints,

	"Client.Boolean":        (*of.Client).Boolean,
	"Client.BooleanDetails": (*of.Client).BooleanDetails,
	"Client.String":         (*of.Client).String,
	"Client.StringDetails":  (*of.Client).StringDetails,
	"Client.Number":         (*of.Client).Number,
	"Client.NumberDetails":  (*of.Client).NumberDetails,
	"Client.Object":         (*of.Client).Object,
	"Client.ObjectDetails":  (*of.Client).ObjectDetails,
	"Client.Track":          (*of.Client).Track,
	"Client.AddHooks":       (*of.Client).AddHooks,
	"Client.Hooks":          (*of.Client).Hooks,
	"Client.ClearHooks":     (*of.Client).ClearHooks,
	"Client.SetLogger":      (*of.Client).SetLogger,
	"Client.AddHandler":     (*of.Client).AddHandler,
	"Client.Handlers":       (*of.Client).Handlers,
	"Client.Metadata":       (*of.Client).Metadata,
	"Client.ProviderStatus": (*of.Client).ProviderStatus,

	"Emitter":                   reflect.TypeFor[of.Emitter](),
	"Emitter.Emit":              (*of.Emitter).Emit,
	"Emitter.AddHandler":        (*of.Emitter).AddHandler,
	"Emitter.Handlers":          (*of.Emitter).Handlers,
	"Emitter.RemoveAllHandlers": (*of.Emitter).RemoveAllHandlers,
	"Emitter.SetLogger":         (*of.Emitter).SetLogger,
	"HookData":                  reflect.TypeFor[of.HookData](),
	"HookData.Set":              (*of.HookData).Set,
	"HookData.Get":              (*of.HookData).Get,
	"HookData.Has":              (*of.HookData).Has,
	"HookData.Delete":           (*of.HookData).Delete,
	"HookData.Clear":            (*of.HookData).Clear,
	"Error":                     reflect.TypeFor[of.Error](),
	"NewError":                  of.NewError,
	"Logger.Error":              of.Logger.Error,
	"Logger.Warn":               of.Logger.Warn,
	"Logger.Info":               of.Logger.Info,
	"Logger.Debug":              of.Logger.Debug,

	"Event":                     reflect.TypeFor[of.Event](),
	"EventReady":                of.EventReady,
	"EventError":                of.EventError,
	"EventConfigurationChanged": of.EventConfigurationChanged,
	"EventContextChanged":       of.EventContextChanged,
	"EventReconciling":          of.EventReconciling,
	"EventStale":                of.EventStale,

	"Status":            reflect.TypeFor[of.Status](),
	"StatusNotReady":    of.StatusNotReady,
	"StatusReady":       of.StatusReady,
	"StatusError":       of.StatusError,
	"StatusStale":       of.StatusStale,
	"StatusFatal":       of.StatusFatal,
	"StatusReconciling": of.StatusReconciling,

	"ErrorCode":               reflect.TypeFor[of.ErrorCode](),
	"CodeProviderNotReady":    of.CodeProviderNotReady,
	"CodeProviderFatal":       of.CodeProviderFatal,
	"CodeFlagNotFound":        of.CodeFlagNotFound,
	"CodeParseError":          of.CodeParseError,
	"CodeTypeMismatch":        of.CodeTypeMismatch,
	"CodeTargetingKeyMissing": of.CodeTargetingKeyMissing,
	"CodeInvalidContext":      of.CodeInvalidContext,
	"CodeGeneral":             of.CodeGeneral,

	"Reason":               reflect.TypeFor[of.Reason](),
	"ReasonStatic":         of.ReasonStatic,
	"ReasonDefault":        of.ReasonDefault,
	"ReasonTargetingMatch": of.ReasonTargetingMatch,
	"ReasonSplit":          of.ReasonSplit,
	"ReasonCached":         of.ReasonCached,
	"ReasonDisabled":       of.ReasonDisabled,
	"ReasonUnknown":        of.ReasonUnknown,
	"ReasonStale":          of.ReasonStale,
	"ReasonError":          of.ReasonError,

	"FlagType":    reflect.TypeFor[of.FlagType](),
	"TypeBoolean": of.TypeBoolean,
	"TypeString":  of.TypeString,
	"TypeNumber":  of.TypeNumber,
	"TypeObject":  of.TypeObject,

	"TelemetryAttribute*": []string{
		of.TelemetryAttributeKey, of.TelemetryAttributeErrorCode, of.TelemetryAttributeErrorMessage,
		of.TelemetryAttributeVariant, of.TelemetryAttributeValue, of.TelemetryAttributeContextID,
		of.TelemetryAttributeReason, of.TelemetryAttributeProvider, of.TelemetryAttributeFlagSetID,
		of.TelemetryAttributeVersion,
	},
	"TelemetryAttributeKey":          of.TelemetryAttributeKey,
	"TelemetryAttributeErrorCode":    of.TelemetryAttributeErrorCode,
	"TelemetryAttributeErrorMessage": of.TelemetryAttributeErrorMessage,
	"TelemetryAttributeVariant":      of.TelemetryAttributeVariant,
	"TelemetryAttributeValue":        of.TelemetryAttributeValue,
	"TelemetryAttributeContextID":    of.TelemetryAttributeContextID,
	"TelemetryAttributeReason":       of.TelemetryAttributeReason,
	"TelemetryAttributeProvider":     of.TelemetryAttributeProvider,
	"TelemetryAttributeFlagSetID":    of.TelemetryAttributeFlagSetID,
	"TelemetryAttributeVersion":      of.TelemetryAttributeVersion,
	"TelemetryFlagMetadata*": []string{
		of.TelemetryFlagMetadataContextID, of.TelemetryFlagMetadataFlagSetID, of.TelemetryFlagMetadataVersion,
	},
	"TelemetryFlagMetadataContextID": of.TelemetryFlagMetadataContextID,
	"TelemetryFlagMetadataFlagSetID": of.TelemetryFlagMetadataFlagSetID,
	"TelemetryFlagMetadataVersion":   of.TelemetryFlagMetadataVersion,

	"Context":           reflect.TypeFor[of.Context](),
	"FlagMetadata":      reflect.TypeFor[of.FlagMetadata](),
	"ResolutionDetails": reflect.TypeFor[of.ResolutionDetails[any]](),
	"EvaluationDetails": reflect.TypeFor[of.EvaluationDetails[any]](),
	"Metadata":          reflect.TypeFor[of.Metadata](),
	"ClientMetadata":    reflect.TypeFor[of.ClientMetadata](),
	"TrackingDetails":   reflect.TypeFor[of.TrackingDetails](),
	"Logger":            reflect.TypeFor[of.Logger](),
	"Handler":           reflect.TypeFor[of.Handler](),
	"EventDetails":      reflect.TypeFor[of.EventDetails](),
	"Hook":              reflect.TypeFor[of.Hook](),
	"HookContext":       reflect.TypeFor[of.HookContext](),
	"HookHints":         reflect.TypeFor[of.HookHints](),

	"Provider":                reflect.TypeFor[of.Provider](),
	"Provider.Metadata":       of.Provider.Metadata,
	"Provider.ResolveBoolean": of.Provider.ResolveBoolean,
	"Provider.ResolveString":  of.Provider.ResolveString,
	"Provider.ResolveNumber":  of.Provider.ResolveNumber,
	"Provider.ResolveObject":  of.Provider.ResolveObject,
	"Initializer":             reflect.TypeFor[of.Initializer](),
	"ContextChanger":          reflect.TypeFor[of.ContextChanger](),
	"Closer":                  reflect.TypeFor[of.Closer](),
	"Tracker":                 reflect.TypeFor[of.Tracker](),
	"Hooked":                  reflect.TypeFor[of.Hooked](),
	"Eventer":                 reflect.TypeFor[of.Eventer](),
	"RunsOn":                  reflect.TypeFor[of.RunsOn](),
	"DomainScoped":            reflect.TypeFor[of.DomainScoped](),

	"memprovider.Provider":                  reflect.TypeFor[memprovider.Provider](),
	"memprovider.Provider.PutConfiguration": (*memprovider.Provider).PutConfiguration,
	"memprovider.Provider.ResolveBoolean":   (*memprovider.Provider).ResolveBoolean,
	"memprovider.Provider.ResolveString":    (*memprovider.Provider).ResolveString,
	"memprovider.Provider.ResolveNumber":    (*memprovider.Provider).ResolveNumber,
	"memprovider.Provider.ResolveObject":    (*memprovider.Provider).ResolveObject,
}

var internalKeys = []string{
	"GenericEventEmitter", "InMemoryProvider.lookupFlagValue", "InMemoryProvider.resolveAndCheckFlag",
	"InMemoryProvider.resolveFlagWithReason", "LOG_LEVELS", "OpenFeature.allProviders",
	"OpenFeature.assertDomainScopedBindingAllowed", "OpenFeature.buildAndCacheEventEmitterForClient",
	"OpenFeature.clearProvidersAndSetDefault", "OpenFeature.getAssociatedEventEmitters",
	"OpenFeature.getProviderForClient", "OpenFeature.getProviderInstanceBinding", "OpenFeature.getUnboundEmitters",
	"OpenFeature.handleShutdownError", "OpenFeature.runProviderContextChangeHandler", "OpenFeature.setAwaitableProvider",
	"OpenFeature.transferListeners", "ProviderWrapper", "SafeLogger", "TypedInMemoryProvider.lookupFlagValue",
	"TypedInMemoryProvider.resolveAndCheckFlag", "TypedInMemoryProvider.resolveFlagWithReason", "isObject", "isString",
	"objectOrUndefined", "statusMatchesEvent", "stringOrUndefined",
}

var absentKeys = []string{
	"type:Client.setContext", "type:EvaluationOptions.context", "type:Hook.before.returnValue",
	"type:TransactionContextPropagator",
}

type surfaceEntry struct {
	Go        json.RawMessage `json:"go"`
	Value     *string         `json:"value"`
	Code      string          `json:"code"`
	Deviation string          `json:"deviation"`
	Internal  string          `json:"internal"`
	Absent    string          `json:"absent"`
}

func TestSurface(t *testing.T) {
	data, err := os.ReadFile("surface.json")
	if err != nil {
		t.Fatal(err)
	}
	var surface map[string]surfaceEntry
	if err := json.Unmarshal(data, &surface); err != nil {
		t.Fatal(err)
	}
	codes := map[string]bool{}
	for k, v := range goSymbols {
		if c, ok := v.(of.ErrorCode); ok && k != "ErrorCode" {
			codes[string(c)] = true
		}
	}
	deviation := regexp.MustCompile(`^D[1-8]$`)
	used := map[string]bool{}
	var internal, absent []string
	for member, e := range surface {
		switch {
		case e.Go != nil:
			var syms []string
			if err := json.Unmarshal(e.Go, &syms); err != nil {
				var s string
				if err := json.Unmarshal(e.Go, &s); err != nil {
					t.Errorf("%s: bad go field %s", member, e.Go)
					continue
				}
				syms = []string{s}
			}
			for _, sym := range syms {
				used[sym] = true
				v, ok := goSymbols[sym]
				if !ok {
					t.Errorf("%s: Go symbol %s missing", member, sym)
					continue
				}
				if e.Value != nil && fmt.Sprint(v) != *e.Value {
					t.Errorf("%s: %s = %q, want %q", member, sym, fmt.Sprint(v), *e.Value)
				}
			}
			if e.Code != "" && !codes[e.Code] {
				t.Errorf("%s: error code %s has no Go constant", member, e.Code)
			}
		case e.Deviation != "":
			if !deviation.MatchString(e.Deviation) {
				t.Errorf("%s: deviation %s is not in the closed list D1-D8", member, e.Deviation)
			}
		case e.Internal != "":
			internal = append(internal, member)
		case e.Absent != "":
			absent = append(absent, member)
		default:
			t.Errorf("%s: entry maps to nothing", member)
		}
	}
	slices.Sort(internal)
	slices.Sort(absent)
	if !slices.Equal(internal, internalKeys) {
		t.Errorf("internal entries = %q, want %q", internal, internalKeys)
	}
	if !slices.Equal(absent, absentKeys) {
		t.Errorf("absent entries = %q, want %q", absent, absentKeys)
	}
	for sym := range goSymbols {
		if !used[sym] {
			t.Errorf("Go symbol %s is not referenced by surface.json", sym)
		}
	}
}
