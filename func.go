package dagRun

import (
	"context"
)

type nopeCtx struct{}

// FuncScheduler simple scheduler for func tasks
type FuncScheduler struct {
	scd *Scheduler[nopeCtx]
}

// NewFuncScheduler build a func task scheduler
func NewFuncScheduler() *FuncScheduler {
	return &FuncScheduler{
		scd: NewScheduler[nopeCtx](),
	}
}

func (d *FuncScheduler) WithInjectorFactory(injectFac InjectorFactory[nopeCtx]) *FuncScheduler {
	d.scd.injectorFac = injectFac
	return d
}

// Submit provide func task to scheduler
// the param `name` is the taskID which should be unique, `deps` are the
// names of tasks that this task dependents, `f` defines what this task really does
func (d *FuncScheduler) Submit(name string, f func() error, deps ...string) *FuncScheduler {
	_ = d.scd.SubmitFunc(name, func(ctx context.Context, t nopeCtx) error {
		return f()
	}, deps...)
	return d
}

// Err check if any error happens
func (d *FuncScheduler) Err() error {
	return d.scd.err
}

// Run start all tasks and block till all of them done or meet critical err
func (d *FuncScheduler) Run() error {
	return d.scd.Run(context.Background(), nopeCtx{})
}

// Dot dump dag in dot language
func (d *FuncScheduler) Dot() string {
	return d.scd.Dot()
}

type nopeTaskImpl[T any] struct {
	name string
	deps []string
	f    func(context.Context, T) error
}

func (t nopeTaskImpl[T]) Name() string {
	return t.name
}

func (t nopeTaskImpl[T]) Dependencies() []string {
	return t.deps
}

func (t nopeTaskImpl[T]) Execute(ctx context.Context, t2 T) error {
	return t.f(ctx, t2)
}
