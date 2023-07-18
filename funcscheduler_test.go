package dagRun

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestFuncScheduler(t *testing.T) {
	scd := NewFuncScheduler()
	runCtx := sync.Map{}
	scd.Submit("T1", nil, func() error {
		runCtx.Store("T1", "T1")
		time.Sleep(1 * time.Second)
		return nil
	})
	scd.Submit("T2", []string{"T1"}, func() error {
		runCtx.Store("T2", "T2")
		time.Sleep(1 * time.Second)
		return nil
	})
	scd.Submit("T3", []string{"T2"}, func() error {
		runCtx.Store("T3", "T3")
		time.Sleep(1 * time.Second)
		return nil
	})
	assert.Nil(t, scd.Run())
	expectValues := []string{"T1", "T2", "T3"}
	for _, v := range expectValues {
		value, _ := runCtx.Load(v)
		assert.Equal(t, v, value.(string))
	}
}
