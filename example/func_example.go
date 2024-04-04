package main

import (
	"log"
	"time"

	dagRun "github.com/ycl2018/dag-run"
)

// this example shows that we have 4 function tasks(eg:TaskA、B、C、D) to run, which dependency relation like
// 		TaskA -- TaskB----\
//   		└ ---- TaskC--- TaskD
// each Task cost 100ms to run
// use dagScheduler, all these 4 tasks will run automatically by dependency relation, and total costs
// will only be about 300ms.

// 这个例子显示我们有4个任务(函数的形式)（例如：TaskA、B、C、D）要运行，
// 		TaskA -- TaskB----\
//   		└ ---- TaskC--- TaskD
// 每个Task花费100ms 使用 dagScheduler 运行，
// 这 4 个任务将通过依赖关系自动运行，总成本仅为 300ms 左右。

func main() {
	// runCtx contains "runParam" and collect all "outputs"
	// NOTE:
	//  dagScheduler does not provide concurrency security guarantees for runCtx,
	//  and situations that require concurrency security (such as concurrently writing maps) need to be maintained by the user

	// runCtx 提供运行参数，收集所有输出
	// NOTE: dagScheduler 不提供对runCtx的并发安全保证，需要并发安全的情况（如并发写map）需要使用方自己维护
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
