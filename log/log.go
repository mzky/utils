package log

import (
	"fmt"
	"github.com/bingoohuang/golog"
)

// InitLog 先初始化配置，使用的时候直接使用logrus.x()
func InitLog(level string, logPath string) {
	layout := `%t{yyyy-MM-dd HH:mm:ss.SSS} [%-5l{length=5}] ☆ %msg ☆ %caller{skip=1} %fields%n`
	spec := fmt.Sprintf("level=%s,file=%s,maxSize=10M,maxAge=1095d,gzipAge=3d,stdout=true", level, logPath)
	golog.Setup(golog.Layout(layout), golog.Spec(spec))
}
