package main

import (
	"context"
	"github.com/ycl2018/dag-run"
	"log"
	"time"
)

// this example shows that we have 4 tasks(TaskA、B、C、D、E) to run, which dependency relation like
// 		TaskA -- TaskB----\
//   		└ ---- TaskC--- TaskD-----TaskE
// each Task cost 100ms to run
// use dagScheduler, all these 5 tasks will run automatically by dependency relation, and total costs
// will only be about 400ms.

// 这个例子显示我们有5个任务（TaskA、B、C、D、E）要运行，
// 		TaskA -- TaskB----\
//   		└ ---- TaskC--- TaskD-----TaskE
// 每个Task花费100ms 使用 dagScheduler 运行，
// 这 5 个任务将通过依赖关系自动运行，总成本仅为 400ms 左右。

func main1() {
	scd := dagRun.NewScheduler[*RunCtx]()
	err := scd.Submit(TaskA{}, TaskB{}, TaskC{}, TaskD{})
	if err != nil {
		log.Panicf("submit task err:%v\n", err)
		return
	}
	// you can submit a func Task
	err = scd.SubmitFunc("TaskE", []string{"TaskD"}, func(ctx context.Context, runCtx *RunCtx) error {
		time.Sleep(100 * time.Millisecond)
		runCtx.TaskEOutput = "TaskEOutput"
		return nil
	})
	if err != nil {
		log.Panicf("submit task err:%v\n", err)
		return
	}
	// runCtx contains "runParam" and collect all "outputs"
	var runCtx = &RunCtx{
		InputParam:  "InputParam",
		TaskAOutput: "",
		TaskBOutput: "",
		TaskCOutput: "",
		TaskDOutput: "",
		TaskEOutput: "",
	}
	fromTime := time.Now()
	err = scd.Run(context.Background(), runCtx)
	if err != nil {
		log.Panicf("run task err:%v", err)
		return
	}
	log.Printf("get runCtx:%+v\n", runCtx)
	log.Printf("cost time:%s\n", time.Since(fromTime))
	// example output
	// 2023/07/14 16:47:23 get runCtx:&{InputParam:InputParam TaskAOutput:TaskAOutput TaskBOutput:TaskBOutput TaskCOutput:TaskCOutput TaskDOutput:TaskDOutput}
	// 2023/07/14 16:47:23 cost time:301.472569ms
}

// RunCtx include inputParam which all task need and every task output values eg:TaskXOutput
type RunCtx struct {
	InputParam  string
	TaskAOutput string
	TaskBOutput string
	TaskCOutput string
	TaskDOutput string
	TaskEOutput string
}

const sleepTime = time.Millisecond * 100

type TaskA struct{}

func (t TaskA) Name() string {
	return "TaskA"
}

func (t TaskA) Dependencies() []string {
	return nil
}

func (t TaskA) Execute(ctx context.Context, runCtx *RunCtx) error {
	time.Sleep(sleepTime)
	runCtx.TaskAOutput = "TaskAOutput"
	return nil
}

type TaskB struct{}

func (t TaskB) Name() string {
	return "TaskB"
}

func (t TaskB) Dependencies() []string {
	return []string{"TaskA"}
}

func (t TaskB) Execute(ctx context.Context, runCtx *RunCtx) error {
	time.Sleep(sleepTime)
	runCtx.TaskBOutput = "TaskBOutput"
	return nil
}

type TaskC struct{}

func (t TaskC) Name() string {
	return "TaskC"
}

func (t TaskC) Dependencies() []string {
	return []string{"TaskA"}
}

func (t TaskC) Execute(ctx context.Context, runCtx *RunCtx) error {
	time.Sleep(sleepTime)
	runCtx.TaskCOutput = "TaskCOutput"
	return nil
}

type TaskD struct{}

func (t TaskD) Name() string {
	return "TaskD"
}

func (t TaskD) Dependencies() []string {
	return []string{"TaskB", "TaskC"}
}

func (t TaskD) Execute(ctx context.Context, runCtx *RunCtx) error {
	time.Sleep(sleepTime)
	runCtx.TaskDOutput = "TaskDOutput"
	return nil
}
