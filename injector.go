package dagRun

import "context"

type Injector[T any] struct {
	Pre   func(ctx context.Context, runCtx T)
	After func(ctx context.Context, runCtx T, err error) error
}

type InjectorFactory[T any] interface {
	Inject(ctx context.Context, task Task[T]) Injector[T]
}

type InjectorFactoryFunc[T any] func(ctx context.Context, task Task[T]) Injector[T]

func (i InjectorFactoryFunc[T]) Inject(ctx context.Context, task Task[T]) Injector[T] {
	return i(ctx, task)
}
