package main

import (
	"github.com/mzky/utils/logger"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	test()
	//fmt.Println(files.GetFilePathInfo("./test/audit.log"))
}

func test() {
	//级别可设置：debug|info|warn|error
	//logPath可设置相对路径也可设置绝对路径

	//文件大小分割，建议使用此方法，归档压缩率高，节省空间
	logPath := "./sizeSplit.log"
	writer := &lumberjack.Logger{
		Filename:   logPath, // 日志文件路径
		MaxSize:    100,     // 每个日志文件保存的最大尺寸,单位：M
		MaxBackups: 0,       // 日志文件最多保存多少个备份,0为不判断文件数量
		MaxAge:     730,     // 文件最多保存多少天
		Compress:   true,    // 是否压缩,文本归档压缩率非常高
		LocalTime:  true,    // 取本地时区,一般需要开启
	}
	//isConsolePrint true时即写log文件也在控制台打印,并且控制台提供颜色输出
	logger.New(logger.GetLevel("debug"), writer, false)
	printLog()

	////时间分割方式，两种方式同时仅生效最后一个设置
	//logPath2 := "./dateSplit.log"
	////文件名只能精确到小时，分秒为0000，此问题待解
	//writerForDate, _ := logger.GenWriter(logPath2, 1, 1)
	//logger.NewForDate(logger.GetLevel("debug"), writerForDate, false)
	//printLog()

	////性能测试
	//for i := 0; i < 1000000; i++ {
	//	go printLog()                     //性能测试过程建议关闭控制台输出，避免内存占用过高卡死ide
	//	time.Sleep(time.Microsecond * 10) //循环里的协程必须加sleep,否则线程锁会导致不保存日志文件
	//}

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

	logger.WithField("key", "value").Info("WithField")
	logger.WithFields(logger.Fields{"key": "value"}).Warn("WithFields")
	logger.WithFields(logger.Fields{
		"component": "component_value",
		"category":  "category_value"}).Error("WithFields")

}
