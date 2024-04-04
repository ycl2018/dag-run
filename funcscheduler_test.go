package dagRun

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFuncScheduler(t *testing.T) {
	scd := NewFuncScheduler()
	runCtx := sync.Map{}
	scd.Submit("T1", func() error {
		runCtx.Store("T1", "T1")
		time.Sleep(1 * time.Second)
		return nil
	})
	scd.Submit("T2", func() error {
		runCtx.Store("T2", "T2")
		time.Sleep(1 * time.Second)
		return nil
	}, "T1")
	scd.Submit("T3", func() error {
		runCtx.Store("T3", "T3")
		time.Sleep(1 * time.Second)
		return nil
	}, "T2")
	assert.Nil(t, scd.Run())
	expectValues := []string{"T1", "T2", "T3"}
	for _, v := range expectValues {
		value, _ := runCtx.Load(v)
		assert.Equal(t, v, value.(string))
	}
}
