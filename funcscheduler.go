package dagRun

import (
	"context"
	"go/types"
)

type FuncScheduler struct {
	scd *Scheduler[types.Nil]
}

func NewFuncScheduler() *FuncScheduler {
	return &FuncScheduler{
		scd: NewScheduler[types.Nil](),
	}
}

func (d *FuncScheduler) Submit(name string, deps []string, f func() error) *FuncScheduler {
	_ = d.scd.SubmitFunc(name, deps, func(ctx context.Context, t types.Nil) error {
		return f()
	})
	return d
}

func (d *FuncScheduler) Err() error {
	return d.scd.err
}

func (d *FuncScheduler) Run() error {
	return d.scd.Run(context.Background(), types.Nil{})
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
