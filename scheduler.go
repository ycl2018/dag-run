package dagRun

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
)

type Scheduler[T any] struct {
	dag   *Graph
	nodes map[string]*node[T]
	tasks []*node[T]
	swg   sync.WaitGroup
	lock  sync.Mutex
	err   error
}

type Task[T any] interface {
	Name() string
	Dependencies() []string
	Execute(context.Context, T) error
}

type taskImpl[T any] struct {
	name string
	deps []string
	f    func(context.Context, T) error
}

func (t taskImpl[T]) Name() string {
	return t.name
}

func (t taskImpl[T]) Dependencies() []string {
	return t.deps
}

func (t taskImpl[T]) Execute(ctx context.Context, t2 T) error {
	return t.f(ctx, t2)
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
			err = n.task.Execute(ctx, t)
		}
	}()
}

func NewScheduler[T any]() *Scheduler[T] {
	return &Scheduler[T]{dag: NewGraph(), nodes: make(map[string]*node[T], 0)}
}

func (d *Scheduler[T]) Submit(tasks ...Task[T]) error {
	for _, task := range tasks {
		if task == nil {
			return errors.New("submit nil task")
		}
		if _, has := d.nodes[task.Name()]; has {
			return fmt.Errorf("submit failed, task:%s already exist", task.Name())
		}
		n := &node[T]{task: task, ds: d}
		d.tasks = append(d.tasks, n)
		d.dag.AddNode(n)
		d.nodes[task.Name()] = n
		d.swg.Add(1)
	}

	return nil
}

func (d *Scheduler[T]) SubmitFunc(name string, deps []string, f func(context.Context, T) error) error {
	if name == "" {
		return errors.New("empty task name")
	}
	if f == nil {
		return errors.New("nil func")
	}
	task := &taskImpl[T]{name: name, deps: deps, f: f}
	if _, has := d.nodes[task.Name()]; has {
		return fmt.Errorf("submit failed, task:%s already exist", task.name)
	}
	n := &node[T]{task: task, ds: d}
	d.tasks = append(d.tasks, n)
	d.dag.AddNode(n)
	d.nodes[task.Name()] = n
	d.swg.Add(1)

	return nil
}

func (d *Scheduler[T]) Run(ctx context.Context, x T) error {
	for _, task := range d.tasks {
		for _, name := range task.task.Dependencies() {
			pre, ok := d.nodes[name]
			if !ok {
				return fmt.Errorf("task :%s's dependency:%s is not submited", task, name)
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

func (d *Scheduler[T]) CancelWithErr(err error) {
	d.lock.Lock()
	if d.err == nil {
		d.err = err
	}
	d.lock.Unlock()
}
