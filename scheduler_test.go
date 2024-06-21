package dagRun

import (
	"context"
	"errors"
	"log"
	"sync"
	"testing"
	"time"
)

func checkNil(t *testing.T, value error) {
	t.Helper()
	if value != nil {
		t.Errorf("want nil but get:%v", value)
	}
}

func checkEqual[T comparable](t *testing.T, want, get T) {
	t.Helper()
	if want != get {
		t.Errorf("want:%v but get:%v", want, get)
	}
}

func checkGreater[T int | float64 | int32 | int64 | float32](t *testing.T, greater, than T) {
	t.Helper()
	if greater <= than {
		t.Errorf("want greater value than %v, but get:%v", than, greater)
	}
}

type task struct {
	name         string
	dependencies []string
	options      []TaskOption
}

func (t task) Options() []TaskOption {
	return t.options
}

func (t task) Name() string {
	return t.name
}

func (t task) Dependencies() []string {
	return t.dependencies
}

func (t task) Execute(_ context.Context, runCtx *sync.Map) error {
	time.Sleep(100 * time.Millisecond)
	runCtx.Store(t.Name(), t.Name())
	return nil
}

func TestNewScheduler(t *testing.T) {
	var nodes = []task{
		{
			name:         "T1",
			dependencies: nil,
		},
		{
			name:         "T2",
			dependencies: []string{"T1"},
		},
		{
			name:         "T3",
			dependencies: []string{"T1"},
		},
		{
			name:         "T4",
			dependencies: []string{"T2", "T3"},
		},
		{
			name:         "T5",
			dependencies: []string{},
		},
	}
	start := time.Now().UnixMilli()
	ds := NewScheduler[*sync.Map]()
	for _, mt := range nodes {
		checkNil(t, ds.Submit(mt))
	}
	runCtx := &sync.Map{}
	checkNil(t, ds.Run(context.Background(), runCtx))
	checkGreater(t, int64(400), time.Now().UnixMilli()-start)
	for _, n := range nodes {
		value, _ := runCtx.Load(n.name)
		checkEqual(t, n.name, value.(string))
	}
	dot := ds.Dot()
	checkEqual(t, `
digraph G {
"start"[shape=box,color="green"]
"end"[shape=box,color="red"]
"T1" -> {"T2","T3"}
"T2" -> {"T4"}
"T3" -> {"T4"}
"start" -> {"T1","T5"}
{"T4","T5"}  -> "end"
}
`, dot)
}

func TestExecuteDagWithPanic(t *testing.T) {
	var nodes = []task{
		{
			name:         "T1",
			dependencies: nil,
		},
		{
			name:         "T2",
			dependencies: []string{"T1"},
		},
		{
			name:         "T4",
			dependencies: []string{"T1"},
		},
		{
			name:         "T5",
			dependencies: []string{"T2", "T3"},
		},
		{
			name:         "T6",
			dependencies: []string{"T5", "T4"},
		},
	}
	start := time.Now().UnixMilli()
	ds := NewScheduler[*sync.Map]()
	for _, mt := range nodes {
		checkNil(t, ds.Submit(mt))
	}
	checkNil(t, ds.SubmitFunc("T3", func(ctx context.Context, s *sync.Map) error {
		panic("expect panic in task T3")
	}, "T2", "T4"))
	runCtx := &sync.Map{}
	err := ds.Run(context.Background(), runCtx)
	checkNotNil(t, err)
	checkGreater(t, int64(300), time.Now().UnixMilli()-start)
	expectRunTask, expectNotRunTask := []string{"T1", "T2", "T4"}, []string{"T3", "T5", "T6"}
	for _, name := range expectRunTask {
		value, ok := runCtx.Load(name)
		checkEqual(t, true, ok)
		checkEqual(t, name, value.(string))
	}
	for _, name := range expectNotRunTask {
		_, ok := runCtx.Load(name)
		checkEqual(t, false, ok)
	}
}

func TestExecuteDagWithError(t *testing.T) {
	var nodes = []task{
		{
			name:         "T1",
			dependencies: nil,
		},
		{
			name:         "T2",
			dependencies: []string{"T1"},
		},
		{
			name:         "T4",
			dependencies: []string{"T1"},
		},
		{
			name:         "T5",
			dependencies: []string{"T2", "T3"},
		},
		{
			name:         "T6",
			dependencies: []string{"T5", "T4"},
		},
	}
	start := time.Now().UnixMilli()
	ds := NewScheduler[*sync.Map]()
	for _, mt := range nodes {
		checkNil(t, ds.Submit(mt))
	}
	checkNil(t, ds.SubmitFunc("T3", func(ctx context.Context, _ *sync.Map) error {
		return errors.New("expect err in T3")
	}, "T2", "T4"))
	runCtx := &sync.Map{}
	err := ds.Run(context.Background(), runCtx)
	checkEqual(t, err.Error(), "expect err in T3")
	checkGreater(t, int64(300), time.Now().UnixMilli()-start)
	expectRunTask, expectNotRunTask := []string{"T1", "T2", "T4"}, []string{"T3", "T5", "T6"}
	for _, name := range expectRunTask {
		value, ok := runCtx.Load(name)
		checkEqual(t, true, ok)
		checkEqual(t, name, value.(string))
	}
	for _, name := range expectNotRunTask {
		_, ok := runCtx.Load(name)
		checkEqual(t, false, ok)
	}
}

