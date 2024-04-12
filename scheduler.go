package dagRun

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
)

// Scheduler simple scheduler for typed tasks
type Scheduler[T any] struct {
	dag         *Graph
	nodes       map[string]*node[T]
	tasks       []*node[T]
	swg         sync.WaitGroup
	lock        sync.Mutex
	err         error
	injectorFac InjectorFactory[T]
	sealed      bool
}

// Task is the interface all your tasks should implement
type Task[T any] interface {
	Name() string
	Dependencies() []string
	Execute(context.Context, T) error
}

type node[T any] struct {
	ds   *Scheduler[T]
	next []*node[T]
	swg  sync.WaitGroup
	task Task[T]
}

func (n *node[T]) String() string {
	return n.task.Name()
}

func (n *node[T]) start(ctx context.Context, t T) {
	go func() {
		var err error
		defer func() {
			if pErr := recover(); pErr != nil {
				n.ds.CancelWithErr(fmt.Errorf("panic:%v \n%s", pErr, debug.Stack()))
			}
			if err != nil {
				n.ds.CancelWithErr(err)
			}
			n.ds.swg.Done()
			for _, n2 := range n.next {
				n2.swg.Done()
			}
		}()
		n.swg.Wait()
		if n.ds.err == nil {
			if n.ds.injectorFac != nil {
				inject := n.ds.injectorFac.Inject(ctx, n.task)
				if inject.Pre != nil {
					inject.Pre(ctx, t)
				}
				err = n.task.Execute(ctx, t)
				if inject.After != nil {
					err = inject.After(ctx, t, err)
				}
			} else {
				err = n.task.Execute(ctx, t)
			}
		}
	}()
}

// NewScheduler build a typed task scheduler
func NewScheduler[T any]() *Scheduler[T] {
	return &Scheduler[T]{dag: NewGraph(), nodes: make(map[string]*node[T], 0)}
}

// WithInjectorFactory add a injectorFactory, which run before/after each task.
func (d *Scheduler[T]) WithInjectorFactory(injectFac InjectorFactory[T]) *Scheduler[T] {
	d.injectorFac = injectFac
	return d
}

// NewWithInjectorFactory is shortcut of NewScheduler.WithInjectorFactory
func NewWithInjectorFactory[T any](injectFac InjectorFactory[T]) *Scheduler[T] {
	s := NewScheduler[T]().WithInjectorFactory(injectFac)
	return s
}

// Submit provide typed task to scheduler, all task should implement interface Task
func (d *Scheduler[T]) Submit(tasks ...Task[T]) error {
	if d.sealed {
		return ErrSealed
	}
	for _, task := range tasks {
		if task == nil {
			d.err = ErrNilTask
			return d.err
		}
		if _, has := d.nodes[task.Name()]; has {
			d.err = ErrTaskExist
			return d.err
		}
		n := &node[T]{task: task, ds: d}
		d.tasks = append(d.tasks, n)
		d.dag.AddNode(n)
		d.nodes[task.Name()] = n
		d.swg.Add(1)
	}
	return nil
}

// SubmitFunc submit a func task to scheduler
func (d *Scheduler[T]) SubmitFunc(name string, f func(context.Context, T) error, deps ...string) error {
	if name == "" {
		d.err = ErrNoTaskName
		return d.err
	}
	if f == nil {
		d.err = ErrNilFunc
		return d.err
	}
	d.err = d.Submit(&nopeTaskImpl[T]{name: name, deps: deps, f: f})
	return d.err
}

// Run start all tasks and block till all of them done or meet critical err
func (d *Scheduler[T]) Run(ctx context.Context, x T) error {
	if d.err != nil {
		return d.err
	}
	d.sealed = true
	for _, task := range d.tasks {
		for _, name := range task.task.Dependencies() {
			pre, ok := d.nodes[name]
			if !ok {
				return fmt.Errorf("%w: task :%s's dependency:%s is not submited", ErrTaskNotExist, task, name)
			}
			d.dag.AddEdge(pre, task)
			pre.next = append(pre.next, task)
			task.swg.Add(1)
		}
	}
	if err := d.dag.DFS(NopeWalker); err != nil {
		return err
	}
	// start work
	if ctx == nil {
		ctx = context.Background()
	}
	for _, task := range d.tasks {
		task.start(ctx, x)
	}
	d.swg.Wait()
	return d.err
}

// CancelWithErr cancel the tasks which has not been stated in the scheduler
func (d *Scheduler[T]) CancelWithErr(err error) {
	d.lock.Lock()
	if d.err == nil {
		d.err = err
	}
	d.lock.Unlock()
}

// Dot dump dag in dot language
func (d *Scheduler[T]) Dot() string {
	return d.dag.DOT()
}
