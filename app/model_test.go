package app

import (
	"encoding/json"
	"reflect"
	"testing"
)

type contextProbe struct {
	Compo
	Tick     int
	captured *Context
}

func (p *contextProbe) OnUpdate(ctx Context) { *p.captured = ctx }

func (p *contextProbe) Render() UI { return Div() }

func newModelContext(t *testing.T) (Context, TestEngine) {
	t.Helper()

	engine := NewTestEngine()
	var ctx Context

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

func storedCounter(ctx Context, key string) counter {
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

func TestNewModelDoesNotTouchTheStore(t *testing.T) {
	ctx, _ := newModelContext(t)
	ctx.SetState("user", &counter{N: 12, Label: "from store"})

	value := &counter{N: 3, Label: "default"}
	m, err := NewModel("user", value)
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	if value.N != 3 || value.Label != "default" {
		t.Fatalf("expected the seed to be left alone, got %+v", value)
	}
	if m.key != "user" {
		t.Fatalf("expected the key to be kept, got %q", m.key)
	}
	if m.state != value {
		t.Fatalf("expected the model to hold the caller's pointer, got %p want %p", m.state, value)
	}
}

func TestTheStoredStateOnlyArrivesOnGet(t *testing.T) {
	ctx, _ := newModelContext(t)
	ctx.SetState("deferred", &counter{N: 12, Label: "from store"})

	m, err := NewModel("deferred", &counter{N: 3, Label: "default"})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	m.Get(ctx)

	if m.state.N != 12 || m.state.Label != "from store" {
		t.Fatalf("expected Get to be what fills the model from the store, got %+v", m.state)
	}
}

func TestNewModelAdoptsTheCallersPointer(t *testing.T) {
	value := &counter{N: 1}
	m, err := NewModel("adopted", value)
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	value.N = 99

	if m.state.N != 99 {
		t.Fatalf("expected the model to share the caller's pointer, got %+v", m.state)
	}
}

func TestNewModelRejectsANilValue(t *testing.T) {
	ctx, _ := newModelContext(t)
	ctx.SetState("nil-seed", &counter{N: 5})

	_, err := NewModel[counter]("nil-seed", nil)

	if err == nil {
		t.Fatalf("expected a nil seed to be rejected at construction")
	}
	if err.Error() != "value is nil" {
		t.Fatalf("expected a value of value is nil, got %q", err.Error())
	}
}

func TestLoadApplySetPublishesTheNewStateToTheStore(t *testing.T) {
	ctx, _ := newModelContext(t)
	m, err := NewModel("published", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	m.Load(ctx).
		Apply(func(ctx Context, state *counter) *counter { return &counter{N: state.N + 1} }).
		Set()

	if got := storedCounter(ctx, "published"); got.N != 2 {
		t.Fatalf("expected the store to hold the applied state, got %+v", got)
	}
}

func TestLoadRefreshesTheStateBeforeTheActionRuns(t *testing.T) {
	ctx, _ := newModelContext(t)
	m, err := NewModel("reloaded", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	ctx.SetState("reloaded", &counter{N: 10})

	var seen int
	m.Load(ctx).Apply(func(ctx Context, state *counter) *counter {
		seen = state.N
		return state
	})

	if seen != 10 {
		t.Fatalf("expected the verb to see the state as Load found it, got %d", seen)
	}
}

func TestApplyHandsTheVerbTheContextLoadWasGiven(t *testing.T) {
	ctx, _ := newModelContext(t)
	m, err := NewModel("verb-ctx", &counter{})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	var seen Context
	m.Load(ctx).Apply(func(ctx Context, state *counter) *counter {
		seen = ctx
		return state
	})

	if seen.Src() != ctx.Src() {
		t.Fatal("expected the verb to run with the context the action was loaded with")
	}
}

func TestSetWithoutApplyRepublishesTheCurrentState(t *testing.T) {
	ctx, _ := newModelContext(t)
	m, err := NewModel("republished", &counter{N: 4})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	m.Load(ctx).Set()

	if got := storedCounter(ctx, "republished"); got.N != 4 {
		t.Fatalf("expected an action with no verbs to publish the state as it stands, got %+v", got)
	}
}

func TestApplyCallsVerbsInOrderAndFeedsTheResultForward(t *testing.T) {
	ctx, _ := newModelContext(t)
	m, err := NewModel("chained", &counter{N: 0})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	var seen []int
	step := func(ctx Context, state *counter) *counter {
		seen = append(seen, state.N)
		return &counter{N: state.N + 1}
	}
	m.Load(ctx).Apply(step).Apply(step).Apply(step).Set()

	if !reflect.DeepEqual(seen, []int{0, 1, 2}) {
		t.Fatalf("expected each verb to receive the previous result, got %v", seen)
	}
	if got := storedCounter(ctx, "chained"); got.N != 3 {
		t.Fatalf("expected the last result to be published, got %+v", got)
	}
}

func TestAVerbReturningNilDoesNotBreakTheModel(t *testing.T) {
	ctx, _ := newModelContext(t)
	m, err := NewModel("nil-verb", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	m.Load(ctx).
		Apply(func(ctx Context, state *counter) *counter { return nil }).
		Set()

	if panicked, value := didPanic(func() { m.Get(ctx) }); panicked {
		t.Fatalf("expected the model to survive a verb returning nil, but Get panicked: %v", value)
	}
}

func TestTheActionChainReturnsTheSameModel(t *testing.T) {
	ctx, _ := newModelContext(t)
	m, err := NewModel("identity", &counter{})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	if got := m.Load(ctx).Set(); got != m {
		t.Fatalf("expected Set to return the original model, got %p want %p", got, m)
	}
}

func TestSetDoesNotPersist(t *testing.T) {
	ctx, _ := newModelContext(t)
	m, err := NewModel("volatile", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	m.Load(ctx).
		Apply(func(ctx Context, state *counter) *counter { return &counter{N: 2} }).
		Set()

	if ctx.LocalStorage().Contains("volatile") {
		t.Fatalf("expected Set to stay in memory, but %q reached local storage", "volatile")
	}
}

func TestApplyMutatesTheModelEvenWhenSetIsNeverCalled(t *testing.T) {
	ctx, _ := newModelContext(t)
	m, err := NewModel("dropped", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	m.Load(ctx).Apply(func(ctx Context, state *counter) *counter { return &counter{N: 42} })

	if m.state.N != 42 {
		t.Fatalf("expected a dropped action to still have moved the model, got %+v", m.state)
	}
	if got := storedCounter(ctx, "dropped"); got.N != 0 {
		t.Fatalf("expected the store to be untouched by a dropped action, got %+v", got)
	}
}

func TestTheInMemoryStoreAliasesTheModelsPointer(t *testing.T) {
	ctx, _ := newModelContext(t)
	m, err := NewModel("aliased", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	m.Load(ctx).
		Apply(func(ctx Context, state *counter) *counter { return &counter{N: 2} }).
		Set()

	m.state.N = 666

	if got := storedCounter(ctx, "aliased"); got.N != 666 {
		t.Fatalf("expected the store to hold the model's pointer, got %+v", got)
	}
}

func TestSliceFieldsAreSharedBetweenTheStoreAndTheModel(t *testing.T) {
	ctx, _ := newModelContext(t)
	tags := []string{"a", "b"}
	m, err := NewModel("slices", &counter{Tags: tags})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	m.Load(ctx).Set()

	tags[0] = "mutated"

	if got := storedCounter(ctx, "slices"); got.Tags[0] != "mutated" {
		t.Fatalf("expected the backing array to be shared, got %+v", got)
	}
}

func TestPersistSnapshotsTheStateAtTheMomentItIsWritten(t *testing.T) {
	ctx, _ := newModelContext(t)
	m, err := NewModel("persisted-snapshot", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	m.Load(ctx).SetAndPersist(&counter{N: 2})

	m.state.N = 666

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
		t.Fatalf("expected the in-memory store to still alias the model, got %+v", got)
	}
}

func TestDispatchedUpdatesAllLand(t *testing.T) {
	ctx, engine := newModelContext(t)
	m, err := NewModel("dispatched", &counter{N: 0})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	increment := func(ctx Context, state *counter) *counter {
		return &counter{N: state.N + 1}
	}

	ctx.Dispatch(func(ctx Context) { m.Load(ctx).Apply(increment).Set() })
	ctx.Dispatch(func(ctx Context) { m.Load(ctx).Apply(increment).Set() })
	engine.ConsumeAll()

	if got := storedCounter(ctx, "dispatched"); got.N != 2 {
		t.Fatalf("expected both dispatched increments to land, got %+v", got)
	}
}

func TestAsyncWorkLandsWhenItIsDispatchedBack(t *testing.T) {
	ctx, engine := newModelContext(t)
	m, err := NewModel("async", &counter{N: 0})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	increment := func(ctx Context, state *counter) *counter {
		return &counter{N: state.N + 1}
	}

	for i := 0; i < 2; i++ {
		ctx.Async(func() {
			ctx.Dispatch(func(ctx Context) { m.Load(ctx).Apply(increment).Set() })
		})
	}
	engine.ConsumeAll()

	if got := storedCounter(ctx, "async"); got.N != 2 {
		t.Fatalf("expected both async increments to land, got %+v", got)
	}
}

func TestSetAndPersistWritesTheStateToLocalStorage(t *testing.T) {
	ctx, _ := newModelContext(t)
	m, err := NewModel("durable", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	m.Load(ctx).SetAndPersist(&counter{N: 55, Label: "saved"})

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
	ctx, _ := newModelContext(t)
	m, err := NewModel("durable-mem", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	m.Load(ctx).SetAndPersist(&counter{N: 9})

	if m.state.N != 9 {
		t.Fatalf("expected the model to hold the persisted value, got %+v", m.state)
	}
	if got := storedCounter(ctx, "durable-mem"); got.N != 9 {
		t.Fatalf("expected the in-memory store to be updated too, got %+v", got)
	}
}

func TestGetRefreshesTheStateFromTheStore(t *testing.T) {
	ctx, _ := newModelContext(t)
	m, err := NewModel("refreshed", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	ctx.SetState("refreshed", &counter{N: 31, Label: "elsewhere"})
	m.Get(ctx)

	if m.state.N != 31 || m.state.Label != "elsewhere" {
		t.Fatalf("expected Get to pull the newest state, got %+v", m.state)
	}
}

func TestGetLeavesTheStateAloneWhenTheKeyWasNeverWritten(t *testing.T) {
	ctx, _ := newModelContext(t)
	m, err := NewModel("absent", &counter{N: 6})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	m.Get(ctx)

	if m.state.N != 6 {
		t.Fatalf("expected Get on an unknown key to be a no-op, got %+v", m.state)
	}
}

func TestGetSurvivesATypeMismatch(t *testing.T) {
	ctx, _ := newModelContext(t)
	m, err := NewModel("wrong-type", &counter{N: 6})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	type other struct{ Message string }
	ctx.SetState("wrong-type", &other{Message: "boom"})
	m.Get(ctx)

	if m.state.N != 6 {
		t.Fatalf("expected the state to survive a mismatched read, got %+v", m.state)
	}
}

func TestObserveNotifiesOnSet(t *testing.T) {
	ctx, engine := newModelContext(t)
	m, err := NewModel("observed", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	changes := 0
	m.Observe(ctx).OnChange(func() { changes++ })

	m.Load(ctx).
		Apply(func(ctx Context, state *counter) *counter { return &counter{N: 2} }).
		Set()
	engine.ConsumeAll()

	if changes != 1 {
		t.Fatalf("expected exactly one change notification, got %d", changes)
	}
	if m.state.N != 2 {
		t.Fatalf("expected the observed state to be up to date, got %+v", m.state)
	}
}

func TestObserveDoesNotNotifyOnADirectFieldWrite(t *testing.T) {
	ctx, engine := newModelContext(t)
	m, err := NewModel("silent", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	changes := 0
	m.Observe(ctx).OnChange(func() { changes++ })

	m.state.N = 5
	engine.ConsumeAll()

	if changes != 0 {
		t.Fatalf("expected a direct write to bypass observers, got %d notifications", changes)
	}
}

func TestTwoModelsOnTheSameKeyShareState(t *testing.T) {
	ctx, _ := newModelContext(t)
	writer, writerErr := NewModel("shared", &counter{N: 1})
	reader, readerErr := NewModel("shared", &counter{})
	if writerErr != nil {
		t.Fatalf("creating writer model failed: %v", writerErr)
	}
	if readerErr != nil {
		t.Fatalf("creating reader model failed: %v", readerErr)
	}
	writer.Load(ctx).
		Apply(func(ctx Context, state *counter) *counter { return &counter{N: 21} }).
		Set()
	reader.Get(ctx)

	if reader.state.N != 21 {
		t.Fatalf("expected the second model to see the write, got %+v", reader.state)
	}
	if reader.State() == writer.State() {
		t.Fatal("expected the two models to hold different pointers")
	}
}

func TestApplyMayReturnAFreshPointerWithoutDetachingObservers(t *testing.T) {
	ctx, engine := newModelContext(t)
	m, err := NewModel("fresh", &counter{N: 1})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	other, err := NewModel("fresh", &counter{})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	m.Get(ctx).Observe(ctx)

	m.Load(ctx).
		Apply(func(_ Context, s *counter) *counter { return &counter{N: s.N + 1} }).
		Set()
	other.Load(ctx).
		Apply(func(_ Context, _ *counter) *counter { return &counter{N: 40} }).
		Set()
	engine.ConsumeAll()

	if m.State().N != 40 {
		t.Fatalf("expected the observing model to follow the second write, got %+v", m.State())
	}
}

func TestObserveAloneLoadsThePersistedState(t *testing.T) {
	ctx, _ := newModelContext(t)
	payload, err := json.Marshal(counter{N: 9, Label: "kept"})
	if err != nil {
		t.Fatalf("encoding the payload failed: %v", err)
	}
	if err := ctx.LocalStorage().Set("persisted-observe", storableState{Value: payload}); err != nil {
		t.Fatalf("seeding local storage failed: %v", err)
	}
	reader, err := NewModel("persisted-observe", &counter{})
	if err != nil {
		t.Fatalf("creating model failed: %v", err)
	}
	reader.Observe(ctx)

	if reader.State().N != 9 || reader.State().Label != "kept" {
		t.Fatalf("expected Observe to load the persisted state, got %+v", reader.State())
	}
}

func TestObservingModelsKeepTheirOwnStructs(t *testing.T) {
	ctx, engine := newModelContext(t)
	m, _ := NewModel("own-structs", &counter{})
	other, _ := NewModel("own-structs", &counter{})
	m.Observe(ctx)

	other.Load(ctx).Apply(func(_ Context, s *counter) *counter { return &counter{N: 3} }).Set()
	engine.ConsumeAll()
	other.State().N = 99

	if m.State().N != 3 || m.State() == other.State() {
		t.Fatalf("expected an independent copy holding 3, got %+v", m.State())
	}
}
