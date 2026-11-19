package openfeature

// Object is a JSON object that keeps its key order.
type Object []Field

// Field is one key of an Object.
type Field struct {
	Key   string
	Value any
}

// Get returns the value of the last field named key, as a JavaScript object
// built from o would.
func (o Object) Get(key string) (any, bool) {
	for i := len(o) - 1; i >= 0; i-- {
		if o[i].Key == key {
			return o[i].Value, true
		}
	}
	return nil, false
}

// MarshalJSON encodes o the way JavaScript's JSON.stringify encodes an
// object: array-index keys first in numeric order, then the other keys in
// slice order. A repeated key appears once, at its first position, with its
// last value.
func (o Object) MarshalJSON() ([]byte, error) {
	s, err := stringify(o)
	return []byte(s), err
}
