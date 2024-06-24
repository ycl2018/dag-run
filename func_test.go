package dagRun

import (
	"sync"
	"testing"
)

func TestFuncScheduler(t *testing.T) {
	scd := NewFuncScheduler()
	runCtx := sync.Map{}
	scd.Submit("T1", func() error {
		runCtx.Store("T1", "T1")
		return nil
	})
	scd.Submit("T2", func() error {
		runCtx.Store("T2", "T2")
		return nil
	}, "T1")
	scd.Submit("T3", func() error {
		runCtx.Store("T3", "T3")
		return nil
	}, "T2")
	err := scd.Run()
	if err != nil {
		t.Errorf("scd run err:%v", err)
	}
	expectValues := []string{"T1", "T2", "T3"}
	for _, v := range expectValues {
		value, _ := runCtx.Load(v)
		if v != value.(string) {
			t.Errorf("expected:%s but get:%s", v, value)
		}
	}
	dotStr := scd.Dot()
	t.Log(dotStr)
}

func TestBranchFuncTask(t *testing.T) {
	scd := NewFuncScheduler()
	runCtx := sync.Map{}
	err := scd.
		Submit("T1", func() error { runCtx.Store("T1", "T1"); return nil }).
		Submit("T2", func() error { runCtx.Store("T2", "T2"); return nil }, "B1").
		Submit("T3", func() error { runCtx.Store("T3", "T3"); return nil }, "B2").
		Submit("T4", func() error { runCtx.Store("T4", "T4"); return nil }, "T2", "T3").
		SubmitBranch("B1", func() (bool, error) { return true, nil }, "T1").
		SubmitBranch("B2", func() (bool, error) { return false, nil }, "T1").
		Run()
	if err != nil {
		t.Errorf("scd run err:%v", err)
	}
	expectValues := []string{"T1", "T2", "T4"}
	for _, v := range expectValues {
		value, _ := runCtx.Load(v)
		if v != value.(string) {
			t.Errorf("expected:%s but get:%s", v, value)
		}
	}
	dotStr := scd.Dot()
	t.Log(dotStr)
}
