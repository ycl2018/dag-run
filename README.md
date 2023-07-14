# dag-run

## A simple multi-task concurrent scheduling library
<p>Generic implementation, support go1.18</p>
<p>Based on sync.WaitGroup, very simple and lightweight implementation</p>
<p>Multiple tasks with dependencies can be automatically scheduled to run concurrently according to their dependencies, minimizing the running time</p>
<p>Support fail fast, if a task returns an error during operation, the rest of the unrunning tasks will be canceled</p>

##  一个简单的多任务并发编排调度库
<p>泛型实现，支持go1.18</p>
<p>基于sync.WaitGroup，非常简单、轻量的实现</p>
<p>可以将具有依赖关系的多个任务，自动按其依赖关系并发调度运行，最小化运行时间</p>
<p>支持fail fast，运行中如果有任务返回错误，则取消其余未运行任务</p>

```go
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
		runCtx.TaskCOutput = "TaskCOutput"
		return nil
	})
	scd.SubmitFunc("TaskD", []string{"TaskB", "TaskC"}, func(ctx context.Context, runCtx *RunCtx) error {
		time.Sleep(time.Millisecond * 100)
		runCtx.TaskDOutput = "TaskDOutput"
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

```
