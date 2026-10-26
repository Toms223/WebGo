package errors

import (
	"encoding/json"
	"sync/atomic"
)

type encoderFunc func(any) ([]byte, error)

var encoder atomic.Value

func init() {
	SetInlineEncoder()
}

func SetEncoder(fn func(any) ([]byte, error)) {
	if fn == nil {
		panic("errors: nil encoder")
	}

	encoder.Store(encoderFunc(fn))
}

func SetInlineEncoder() {
	SetEncoder(json.Marshal)
}

func SetIndentEncoder() {
	SetEncoder(func(v any) ([]byte, error) {
		return json.MarshalIndent(v, "", "  ")
	})
}

func getEncoder() encoderFunc {
	return encoder.Load().(encoderFunc)
}
