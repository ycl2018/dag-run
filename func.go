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

// SubmitBranch submit branch func task to scheduler
func (d *FuncScheduler) SubmitBranch(name string, f func() (bool, error), deps ...string) *FuncScheduler {
	_ = d.scd.SubmitBranchFunc(name, func(ctx context.Context, t nopeCtx) (bool, error) {
		return f()
	}, deps...)
	return d
}

// SubmitBranchWithOpts submit branch func task to scheduler with some options
func (d *FuncScheduler) SubmitBranchWithOpts(name string, f func() (bool, error), ops []TaskOption, deps ...string) *FuncScheduler {
	_ = d.scd.SubmitBranchFuncWithOps(name, func(ctx context.Context, t nopeCtx) (bool, error) {
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
func (d *FuncScheduler) Dot(ops ...DotOption) string {
	return d.scd.Dot(ops...)
}

// DOTOnlineURL return a graphvizOnline url
func (d *FuncScheduler) DOTOnlineURL(ops ...DotOption) string {
	return d.scd.DOTOnlineURL(ops...)
}

type funcTaskImpl[T any] struct {
	name    string
	deps    []string
	f       func(context.Context, T) error
	options []TaskOption
}

func (t *funcTaskImpl[T]) Options() []TaskOption {
	return t.options
}

func (t *funcTaskImpl[T]) Name() string {
	return t.name
}

func (t *funcTaskImpl[T]) Dependencies() []string {
	return t.deps
}

func (t *funcTaskImpl[T]) Execute(ctx context.Context, t2 T) error {
	return t.f(ctx, t2)
}

type branchFuncTaskImpl[T any] struct {
	name        string
	deps        []string
	f           func(context.Context, T) (bool, error)
	options     []TaskOption
	validBranch bool
}

func (b *branchFuncTaskImpl[T]) Name() string {
	return b.name
}

func (b *branchFuncTaskImpl[T]) Dependencies() []string {
	return b.deps
}

func (b *branchFuncTaskImpl[T]) Execute(ctx context.Context, t T) error {
	valid, err := b.f(ctx, t)
	if err != nil {
		return err
	}
	b.validBranch = valid
	return nil
}

func (b *branchFuncTaskImpl[T]) ValidBranch() (valid bool) {
	return b.validBranch
}