func TestExecuteDagWithTimeout(t *testing.T) {
	var nodes = []task{
		{
			name:         "T1",
			dependencies: nil,
		},
		{
			name:         "T2",
			dependencies: []string{"T1"},
			options:      []TaskOption{Timeout(200 * time.Millisecond)},
		},
		{
			name:         "T3",
			dependencies: []string{"T2", "T4"},
			options:      []TaskOption{Timeout(50 * time.Millisecond)},
		},
		{
			name:         "T4",
			dependencies: []string{"T1"},
		},
		{
			name:         "T5",
			dependencies: []string{"T2", "T3"},
		},
		{
			name:         "T6",
			dependencies: []string{"T5", "T4"},
		},
	}
	start := time.Now().UnixMilli()
	ds := NewScheduler[*sync.Map]()
	for _, mt := range nodes {
		checkNil(t, ds.Submit(mt))
	}
	runCtx := &sync.Map{}
	err := ds.Run(context.Background(), runCtx)
	checkEqual(t, err.Error(), "task:T3 run timeout")
	checkGreater(t, int64(300), time.Now().UnixMilli()-start)
	expectRunTask, expectNotRunTask := []string{"T1", "T2", "T4"}, []string{"T3", "T5", "T6"}
	for _, name := range expectRunTask {
		value, ok := runCtx.Load(name)
		checkEqual(t, true, ok)
		checkEqual(t, name, value.(string))
	}
	for _, name := range expectNotRunTask {
		_, ok := runCtx.Load(name)
		checkEqual(t, false, ok)
	}
}

func TestSchedulerInjector(t *testing.T) {
	var nodes = []task{
		{
			name:         "T1",
			dependencies: nil,
		},
		{
			name:         "T2",
			dependencies: []string{"T1"},
		},
		{
			name:         "T3",
			dependencies: []string{"T1"},
		},
		{
			name:         "T4",
			dependencies: []string{"T2", "T3"},
		},
		{
			name:         "T5",
			dependencies: []string{},
		},
	}
	start := time.Now().UnixMilli()
	ds := NewScheduler[*sync.Map]()
	for _, mt := range nodes {
		checkNil(t, ds.Submit(mt))
	}
	ds = ds.WithInjectorFactory(InjectorFactoryFunc[*sync.Map](func(ctx context.Context, task Task[*sync.Map]) Injector[*sync.Map] {
		return Injector[*sync.Map]{
			Pre: func(ctx context.Context, runCtx *sync.Map) error {
				log.Printf("task:%s start at:%s\n", task.Name(), time.Now())
				return nil
			},
			After: func(ctx context.Context, runCtx *sync.Map, err error) error {
				log.Printf("task:%s end at:%s\n", task.Name(), time.Now())
				if err == nil {
					runCtx.Store(task.Name(), "modified by injector "+task.Name())
				}
				return err
			},
		}
	}))
	runCtx := &sync.Map{}
	checkNil(t, ds.Run(context.Background(), runCtx))
	checkGreater(t, int64(400), time.Now().UnixMilli()-start) // program expect running 3 seconds
	for _, n := range nodes {
		value, _ := runCtx.Load(n.name)
		checkEqual(t, "modified by injector "+n.name, value.(string))
	}
}

func TestExecuteDagconditionBranch(t *testing.T) {
	var nodes = []task{
		{
			name:         "T1",
			dependencies: nil,
		},
		{
			name:         "T2",
			dependencies: []string{"T1"},
		},
		{
			name:         "T4",
			dependencies: []string{"T3"},
		},
		{
			name:         "T5",
			dependencies: []string{"T2"},
		},
		{
			name:         "T6",
			dependencies: []string{"T4", "T5"},
		},
	}
	ds := NewScheduler[*sync.Map]()
	for _, mt := range nodes {
		checkNil(t, ds.Submit(mt))
	}

	checkNil(t, ds.Submit(conditionBranch{name: "T3", deps: []string{"T2"}, valid: false}))
	runCtx := &sync.Map{}
	err := ds.Run(context.Background(), runCtx)
	checkNil(t, err)
	expectRunTask, expectNotRunTask := []string{"T1", "T2", "T3", "T5", "T6"}, []string{"T4"}
	for _, name := range expectRunTask {
		value, ok := runCtx.Load(name)
		t.Log("check " + name)
		checkEqual(t, true, ok)
		checkEqual(t, name, value.(string))
	}
	for _, name := range expectNotRunTask {
		_, ok := runCtx.Load(name)
		checkEqual(t, false, ok)
	}
}

type conditionBranch struct {
	name  string
	deps  []string
	valid bool
}

func (c conditionBranch) Name() string {
	return c.name
}

func (c conditionBranch) Dependencies() []string {
	return c.deps
}

func (c conditionBranch) Execute(ctx context.Context, t *sync.Map) error {
	t.Store(c.name, c.name)
	return nil
}

func (c conditionBranch) ValidBranch() (execute bool) {
	return c.valid
}
