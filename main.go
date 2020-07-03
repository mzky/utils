package main

import (
	"time"
	"utils/logger"

	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	test()
	//fmt.Println(files.GetFilePathInfo("./test/audit.log"))
}

func test() {
	//级别可设置：debug|info|warn|error
	//logPath可设置相对路径也可设置绝对路径
	logPath := "./test.log"
	writer := &lumberjack.Logger{
		Filename:   logPath, // 日志文件路径
		MaxSize:    100,     // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: 0,       // 日志文件最多保存多少个备份,0为不判断文件数量
		MaxAge:     730,     // 文件最多保存多少天
		Compress:   true,    // 是否压缩,文本归档压缩率非常高
		LocalTime:  true,    //取本地时区,一般需要开启
	}
	//isConsolePrint true时即写log文件也在控制台打印,并且控制台提供颜色输出
	logger.New(logger.GetLevel("debug"), writer, false)

	for i := 0; i < 1000000; i++ {
		go printLog()
		time.Sleep(time.Microsecond * 10) //循环里的协程必须加sleep,否则线程锁会导致不保存日志
	}
	//logger.Panic("Panic")
	//logger.Panicf("%sf", "Panic")
	//logger.Fatal("fatal")         //开启后会自动退出
	//logger.Fatalf("%sf", "fatal") //上一条已退出 本条不能执行
	//select {}
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
