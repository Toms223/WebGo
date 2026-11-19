package openfeature

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

type panicLogger struct{}

func (panicLogger) Error(...any) { panic("boom") }
func (panicLogger) Warn(...any)  { panic("boom") }
func (panicLogger) Info(...any)  { panic("boom") }
func (panicLogger) Debug(...any) { panic("boom") }

func captureLog(t *testing.T) *bytes.Buffer {
	var b bytes.Buffer
	log.SetOutput(&b)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	return &b
}

func TestSafeLoggerRecoversPanic(t *testing.T) {
	b := captureLog(t)
	l := newSafeLogger(panicLogger{})
	l.Error("e")
	l.Warn("w")
	l.Info("i")
	l.Debug("d")
	if b.Len() == 0 {
		t.Fatal("expected fallback to default logger for Error/Warn")
	}
}

func TestDefaultLoggerDropsInfoDebug(t *testing.T) {
	b := captureLog(t)
	l := newSafeLogger(nil)
	if got := b.String(); !strings.Contains(got, "TypeError: Cannot read properties of undefined (reading 'error')\n") || !strings.HasSuffix(got, "Falling back to the default logger.\n") {
		t.Fatalf("nil logger printed %q", got)
	}
	b.Reset()
	l.Info("i")
	l.Debug("d")
	if b.Len() != 0 {
		t.Fatalf("got %q", b.String())
	}
	l.Error("e")
	l.Warn("w")
	if b.Len() == 0 {
		t.Fatal("expected output")
	}
}
