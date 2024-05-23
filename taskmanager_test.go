package dagRun

import (
	"context"
	"testing"
)

func checkNotNil(t *testing.T, value error) {
	t.Helper()
	if value == nil {
		t.Errorf("want not nil, but get nil")
	}
}

type myTask struct {
	name string
	deps []string
}

func (t myTask) Name() string {
	return t.name
}

func (t myTask) Dependencies() []string {
	return t.deps
}

func (t myTask) Execute(ctx context.Context, i interface{}) error {
	return nil
}

func TestTaskManager(t *testing.T) {
	var nodes = []struct {
		name string
		deps []string
	}{
		{name: "taskA", deps: nil},
		{name: "taskB", deps: []string{"taskA"}},
		{name: "taskC", deps: []string{"taskA"}},
		{name: "taskD", deps: []string{"taskB", "taskC"}},
	}

	lm := NewTaskManager[myTask]()
	for _, node := range nodes {
		lm.Register(myTask{name: node.name, deps: node.deps})
	}
	// test1
	tasks, err := lm.GetAllTaskWithDepsByName([]string{"taskD"})
	checkNil(t, err)
	checkEqual(t, 4, len(tasks))
	wants := []string{"taskA", "taskB", "taskC", "taskD"}
	for _, w := range wants {
		if _, ok := tasks[w]; !ok {
			t.Errorf("not get task:%s", w)
		}
	}
	// test2
	tasks, err = lm.GetAllTaskWithDepsByName([]string{"taskC"})
	checkNil(t, err)
	checkEqual(t, 2, len(tasks))
	wants = []string{"taskA", "taskC"}
	for _, w := range wants {
		if _, ok := tasks[w]; !ok {
			t.Errorf("not get task:%s", w)
		}
	}
}
