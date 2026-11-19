package openfeature

// Event is a provider lifecycle event name.
type Event string

// Status is a provider status.
type Status string

// ErrorCode is an evaluation error code.
type ErrorCode string

// Reason is a resolution reason; any string is allowed.
type Reason string

// FlagType is the type of a flag value.
type FlagType string

const (
	EventReady                Event = "PROVIDER_READY"
	EventError                Event = "PROVIDER_ERROR"
	EventConfigurationChanged Event = "PROVIDER_CONFIGURATION_CHANGED"
	EventContextChanged       Event = "PROVIDER_CONTEXT_CHANGED"
	EventReconciling          Event = "PROVIDER_RECONCILING"
	EventStale                Event = "PROVIDER_STALE"
)

const (
	StatusNotReady    Status = "NOT_READY"
	StatusReady       Status = "READY"
	StatusError       Status = "ERROR"
	StatusStale       Status = "STALE"
	StatusFatal       Status = "FATAL"
	StatusReconciling Status = "RECONCILING"
)

const (
	CodeProviderNotReady    ErrorCode = "PROVIDER_NOT_READY"
	CodeProviderFatal       ErrorCode = "PROVIDER_FATAL"
	CodeFlagNotFound        ErrorCode = "FLAG_NOT_FOUND"
	CodeParseError          ErrorCode = "PARSE_ERROR"
	CodeTypeMismatch        ErrorCode = "TYPE_MISMATCH"
	CodeTargetingKeyMissing ErrorCode = "TARGETING_KEY_MISSING"
	CodeInvalidContext      ErrorCode = "INVALID_CONTEXT"
	CodeGeneral             ErrorCode = "GENERAL"
)

const (
	ReasonStatic         Reason = "STATIC"
	ReasonDefault        Reason = "DEFAULT"
	ReasonTargetingMatch Reason = "TARGETING_MATCH"
	ReasonSplit          Reason = "SPLIT"
	ReasonCached         Reason = "CACHED"
	ReasonDisabled       Reason = "DISABLED"
	ReasonUnknown        Reason = "UNKNOWN"
	ReasonStale          Reason = "STALE"
	ReasonError          Reason = "ERROR"
)

const (
	TypeBoolean FlagType = "boolean"
	TypeString  FlagType = "string"
	TypeNumber  FlagType = "number"
	TypeObject  FlagType = "object"
)
