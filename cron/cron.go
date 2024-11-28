package cron

import (
	"reflect"
	"sync"
	"time"
)

// Task 定义一个定时任务
type Task struct {
	interval time.Duration
	fn       reflect.Value
	args     []reflect.Value
}

// TickerManager 管理定时器的启动和停止
type TickerManager struct {
	tasks     []Task
	isRunning bool
	wg        sync.WaitGroup
	mu        sync.Mutex
}

// NewTickerManager 创建一个新的 TickerManager 实例
func NewTickerManager() *TickerManager {
	return &TickerManager{
		isRunning: false,
	}
}

// RegisterTask 注册一个定时任务，支持任意数量和类型的参数
func (tm *TickerManager) RegisterTask(interval time.Duration, fn interface{}, args ...interface{}) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.tasks = append(tm.tasks, Task{
		interval: interval,
		fn:       reflect.ValueOf(fn),
		args:     toReflectValues(args),
	})
}

// toReflectValues 将接口切片转换为反射值切片
func toReflectValues(args []interface{}) []reflect.Value {
	var result []reflect.Value
	for _, arg := range args {
		result = append(result, reflect.ValueOf(arg))
	}
	return result
}

// Start 启动所有定时任务
func (tm *TickerManager) Start() {
	tm.mu.Lock()
	if tm.isRunning {
		tm.mu.Unlock()
		return
	}
	tm.isRunning = true
	tm.mu.Unlock()

	for _, task := range tm.tasks {
		tm.wg.Add(1)
		go tm.startTask(task)
	}
}

// startTask 启动单个定时任务
func (tm *TickerManager) startTask(task Task) {
	defer tm.wg.Done()
	ticker := time.NewTicker(task.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			task.fn.Call(task.args)
		}
	}
}

// Stop 停止所有定时任务
func (tm *TickerManager) Stop() {
	tm.mu.Lock()
	if !tm.isRunning {
		tm.mu.Unlock()
		return
	}
	tm.isRunning = false
	tm.mu.Unlock()

	tm.wg.Wait()
}