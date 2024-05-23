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

// Submit func task to scheduler
func (d *FuncScheduler) Submit(name string, f func() error, deps ...string) *FuncScheduler {
	_ = d.scd.SubmitFunc(name, func(ctx context.Context, t nopeCtx) error {
		return f()
	}, deps...)
	return d
}

// SubmitWithOps submit func task to scheduler with some options
func (d *FuncScheduler) SubmitWithOps(name string, f func() error, ops []TaskOption, deps ...string) *FuncScheduler {
	_ = d.scd.SubmitFuncWithOps(name, func(ctx context.Context, t nopeCtx) error {
		return f()
	}, ops, deps...)
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

type defaultTaskImpl[T any] struct {
	name    string
	deps    []string
	f       func(context.Context, T) error
	options []TaskOption
}

func (t defaultTaskImpl[T]) Options() []TaskOption {
	return t.options
}

func (t defaultTaskImpl[T]) Name() string {
	return t.name
}

func (t defaultTaskImpl[T]) Dependencies() []string {
	return t.deps
}

func (t defaultTaskImpl[T]) Execute(ctx context.Context, t2 T) error {
	return t.f(ctx, t2)
}
