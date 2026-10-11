package style

import "errors"

var (
	ErrNoName             = errors.New("style: rule has no class, id or tag")
	ErrBadSelector        = errors.New("style: selector cannot be expressed")
	ErrBadMediaQuery      = errors.New("style: media query cannot be expressed")
	ErrBadBinding         = errors.New("style: binding needs an action name and a function that fits the rule")
	ErrDuplicateRule      = errors.New("style: selector defined twice in one generated stylesheet")
	ErrDynamicInGenerated = errors.New("style: rule with On passed to Generated")
	ErrNotDynamic         = errors.New("style: rule without On passed to Dynamic")
	ErrBadAtRule          = errors.New("style: at-rule is missing a required part")
	ErrBadSupportsQuery   = errors.New("style: supports query cannot be expressed")
	ErrBadContainerQuery  = errors.New("style: container query cannot be expressed")
	ErrNested             = errors.New("style: rule is nested inside another rule")
	ErrNestedBinding      = errors.New("style: nested rule cannot carry a binding")
)
