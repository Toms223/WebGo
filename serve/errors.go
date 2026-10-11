package serve

import "errors"

var (
	ErrMissingStyle   = errors.New("serve: stylesheet file not found")
	ErrPathOutsideWeb = errors.New("serve: path is not under /web/")
	ErrNotCSSPath     = errors.New("serve: generated path must end in .css")
	ErrDuplicatePath  = errors.New("serve: stylesheet path used twice")
)
