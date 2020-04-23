package main

import (
	"fmt"
	"utils/log"
)

func main() {
	//级别可设置：debug|info|warn|error|fatal
	//logPath可设置相对路径也可设置绝对路径
	logger, err := log.New("audit", "debug", "./test")
	if err != nil {
		fmt.Printf("%s", err)
	}
	//第一种方法，局部引用
	logger.Info("info")
	logger.Debug("debug")
	logger.Warn("warn")//warn级别以上会显示错误函数及所在行数
	logger.Error("error")
	
	//第二种方法，全局使用
	log.Export(logger)
	log.Info("info")
	log.Debug("debug")
	log.Warn("warn")
	log.Error("error")
	log.ReceiveMsg("ReceiveMsg")
	log.FightValueChange("FightValueChange")
	log.SendMsg("SendMsg")
	log.Recover("Recover")//recover级别会显示错误函数上下文
	log.Fatal("fatal")//fatal级别会强制退出程序
}
