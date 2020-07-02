package main

import (
	"net/http"
	"utils/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	test()
	//fmt.Println(files.GetFilePathInfo("./test/audit.log"))
}

func test() {
	//级别可设置：debug|info|warn|error
	//logPath可设置相对路径也可设置绝对路径
	logPath := "/root/go/src/utils/test.log"
	writer, _ := logger.GenWriter(logPath, 30, 24)
	logger.New(logger.GetLevel("debug"), writer)

	r := gin.New()
	r.GET("/", func(context *gin.Context) {
		printLog()
		context.JSON(http.StatusOK, "ok")
	})
	for i := 0; i < 1000000; i++ {
		printLog()
	}
	//logger.Panic("Panic")
	//logger.Panicf("%sf", "Panic")
	//logger.Fatal("fatal")         //开启后会自动退出
	//logger.Fatalf("%sf", "fatal") //上一条已退出 本条不能执行
	_ = r.Run(":6666")
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
