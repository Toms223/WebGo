package app

// Library is a component library that ships its own CSS. The Handler serves
// styles at path and links it on every page; a path that does not start with
// "/" or empty styles are skipped.
type Library interface {
	Styles() (path, styles string)
}
