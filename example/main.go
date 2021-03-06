package main

import (
	"time"

	"github.com/mzky/utils/logger"
	"github.com/sirupsen/logrus"
)

func main() {
	//级别可设置：debug|info|warn|error
	//logPath可设置相对路径也可设置绝对路径

	custom()
	size()
	date()
	//test()

	logger.Panic("Panic")
	logger.Panicf("%sf", "Panic")
	logger.Fatal("fatal")         //开启后会自动退出
	logger.Fatalf("%sf", "fatal") //上一条已退出 本条不能执行
}

func custom() {
	logPath := "./customSplit.log"
	writer := logger.SizeWriter(logPath, 100, 730)
	logger.NewCustom(func() {
		logrus.SetLevel(logger.GetLevel("debug"))
		logrus.SetFormatter(&logger.Formatter{ //写log日志不需要颜色
			HideKeys:      true,
			NoColors:      true,
			ShowFullLevel: false,
		})
		logrus.SetOutput(writer)
	})
	printLog()
}

func size() {
	//文件大小分割，建议使用此方法，归档压缩率高，节省空间
	logPath := "./sizeSplit.log"
	writer := logger.SizeWriter(logPath, 100, 730)
	logger.New(logger.GetLevel("debug"), writer, true)
	printLog()
}

func date() {
	//时间分割方式，两种方式同时仅生效最后一个设置
	logPath := "./dateSplit.log"
	//文件名只能精确到小时，此问题为第三方组件限制
	writer, _ := logger.DateWriter(logPath, 1, 1)
	logger.New(logger.GetLevel("debug"), writer, true)
	printLog()
}

func test() {
	//性能测试
	for i := 0; i < 1000000; i++ {
		go printLog()                     //性能测试过程建议关闭控制台输出，避免内存占用过高卡死ide
		time.Sleep(time.Microsecond * 10) //循环里的协程必须加sleep,否则线程锁会导致不保存日志文件
	}
}

func printLog() {
	logger.Info("info")
	logger.Infof("%sf", "info")
	logger.Debug("debug")
	logger.Debugf("%sf", "debug")
	logger.Warn("warn") //warn级别以上会显示错误函数及所在行数
	logger.Warnf("%sf", "warn")
	logger.Error("error")
	logger.Errorf("%sf", "error")

	logger.WithField("key", "value").Info("WithField")
	logger.WithFields(logger.Fields{"key": "value"}).Warn("WithFields")
	logger.WithFields(logger.Fields{
		"component": "component_value",
		"category":  "category_value"}).Error("WithFields")
}
