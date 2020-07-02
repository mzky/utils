package main

import (
	"utils/logger"
)

func main() {
	test()
	//fmt.Println(files.GetFilePathInfo("./test/audit.log"))
}

func test() {
	//级别可设置：debug|info|warn|error
	//logPath可设置相对路径也可设置绝对路径
	logPath := "/root/go/src/utils/test.log"
	writer, _ := logger.GenWriter(logPath, 1, 1)
	logger.New(logger.GetLevel("debug"), writer)

	//性能测试
	for i := 0; i < 100; i++ {
		printLog()
	}

	logger.Panic("Panic")
	logger.Panicf("%sf", "Panic")
	logger.Fatal("fatal")         //开启后会自动退出
	logger.Fatalf("%sf", "fatal") //上一条已退出 本条不能执行
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
}
