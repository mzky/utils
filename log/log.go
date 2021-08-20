package log

import (
	"github.com/bingoohuang/golog/pkg/logfmt"
	"github.com/bingoohuang/golog/pkg/spec"
	"github.com/bingoohuang/golog/pkg/timex"
)

func InitLog(level string, logPath string) {
	var size spec.Size
	_ = size.Parse("10M") // 最大单个日志文件10M
	layout := `%t{yyyy-MM-dd HH:mm:ss.SSS} [%-5l{length=5}] ☆ %msg ☆ %caller %fields %n`
	maxAge, _ := timex.ParseDuration("1095d") // 最大保留3年
	gzipAge, _ := timex.ParseDuration("3d")   // 归档压缩3天前的日志

	logs := logfmt.LogrusOption{
		Level:       level,
		LogPath:     logPath,
		Rotate:      ".yyyy-MM-dd",
		MaxAge:      maxAge,
		GzipAge:     gzipAge,
		MaxSize:     int64(size),
		PrintColor:  true,
		PrintCaller: true,
		Stdout:      true,
		Layout:      layout,
	}
	logs.Setup(nil)
}
