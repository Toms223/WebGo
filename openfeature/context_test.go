package openfeature

import (
	"errors"
	"testing"
)

type ctxProvider struct {
	*testProvider
	fail bool
}

func (p *ctxProvider) OnContextChange(_, _ Context) (Pending, error) {
	if p.fail {
		return nil, errors.New("boom")
	}
	return nil, nil
}

func TestContextIsCopied(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	in := Context{Attributes: map[string]any{"a": 1}}
	SetContext(in)
	in.Attributes["a"] = 2
	in.Attributes["b"] = 3
	got := GetContext().Attributes
	if len(got) != 1 || got["a"] != 1 {
		t.Fatalf("got %v", got)
	}
	got["a"] = 9
	if GetContext().Attributes["a"] != 1 {
		t.Fatal("GetContext result aliases stored context")
	}
}

func TestDeviation_ctx_on_change_sync_throw(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	p := &ctxProvider{testProvider: newTestProvider("ctxp")}
	if err := SetProviderAndWait(p); err != nil {
		t.Fatal(err)
	}
	var errs []string
	changed, reconciling := 0, 0
	AddHandler(EventReconciling, func(EventDetails) { reconciling++ })
	AddHandler(EventError, func(d EventDetails) { errs = append(errs, d.Message) })
	AddHandler(EventContextChanged, func(EventDetails) { changed++ })
	p.fail = true
	SetContext(Context{Attributes: map[string]any{"a": 1}})
	if ProviderStatus() != StatusError {
		t.Fatalf("status %v", ProviderStatus())
	}
	if len(errs) != 1 || errs[0] != "Error running ctxp's context change handler: boom" {
		t.Fatalf("errors %v", errs)
	}
	p.fail = false
	SetContext(Context{Attributes: map[string]any{"a": 2}})
	if changed != 1 || reconciling != 0 || ProviderStatus() != StatusReady {
		t.Fatalf("changed %d reconciling %d status %v", changed, reconciling, ProviderStatus())
	}
}
