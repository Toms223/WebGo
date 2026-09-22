package model

import (
	"reflect"
	"testing"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type contextProbe struct {
	app.Compo
	Tick     int
	captured *app.Context
}

func (p *contextProbe) OnUpdate(ctx app.Context) { *p.captured = ctx }

func (p *contextProbe) Render() app.UI { return app.Div() }

func newContext(t *testing.T) (app.Context, app.TestEngine) {
	t.Helper()

	engine := app.NewTestEngine()
	var ctx app.Context

	if err := engine.Load(&contextProbe{captured: &ctx, Tick: 0}); err != nil {
		t.Fatalf("mounting the context probe failed: %v", err)
	}
	if err := engine.Load(&contextProbe{Tick: 1}); err != nil {
		t.Fatalf("updating the context probe failed: %v", err)
	}
	engine.ConsumeAll()

	if ctx.Src() == nil {
		t.Fatal("the context probe was never updated, so no Context was captured")
	}
	return ctx, engine
}

type counter struct {
	N     int
	Tags  []string
	Label string
}

func storedCounter(ctx app.Context, key string) counter {
	var out counter
	ctx.GetState(key, &out)
	return out
}

func didPanic(f func()) (panicked bool, value any) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
			value = r
		}
	}()
	f()
	return false, nil
}

func exportedMethods(t reflect.Type) []string {
	var names []string
	for method := range t.Methods() {
		names = append(names, method.Name)
	}
	return names
}

func TestNewViewModelDoesNotTouchTheStore(t *testing.T) {
	ctx, _ := newContext(t)
	ctx.SetState("user", &counter{N: 12, Label: "from store"})

	value := &counter{N: 3, Label: "default"}
	vm, err := NewViewModel("user", value)
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	if value.N != 3 || value.Label != "default" {
		t.Fatalf("expected the seed to be left alone, got %+v", value)
	}
	if vm.key != "user" {
		t.Fatalf("expected the key to be kept, got %q", vm.key)
	}
	if vm.state != value {
		t.Fatalf("expected the view model to hold the caller's pointer, got %p want %p", vm.state, value)
	}
}

func TestTheStoredStateOnlyArrivesOnGet(t *testing.T) {
	ctx, _ := newContext(t)
	ctx.SetState("deferred", &counter{N: 12, Label: "from store"})

	vm, err := NewViewModel("deferred", &counter{N: 3, Label: "default"})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	vm.Get(ctx)

	if vm.state.N != 12 || vm.state.Label != "from store" {
		t.Fatalf("expected Get to be what fills the model from the store, got %+v", vm.state)
	}
}

func TestNewViewModelAdoptsTheCallersPointer(t *testing.T) {
	value := &counter{N: 1}
	vm, err := NewViewModel("adopted", value)
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	value.N = 99

	if vm.state.N != 99 {
		t.Fatalf("expected the view model to share the caller's pointer, got %+v", vm.state)
	}
}

func TestNewViewModelRejectsANilValue(t *testing.T) {
	ctx, _ := newContext(t)
	ctx.SetState("nil-seed", &counter{N: 5})

	_, err := NewViewModel[counter]("nil-seed", nil)

	if err == nil {
		t.Fatalf("expected a nil seed to be rejected at construction")
	}
	if err.Error() != "value is nil" {
		t.Fatalf("expected a value of value is nil, got %q", err.Error())
	}
}

