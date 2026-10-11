package topic

type State struct {
	Source   string
	Intro    string
	Sections []Section
	Active   string
	Err      string
	Loaded   bool
}
