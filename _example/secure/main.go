package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mzky/utils/secure"
	"time"
)

func main() {
	r := gin.Default()
	// 初始化缓慢攻击防御中间件
	defender := secure.NewSlowHTTPDefender(500, 10, 20, 1<<20, 10*time.Second)
	// 应用防御中间件
	r.Use(defender.DefenseMiddleware())
}
