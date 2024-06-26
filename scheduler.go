package dagRun

import (
	"context"
	"fmt"
	"net/url"
	"runtime/debug"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Scheduler simple scheduler for typed tasks
type Scheduler[T any] struct {
	dag         *Graph
	nodes       map[string]*node[T]
	swg         *sync.WaitGroup
	lock        sync.Mutex
	err         error
	injectorFac InjectorFactory[T]
	sealed      bool
	done        chan error
}

// Task is the interface all your tasks should implement
type Task[T any] interface {
	Name() string
	Dependencies() []string
	Execute(context.Context, T) error
}

// Optioned means Task with options
type Optioned interface {
	Options() []TaskOption
}

// Conditioned means task is a branch Task
type Conditioned interface {
	ValidBranch() (valid bool)
}

// OptTask extends Task with options, not really used here, only for benefit of implements
type OptTask[T any] interface {
	Task[T]
	Optioned
}

// ConditionBranch start a branch which execute when ValidBranch() returns true
// not really used here, only for benefit of implements
type ConditionBranch[T any] interface {
	Task[T]
	Conditioned
}

type node[T any] struct {
	ds       *Scheduler[T]
	next     []*node[T]
	task     Task[T]
	preBreak atomic.Int64
}

func (n *node[T]) Name() string {
	return n.task.Name()
}

func (n *node[T]) start(ctx context.Context, t T) {
	go func() {
		var err error
		var breakNext bool
		defer func() {
			if pErr := recover(); pErr != nil {
				n.ds.CancelWithErr(fmt.Errorf("dag: panic:%v \n%s", pErr, debug.Stack()))
			}
			if err != nil {
				n.ds.CancelWithErr(err)
			}
			if breakNext {
				n.breakNext()
			} else {
				// when task is a branch node
				if ct, ok := n.task.(Conditioned); ok {
					if valid := ct.ValidBranch(); !valid {
						n.breakNext()
					}
				}
			}
			n.ds.swg.Done()
		}()
		//  break next nodes on this branch
		if n.preBreak.Load() == 0 {
			breakNext = true
			return
		}
		var execute = func() {
			var o option
			opT, ok := n.task.(Optioned)
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
					err = fmt.Errorf("dag: task:%s panic:%v", task.Name(), r)
				}
				done <- struct{}{}
			}()
			runWithRetry()
		}()
		select {
		case <-time.After(op.timeout):
			// timeout
			err = fmt.Errorf("dag: task:%s run timeout", task.Name())
		case <-done:
		}
	} else {
		runWithRetry()
	}
	return err
}

func (n *node[T]) breakNext() {
	for _, n2 := range n.next {
		n2.preBreak.Add(-1)
	}
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
	d.err = d.Submit(&funcTaskImpl[T]{name: name, deps: deps, f: f, options: ops})
	return d.err
}

// SubmitBranchFunc submit a func branch task to scheduler
func (d *Scheduler[T]) SubmitBranchFunc(name string, f func(context.Context, T) (bool, error), deps ...string) error {
	return d.SubmitBranchFuncWithOps(name, f, nil, deps...)
}

// SubmitBranchFuncWithOps submit a func branch task to scheduler with options
func (d *Scheduler[T]) SubmitBranchFuncWithOps(name string, f func(context.Context, T) (bool, error), ops []TaskOption, deps ...string) error {
	if name == "" {
		d.err = ErrNoTaskName
		return d.err
	}
	if f == nil {
		d.err = ErrNilFunc
		return d.err
	}
	d.err = d.Submit(&branchFuncTaskImpl[T]{name: name, deps: deps, f: f, options: ops})
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
				return fmt.Errorf("dag:%w: task :%s's dependency:%s not found", ErrTaskNotExist, n.Name(), name)
			}
			d.dag.AddEdge(pre, n)
			pre.next = append(pre.next, n)
		}
	}
	var visitedNodesNum int
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
	var toStartNodes []*node[T]
	for name, inDegree := range inDegrees {
		curN := d.nodes[name]
		if inDegree == 0 {
			toStartNodes = append(toStartNodes, curN)
			// set dummy head for start nodes to ensure start nodes execute once
			curN.preBreak.Store(int64(1))
		} else {
			curN.preBreak.Store(int64(inDegree))
		}
	}
	for len(toStartNodes) > 0 {
		d.swg = new(sync.WaitGroup)
		d.swg.Add(len(toStartNodes))
		for _, n := range toStartNodes {
			n.start(ctx, x)
			visitedNodesNum++
		}
		d.swg.Wait()
		if d.err != nil {
			return d.err
		}
		pre := toStartNodes
		toStartNodes = []*node[T]{}
		for _, startNode := range pre {
			for _, n := range startNode.next {
				inDegrees[n.Name()]--
				if inDegrees[n.Name()] == 0 {
					toStartNodes = append(toStartNodes, n)
				}
			}
		}
	}
	// check circle
	if visitedNodesNum < len(d.dag.Nodes) {
		var circleNodes []string
		for n, inDegree := range inDegrees {
			if inDegree != 0 {
				circleNodes = append(circleNodes, n)
			}
		}
		sort.Slice(circleNodes, func(i, j int) bool {
			return circleNodes[i] < circleNodes[j]
		})
		return fmt.Errorf("dag:graph has circle in nodes:%v", circleNodes)
	}

	return nil
}

// RunAsync start all tasks not block
func (d *Scheduler[T]) RunAsync(ctx context.Context, x T) {
	d.done = make(chan error)
	go func() {
		err := d.Run(ctx, x)
		d.done <- err
	}()
	return
}

// Wait all tasks finish, goroutine will block here if not
func (d *Scheduler[T]) Wait() error {
	if d.done == nil {
		return ErrNotAsyncJob
	}
	return <-d.done
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
func (d *Scheduler[T]) Dot(ops ...DotOption) string {
	var branchNodesOps []DotOption
	for _, n := range d.nodes {
		if _, ok := n.task.(Conditioned); ok {
			branchNodesOps = append(branchNodesOps, WithNodeAttr(n.Name(), "shape=diamond", `color="blue"`))
		}
	}
	return d.dag.DOT(append(branchNodesOps, ops...)...)
}

const graphvizOnlineURL = "https://dreampuf.github.io/GraphvizOnline/#"

// DOTOnlineURL return a graphvizOnline url
func (d *Scheduler[T]) DOTOnlineURL(ops ...DotOption) string {
	dot := d.Dot(ops...)
	escape := url.PathEscape(dot)
	return graphvizOnlineURL + escape
}
