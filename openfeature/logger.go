package openfeature

import "log"

// Logger receives SDK log output.
type Logger interface {
	Error(...any)
	Warn(...any)
	Info(...any)
	Debug(...any)
}

type defaultLogger struct{}

func (defaultLogger) Error(a ...any) { log.Println(a...) }
func (defaultLogger) Warn(a ...any)  { log.Println(a...) }
func (defaultLogger) Info(...any)    {}
func (defaultLogger) Debug(...any)   {}

type safeLogger struct{ l Logger }

func newSafeLogger(l Logger) Logger {
	if l == nil {
		defaultLogger{}.Error("TypeError: Cannot read properties of undefined (reading 'error')")
		defaultLogger{}.Error("Falling back to the default logger.")
		return defaultLogger{}
	}
	return safeLogger{l}
}

func (s safeLogger) call(f func(Logger), fallback func(Logger)) {
	defer func() {
		if recover() != nil {
			fallback(defaultLogger{})
		}
	}()
	f(s.l)
}

func (s safeLogger) Error(a ...any) {
	s.call(func(l Logger) { l.Error(a...) }, func(l Logger) { l.Error(a...) })
}

func (s safeLogger) Warn(a ...any) {
	s.call(func(l Logger) { l.Warn(a...) }, func(l Logger) { l.Warn(a...) })
}

func (s safeLogger) Info(a ...any) {
	s.call(func(l Logger) { l.Info(a...) }, func(l Logger) { l.Info(a...) })
}

func (s safeLogger) Debug(a ...any) {
	s.call(func(l Logger) { l.Debug(a...) }, func(l Logger) { l.Debug(a...) })
}

// DefaultLogger returns the logger the SDK uses when none is set.
func DefaultLogger() Logger { return defaultLogger{} }
