### 下面两个定时任务库执行过程，修改系统时间会导致cpu使用率暴增或卡住
https://github.com/go-co-op/gocron (v1存在问题，v2可以使用参数进行限制 https://pkg.go.dev/github.com/go-co-op/gocron/v2#WithSingletonMode)

https://github.com/robfig/cron


```go
// 示例任务函数
func (t *T) task1(arg1 string, arg2 int) {
	fmt.Printf("Task 1: %s, %d === %d \n", arg1, arg2, t.i)
	t.i++
}

func task2(arg1 int, arg2 string, arg3 float64) {
	fmt.Printf("Task 2: %d, %s, %.2f\n", arg1, arg2, arg3)
}

type T struct {
	i int
}

func main() {
	tm := NewTickerManager()
	var t = T{i: 0}
	// 注册多个任务，每个任务有不同的执行间隔和参数
	tm.RegisterTask(1*time.Second, t.task1, "Hello", 42)
	tm.RegisterTask(2*time.Second, task2, 100, "World", 3.14)

	// 启动定时器
	tm.Start()

	// 模拟运行一段时间后停止定时器
	time.Sleep(15 * time.Second)
	tm.Stop()

	// 可以再次启动定时器
	time.Sleep(2 * time.Second)
	tm.Start()

	// 再次模拟运行一段时间后停止定时器
	time.Sleep(15 * time.Second)
	tm.Stop()
}
```
