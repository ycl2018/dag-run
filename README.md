# dag-run
<p>A simple multi-task concurrent scheduling library，Multiple tasks with dependencies can be automatically scheduled to run concurrently according to their dependencies, minimizing the running time</p>

``` go get github.com/ycl2018/dag-run ```
## Feature
- <p>install `go get github.com/ycl2018/dag-run`</p>
- <p>Generic implementation, support go1.18</p>
- <p>Based on sync.WaitGroup, very simple and lightweight implementation</p>
- <p>Support fail fast, if a task returns an error during operation, the rest of the unrunning tasks will be canceled</p>
- <p>You can use TaskManager to easily register and get your Task tasks</p>
- <p>support add injector</p>

## 中文说明

一个简单的多任务并发调度工具，可以将具有依赖关系的多个任务，自动按其依赖关系并发调度运行

``` go get github.com/ycl2018/dag-run ```

特性：

- <p>泛型实现，支持go1.18</p>
- <p>基于sync.WaitGroup，非常简单、轻量的实现</p>
- <p>支持fail fast，运行中如果有任务返回错误，则取消其余未运行任务</p>
- <p>可以使用TaskManager来方便的注册和获取你的Task任务</p>
- <p>支持提交函数任务/结构体任务</p>
- <p>支持注入injector，在每个任务执行前后插入通用的业务逻辑，如打点、监控等</p>

## Example1：提交函数
![example1](iamges/example1.png)
```go
err := dagRun.NewFuncScheduler().
		Submit("A", a).
		Submit("B", b, "A").
		Submit("C", c, "A").
		Submit("D", d, "B", "C").
		Run()

var a = func() error {return nil}
var b = func() error {return nil}
var c = func() error {return nil}
var d = func() error {return nil}
```