func TestLoadApplySetPublishesTheNewStateToTheStore(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("published", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	vm.Load(ctx).
		Apply(func(ctx app.Context, state *counter) *counter { return &counter{N: state.N + 1} }).
		Set()

	if got := storedCounter(ctx, "published"); got.N != 2 {
		t.Fatalf("expected the store to hold the applied state, got %+v", got)
	}
}

func TestLoadRefreshesTheStateBeforeTheActionRuns(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("reloaded", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	ctx.SetState("reloaded", &counter{N: 10})

	var seen int
	vm.Load(ctx).Apply(func(ctx app.Context, state *counter) *counter {
		seen = state.N
		return state
	})

	if seen != 10 {
		t.Fatalf("expected the verb to see the state as Load found it, got %d", seen)
	}
}

func TestApplyHandsTheVerbTheContextLoadWasGiven(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("verb-ctx", &counter{})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	var seen app.Context
	vm.Load(ctx).Apply(func(ctx app.Context, state *counter) *counter {
		seen = ctx
		return state
	})

	if seen.Src() != ctx.Src() {
		t.Fatal("expected the verb to run with the context the action was loaded with")
	}
}

func TestSetWithoutApplyRepublishesTheCurrentState(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("republished", &counter{N: 4})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	vm.Load(ctx).Set()

	if got := storedCounter(ctx, "republished"); got.N != 4 {
		t.Fatalf("expected an action with no verbs to publish the state as it stands, got %+v", got)
	}
}

func TestApplyCallsVerbsInOrderAndFeedsTheResultForward(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("chained", &counter{N: 0})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	var seen []int
	step := func(ctx app.Context, state *counter) *counter {
		seen = append(seen, state.N)
		return &counter{N: state.N + 1}
	}
	vm.Load(ctx).Apply(step).Apply(step).Apply(step).Set()

	if !reflect.DeepEqual(seen, []int{0, 1, 2}) {
		t.Fatalf("expected each verb to receive the previous result, got %v", seen)
	}
	if got := storedCounter(ctx, "chained"); got.N != 3 {
		t.Fatalf("expected the last result to be published, got %+v", got)
	}
}

func TestAVerbReturningNilDoesNotBreakTheModel(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("nil-verb", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	vm.Load(ctx).
		Apply(func(ctx app.Context, state *counter) *counter { return nil }).
		Set()

	if panicked, value := didPanic(func() { vm.Get(ctx) }); panicked {
		t.Fatalf("expected the model to survive a verb returning nil, but Get panicked: %v", value)
	}
}

func TestTheActionChainReturnsTheSameViewModel(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("identity", &counter{})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	if got := vm.Load(ctx).Set(); got != vm {
		t.Fatalf("expected Set to return the original view model, got %p want %p", got, vm)
	}
}

func TestSetDoesNotPersist(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("volatile", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	vm.Load(ctx).
		Apply(func(ctx app.Context, state *counter) *counter { return &counter{N: 2} }).
		Set()

	if ctx.LocalStorage().Contains("volatile") {
		t.Fatalf("expected Set to stay in memory, but %q reached local storage", "volatile")
	}
}

func TestApplyMutatesTheViewModelEvenWhenSetIsNeverCalled(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("dropped", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	vm.Load(ctx).Apply(func(ctx app.Context, state *counter) *counter { return &counter{N: 42} })

	if vm.state.N != 42 {
		t.Fatalf("expected a dropped action to still have moved the view model, got %+v", vm.state)
	}
	if got := storedCounter(ctx, "dropped"); got.N != 0 {
		t.Fatalf("expected the store to be untouched by a dropped action, got %+v", got)
	}
}

func TestTheInMemoryStoreAliasesTheViewModelsPointer(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("aliased", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	vm.Load(ctx).
		Apply(func(ctx app.Context, state *counter) *counter { return &counter{N: 2} }).
		Set()

	vm.state.N = 666

	if got := storedCounter(ctx, "aliased"); got.N != 666 {
		t.Fatalf("expected the store to hold the view model's pointer, got %+v", got)
	}
}

func TestSliceFieldsAreSharedBetweenTheStoreAndTheViewModel(t *testing.T) {
	ctx, _ := newContext(t)
	tags := []string{"a", "b"}
	vm, err := NewViewModel("slices", &counter{Tags: tags})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	vm.Load(ctx).Set()

	tags[0] = "mutated"

	if got := storedCounter(ctx, "slices"); got.Tags[0] != "mutated" {
		t.Fatalf("expected the backing array to be shared, got %+v", got)
	}
}

func TestPersistSnapshotsTheStateAtTheMomentItIsWritten(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("persisted-snapshot", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	vm.Load(ctx).SetAndPersist(&counter{N: 2})

	vm.state.N = 666

	var stored struct {
		Value counter
	}
	if err := ctx.LocalStorage().Get("persisted-snapshot", &stored); err != nil {
		t.Fatalf("reading the persisted state failed: %v", err)
	}
	if stored.Value.N != 2 {
		t.Fatalf("expected local storage to keep the value as it was when persisted, got %+v", stored.Value)
	}
	if got := storedCounter(ctx, "persisted-snapshot"); got.N != 666 {
		t.Fatalf("expected the in-memory store to still alias the view model, got %+v", got)
	}
}

func TestDispatchedUpdatesAllLand(t *testing.T) {
	ctx, engine := newContext(t)
	vm, err := NewViewModel("dispatched", &counter{N: 0})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	increment := func(ctx app.Context, state *counter) *counter {
		return &counter{N: state.N + 1}
	}

	ctx.Dispatch(func(ctx app.Context) { vm.Load(ctx).Apply(increment).Set() })
	ctx.Dispatch(func(ctx app.Context) { vm.Load(ctx).Apply(increment).Set() })
	engine.ConsumeAll()

	if got := storedCounter(ctx, "dispatched"); got.N != 2 {
		t.Fatalf("expected both dispatched increments to land, got %+v", got)
	}
}

func TestAsyncWorkLandsWhenItIsDispatchedBack(t *testing.T) {
	ctx, engine := newContext(t)
	vm, err := NewViewModel("async", &counter{N: 0})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	increment := func(ctx app.Context, state *counter) *counter {
		return &counter{N: state.N + 1}
	}

	for i := 0; i < 2; i++ {
		ctx.Async(func() {
			ctx.Dispatch(func(ctx app.Context) { vm.Load(ctx).Apply(increment).Set() })
		})
	}
	engine.ConsumeAll()

	if got := storedCounter(ctx, "async"); got.N != 2 {
		t.Fatalf("expected both async increments to land, got %+v", got)
	}
}

func TestSetAndPersistWritesTheStateToLocalStorage(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("durable", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	vm.Load(ctx).SetAndPersist(&counter{N: 55, Label: "saved"})

	if !ctx.LocalStorage().Contains("durable") {
		t.Fatal("expected SetAndPersist to write to local storage")
	}

	var stored struct {
		Value counter
	}
	if err := ctx.LocalStorage().Get("durable", &stored); err != nil {
		t.Fatalf("reading the persisted state failed: %v", err)
	}
	if stored.Value.N != 55 || stored.Value.Label != "saved" {
		t.Fatalf("expected the persisted payload to carry the state, got %+v", stored.Value)
	}
}

func TestSetAndPersistAlsoUpdatesTheInMemoryState(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("durable-mem", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	vm.Load(ctx).SetAndPersist(&counter{N: 9})

	if vm.state.N != 9 {
		t.Fatalf("expected the view model to hold the persisted value, got %+v", vm.state)
	}
	if got := storedCounter(ctx, "durable-mem"); got.N != 9 {
		t.Fatalf("expected the in-memory store to be updated too, got %+v", got)
	}
}

func TestGetRefreshesTheStateFromTheStore(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("refreshed", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	ctx.SetState("refreshed", &counter{N: 31, Label: "elsewhere"})
	vm.Get(ctx)

	if vm.state.N != 31 || vm.state.Label != "elsewhere" {
		t.Fatalf("expected Get to pull the newest state, got %+v", vm.state)
	}
}

func TestGetLeavesTheStateAloneWhenTheKeyWasNeverWritten(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("absent", &counter{N: 6})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	vm.Get(ctx)

	if vm.state.N != 6 {
		t.Fatalf("expected Get on an unknown key to be a no-op, got %+v", vm.state)
	}
}

func TestGetSurvivesATypeMismatch(t *testing.T) {
	ctx, _ := newContext(t)
	vm, err := NewViewModel("wrong-type", &counter{N: 6})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	type other struct{ Message string }
	ctx.SetState("wrong-type", &other{Message: "boom"})
	vm.Get(ctx)

	if vm.state.N != 6 {
		t.Fatalf("expected the state to survive a mismatched read, got %+v", vm.state)
	}
}

func TestObserveNotifiesOnSet(t *testing.T) {
	ctx, engine := newContext(t)
	vm, err := NewViewModel("observed", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	changes := 0
	vm.Observe(ctx).OnChange(func() { changes++ })

	vm.Load(ctx).
		Apply(func(ctx app.Context, state *counter) *counter { return &counter{N: 2} }).
		Set()
	engine.ConsumeAll()

	if changes != 1 {
		t.Fatalf("expected exactly one change notification, got %d", changes)
	}
	if vm.state.N != 2 {
		t.Fatalf("expected the observed state to be up to date, got %+v", vm.state)
	}
}

func TestObserveDoesNotNotifyOnADirectFieldWrite(t *testing.T) {
	ctx, engine := newContext(t)
	vm, err := NewViewModel("silent", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating view model failed: %v", err)
	}
	changes := 0
	vm.Observe(ctx).OnChange(func() { changes++ })

	vm.state.N = 5
	engine.ConsumeAll()

	if changes != 0 {
		t.Fatalf("expected a direct write to bypass observers, got %d notifications", changes)
	}
}

func TestTwoViewModelsOnTheSameKeyShareState(t *testing.T) {
	ctx, _ := newContext(t)
	writer, writerErr := NewViewModel("shared", &counter{N: 1})
	reader, readerErr := NewViewModel("shared", &counter{})
	if writerErr != nil {
		t.Fatalf("creating writer view model failed: %v", writerErr)
	}
	if readerErr != nil {
		t.Fatalf("creating reader view model failed: %v", readerErr)
	}
	writer.Load(ctx).
		Apply(func(ctx app.Context, state *counter) *counter { return &counter{N: 21} }).
		Set()
	reader.Get(ctx)

	if reader.state.N != 21 {
		t.Fatalf("expected the second view model to see the write, got %+v", reader.state)
	}
}
