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
	IsRunning bool
	wg        sync.WaitGroup
	mu        sync.Mutex
	stop      bool
}

// NewTickerManager 创建一个新的 TickerManager 实例
func NewTickerManager() *TickerManager {
	return &TickerManager{
		IsRunning: false,
		stop:      false,
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
	if tm.IsRunning {
		tm.mu.Unlock()
		return
	}
	tm.IsRunning = true
	tm.stop = false
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
			if tm.stop {
				return
			}
			task.fn.Call(task.args)
		case <-time.After(time.Second): // 避免阻塞
			if tm.stop {
				return
			}
		}
	}
}

// Stop 停止所有定时任务
func (tm *TickerManager) Stop() {
	tm.mu.Lock()
	if !tm.IsRunning {
		tm.mu.Unlock()
		return
	}
	tm.stop = true
	tm.IsRunning = false
	tm.mu.Unlock()

	tm.wg.Wait()
}

// ClearTasks 清空所有任务
func (tm *TickerManager) ClearTasks() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.tasks = []Task{}
}
