package main

import (
	"log"
	"time"

	dagRun "github.com/ycl2018/dag-run"
)

func main() {
	var runCtx = &RunCtx{
		InputParam:  "InputParam",
		TaskAOutput: "",
		TaskBOutput: "",
		TaskCOutput: "",
		TaskDOutput: "",
	}
	fromTime := time.Now()
	scd := dagRun.
		NewFuncScheduler().
		Submit("TaskA", func() error {
			time.Sleep(time.Millisecond * 100)
			runCtx.TaskAOutput = "TaskAOutput"
			return nil
		}).
		Submit("TaskB", func() error {
			time.Sleep(time.Millisecond * 100)
			// 可以安全使用上游依赖的输出
			log.Printf("TaskAOutPut:%s", runCtx.TaskAOutput)
			runCtx.TaskBOutput = "TaskBOutput"
			return nil
		}, "TaskA").
		Submit("TaskC", func() error {
			time.Sleep(time.Millisecond * 100)
			runCtx.TaskCOutput = "TaskCOutput"
			return nil
		}, "TaskA").
		Submit("TaskD", func() error {
			time.Sleep(time.Millisecond * 100)
			runCtx.TaskDOutput = "TaskDOutput"
			return nil
		}, "TaskB", "TaskC")
	err := scd.
		Run()
	if err != nil {
		log.Panicf("run task err:%v", err)
		return
	}
	log.Printf("get runCtx:%+v\n", runCtx)
	log.Printf("cost time:%s\n", time.Since(fromTime))
	// example output
	//2023/07/17 11:42:24 TaskAOutPut:TaskAOutput
	//2023/07/17 11:42:24 get runCtx:&{InputParam:InputParam TaskAOutput:TaskAOutput TaskBOutput:TaskBOutput TaskCOutput:TaskCOutput TaskDOutput:TaskDOutput TaskEOutput:}
	//2023/07/17 11:42:24 cost time:302.270165ms

}
