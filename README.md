# dag-run
A simple multi-task concurrent scheduling library
## English
- <p>install `go get github.com/ycl2018/dag-run`</p>
- <p>Generic implementation, support go1.18</p>
- <p>Based on sync.WaitGroup, very simple and lightweight implementation</p>
- <p>Multiple tasks with dependencies can be automatically scheduled to run concurrently according to their dependencies, minimizing the running time</p>
- <p>Support fail fast, if a task returns an error during operation, the rest of the unrunning tasks will be canceled</p>
- <p>You can use TaskManager to easily register and get your Task tasks</p>

## 中文说明
- 安装 `go get github.com/ycl2018/dag-run`
- <p>一个简单的多任务并发调度工具</p>
- <p>泛型实现，支持go1.18</p>
- <p>基于sync.WaitGroup，非常简单、轻量的实现</p>
- <p>可以将具有依赖关系的多个任务，自动按其依赖关系并发调度运行，最小化运行时间</p>
- <p>支持fail fast，运行中如果有任务返回错误，则取消其余未运行任务</p>
- <p>可以使用TaskManager来方便的注册和获取你的Task任务</p>
```go

package main

import (
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
	err := dagRun.
		NewFuncScheduler().
		Submit("TaskA", nil, func() error {
			time.Sleep(time.Millisecond * 100)
			runCtx.TaskAOutput = "TaskAOutput"
			return nil
		}).
		Submit("TaskB", []string{"TaskA"}, func() error {
			time.Sleep(time.Millisecond * 100)
			// 可以安全使用上游依赖的输出
			log.Printf("TaskAOutPut:%s", runCtx.TaskAOutput)
			runCtx.TaskBOutput = "TaskBOutput"
			return nil
		}).
		Submit("TaskC", []string{"TaskA"}, func() error {
			time.Sleep(time.Millisecond * 100)
			runCtx.TaskCOutput = "TaskCOutput"
			return nil
		}).
		Submit("TaskD", []string{"TaskB", "TaskC"}, func() error {
			time.Sleep(time.Millisecond * 100)
			runCtx.TaskDOutput = "TaskDOutput"
			return nil
		}).
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


```
