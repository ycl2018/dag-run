package dagRun

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

// Scheduler simple scheduler for typed tasks
type Scheduler[T any] struct {
	dag         *Graph
	nodes       map[string]*node[T]
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

// OptTask extends Task with options
type OptTask[T any] interface {
	Task[T]
	Options() []TaskOption
}

// ConditionBranch start a branch which execute when ValidBranch() returns true
type ConditionBranch[T any] interface {
	Task[T]
	ValidBranch() (valid bool)
}

type node[T any] struct {
	ds   *Scheduler[T]
	next []*node[T]
	swg  sync.WaitGroup
	task Task[T]
	// initial value is inDegree, but zero inDegree set to 1 to ensure node
	// can execute once, cancelled if value <= 0
	cancelled atomic.Int64
}

func (n *node[T]) Name() string {
	return n.task.Name()
}

func (n *node[T]) start(ctx context.Context, t T) {
	go func() {
		n.swg.Wait()
		var err error
		defer func() {
			if pErr := recover(); pErr != nil {
				n.ds.CancelWithErr(fmt.Errorf("panic:%v \n%s", pErr, debug.Stack()))
			}
			if err != nil {
				n.ds.CancelWithErr(err)
			}
			if ct, ok := n.task.(interface{ ValidBranch() (valid bool) }); ok {
				if valid := ct.ValidBranch(); !valid {
					n.breakNext()
				}
			}
			n.ds.swg.Done()
			for _, n2 := range n.next {
				n2.swg.Done()
			}
		}()
		if n.ds.err != nil {
			return
		}
		// condition Task
		if n.cancelled.Load() == 0 {
			// broadcast cancelled to next nodes
			n.breakNext()
			return
		}

		var execute = func() {
			var o option
			opT, ok := n.task.(interface{ Options() []TaskOption })
			if ok {
				ops := opT.Options()
				for _, op := range ops {
					op(&o)
				}
			}
			err = n.executeTask(ctx, n.task, t, o)
		}
		if n.ds.injectorFac != nil {
			inject := n.ds.injectorFac.Inject(ctx, n.task)
			if inject.Pre != nil {
				err = inject.Pre(ctx, t)
				if err != nil {
					return
				}
			}
			execute()
			if inject.After != nil {
				err = inject.After(ctx, t, err)
			}
		} else {
			execute()
		}
	}()
}
func (n *node[T]) breakNext() {
	for _, n2 := range n.next {
		n2.cancelled.Add(-1)
	}
}

func (n *node[T]) executeTask(ctx context.Context, task Task[T], t T, op option) error {
	var err error
	var runWithRetry = func() {
		if op.retry < 1 {
			op.retry = 1
		}
		for i := 0; i < op.retry; i++ {
			err = task.Execute(ctx, t)
			if err == nil {
				break
			}
		}
	}
	if op.timeout > 0 {
		var done = make(chan struct{})
		go func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("task:%s panic:%v", task.Name(), r)
				}
				done <- struct{}{}
			}()
			runWithRetry()
		}()
		select {
		case <-time.After(op.timeout):
			// timeout
			err = fmt.Errorf("task:%s run timeout", task.Name())
		case <-done:
		}
	} else {
		runWithRetry()
	}
	return err
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
		d.dag.AddNode(n)
		d.nodes[task.Name()] = n
		d.swg.Add(1)
	}
	return nil
}

// SubmitFunc submit a func task to scheduler
func (d *Scheduler[T]) SubmitFunc(name string, f func(context.Context, T) error, deps ...string) error {
	return d.SubmitFuncWithOps(name, f, nil, deps...)
}

// SubmitFuncWithOps submit a func task to scheduler with options
func (d *Scheduler[T]) SubmitFuncWithOps(name string, f func(context.Context, T) error, ops []TaskOption, deps ...string) error {
	if name == "" {
		d.err = ErrNoTaskName
		return d.err
	}
	if f == nil {
		d.err = ErrNilFunc
		return d.err
	}
	d.err = d.Submit(&defaultTaskImpl[T]{name: name, deps: deps, f: f, options: ops})
	return d.err
}

// Run start all tasks and block till all of them done or meet critical err
func (d *Scheduler[T]) Run(ctx context.Context, x T) error {
	if d.err != nil {
		return d.err
	}
	d.sealed = true
	for _, n := range d.nodes {
		for _, name := range n.task.Dependencies() {
			pre, ok := d.nodes[name]
			if !ok {
				return fmt.Errorf("%w: task :%s's dependency:%s not found", ErrTaskNotExist, n.Name(), name)
			}
			d.dag.AddEdge(pre, n)
			pre.next = append(pre.next, n)
			n.swg.Add(1)
		}
	}
	if err := d.dag.BFS(NopeWalker); err != nil {
		return err
	}
	// init nodes inDegrees
	inDegrees := make(map[string]int, len(d.nodes))
	for _, n := range d.nodes {
		inDegrees[n.Name()] = 0
	}
	for _, to := range d.dag.Edges {
		for _, n := range to {
			inDegrees[n.Name()]++
		}
	}

	for _, n := range d.nodes {
		if inDegrees[n.Name()] == 0 {
			// 0 inDegree nodes set 1 to ensure node execute at least once
			n.cancelled.Store(1)
		} else {
			n.cancelled.Store(int64(inDegrees[n.Name()]))
		}
	}

	for _, n := range d.nodes {
		n.start(ctx, x)
	}
	d.swg.Wait()
	return d.err
}

// CancelWithErr cancel the tasks which has not been stated in the scheduler
func (d *Scheduler[T]) CancelWithErr(err error) {
	d.lock.Lock()
	if d.err == nil {
		d.err = err
	} else {
		// group errors
		d.err = fmt.Errorf("%s,<%s>", d.err, err.Error())
	}
	d.lock.Unlock()
}

// Dot dump dag in dot language
func (d *Scheduler[T]) Dot() string {
	return d.dag.DOT()
}
