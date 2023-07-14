package main

import (
	"context"
	dagRun "github.com/ycl2018/dag-run"
	"log"
	"time"
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
	var runCtx = &RunCtx{
		InputParam:  "InputParam",
		TaskAOutput: "",
		TaskBOutput: "",
		TaskCOutput: "",
		TaskDOutput: "",
	}
	scd := dagRun.NewScheduler[*RunCtx]()
	scd.SubmitFunc("TaskA", nil, func(ctx context.Context, runCtx *RunCtx) error {
		time.Sleep(time.Millisecond * 100)
		runCtx.TaskAOutput = "TaskAOutput"
		return nil
	})
	scd.SubmitFunc("TaskB", []string{"TaskA"}, func(ctx context.Context, runCtx *RunCtx) error {
		time.Sleep(time.Millisecond * 100)
		runCtx.TaskBOutput = "TaskBOutput"
		return nil
	})
	scd.SubmitFunc("TaskC", []string{"TaskA"}, func(ctx context.Context, runCtx *RunCtx) error {
		time.Sleep(time.Millisecond * 100)
		runCtx.TaskBOutput = "TaskCOutput"
		return nil
	})
	scd.SubmitFunc("TaskD", []string{"TaskB", "TaskC"}, func(ctx context.Context, runCtx *RunCtx) error {
		time.Sleep(time.Millisecond * 100)
		runCtx.TaskBOutput = "TaskDOutput"
		return nil
	})
	fromTime := time.Now()
	err := scd.Run(context.Background(), runCtx)
	if err != nil {
		log.Panicf("run task err:%v", err)
		return
	}
	log.Printf("get runCtx:%+v\n", runCtx)
	log.Printf("cost time:%s\n", time.Since(fromTime))
	// example output
	// 2023/07/14 17:13:55 get runCtx:&{InputParam:InputParam TaskAOutput:TaskAOutput TaskBOutput:TaskBOutput TaskCOutput:TaskCOutput TaskDOutput:TaskDOutput}
	// 2023/07/14 17:13:55 cost time:301.555504ms
}
