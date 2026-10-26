package app

type Tagger interface {
	Tags() Tags
}

type Tags map[string]string

func (t Tags) Tags() Tags {
	return t
}

func (t Tags) Set(name string, v any) {
	t[name] = toString(v)
}

func (t Tags) Get(name string) string {
	return t[name]
}

type Tag struct {
	Name  string
	Value string
}

func (t Tag) Tags() Tags {
	return Tags{t.Name: t.Value}
}

func T(name string, value any) Tag {
	return Tag{
		Name:  name,
		Value: toString(value),
	}
}
