package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
}

type Fields logrus.Fields

/* Size分割方案
logPath:        日志文件路径
maxSize:        每个日志文件保存的最大尺寸,单位：M
maxAge:         文件最多保存时间,单位：天
*/
func SizeWriter(logPath string, maxSize, maxAge int) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   logPath, // 日志文件路径
		MaxSize:    maxSize, // 每个日志文件保存的最大尺寸,单位：M
		MaxBackups: 0,       // 日志文件最多保存多少个备份,0为不判断文件数量
		MaxAge:     maxAge,  // 文件最多保存多少天
		Compress:   true,    // 是否压缩,文本归档压缩率非常高
		LocalTime:  true,    // 取本地时区,一般需要开启
	}
}

/* Date分割方案
logPath:        日志文件路径
maxRetainDay:   文件最大保存时间,单位:天
splitTime:      日志切割时间,单位:小时
*/
func DateWriter(logPath string, maxRetainDay, splitTime time.Duration) (*rotatelogs.RotateLogs, error) {
	return rotatelogs.New(
		strings.Join([]string{logPath, "%Y%m%d%H%M%S"}, "."),
		rotatelogs.WithLinkName(logPath),                 // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(24*maxRetainDay*time.Hour), // 文件最大保存时间
		rotatelogs.WithRotationTime(splitTime*time.Hour), // 日志切割时间
		rotatelogs.WithClock(rotatelogs.Local),           // 本地时间
	)
}

//自定义日志格式，可参考new方法
func NewCustom(custom func()) {
	custom()
}

//writer可以为date分割方案也可以为size分割方案,size方案提供归档压缩,压缩率高节省空间
func New(level logrus.Level, writer io.Writer, isConsolePrint bool) {
	if isConsolePrint {
		lfHook := lfshook.NewHook(lfshook.WriterMap{
			logrus.DebugLevel: os.Stdout, // 为不同级别设置不同的输出目的,现在控制台输出
			logrus.InfoLevel:  os.Stdout,
			logrus.WarnLevel:  os.Stdout,
			logrus.ErrorLevel: os.Stdout,
			logrus.FatalLevel: os.Stdout,
			logrus.PanicLevel: os.Stdout,
		}, &Formatter{ //用上边的格式设置会失效
			HideKeys:      false,
			NoColors:      false,
			ShowFullLevel: true,
		})
		logrus.AddHook(lfHook)
	} else {
		logrus.AddHook(&NoConsolePrint{})
	}

	logrus.SetLevel(level)
	logrus.SetFormatter(&Formatter{ //写log日志不需要颜色
		HideKeys:      false,
		NoColors:      true,
		ShowFullLevel: false,
	})

	logrus.SetOutput(writer)
}

type NoConsolePrint struct{}

func (hook *NoConsolePrint) Fire(entry *logrus.Entry) error {
	return nil
}

func (hook *NoConsolePrint) Levels() []logrus.Level {
	return logrus.AllLevels
}

func output(logFunc func(...interface{}), level logrus.Level, format string, a ...interface{}) {
	message := fmt.Sprint(a...)
	if format != "" {
		message = fmt.Sprintf(format, a...)
	}

	pc, _, line, _ := runtime.Caller(2) //放在底层层极过多，拿不到最上层的函数，放到这层正常
	if level <= logrus.WarnLevel {      //warn以上级别打印错误位置和行号
		message = fmt.Sprintf("<%v#%v> %s",
			runtime.FuncForPC(pc).Name(), line, message)
	}
	logFunc(message)
}

/*
仅保留此种方式，适用于全局的log初始化
logger.new(...)
logger.Debug(...)
*/

func Debug(a ...interface{}) {
	output(logrus.Debug, logrus.DebugLevel, "", a...)
}

func Debugf(format string, a ...interface{}) {
	output(logrus.Debug, logrus.DebugLevel, format, a...)
}

func Info(a ...interface{}) {
	output(logrus.Info, logrus.InfoLevel, "", a...)
}

func Infof(format string, a ...interface{}) {
	output(logrus.Info, logrus.InfoLevel, format, a...)
}

func Warn(a ...interface{}) {
	output(logrus.Warn, logrus.WarnLevel, "", a...)
}

func Warnf(format string, a ...interface{}) {
	output(logrus.Warn, logrus.WarnLevel, format, a...)
}

func Error(a ...interface{}) {
	output(logrus.Error, logrus.ErrorLevel, "", a...)
}

func Errorf(format string, a ...interface{}) {
	output(logrus.Error, logrus.ErrorLevel, format, a...)
}

func Fatal(a ...interface{}) {
	output(logrus.Fatal, logrus.FatalLevel, "", a...)
}

func Fatalf(format string, a ...interface{}) {
	output(logrus.Fatal, logrus.FatalLevel, format, a...)
}

func Panic(a ...interface{}) {
	output(logrus.Panic, logrus.PanicLevel, "", a...)
}

func Panicf(format string, a ...interface{}) {
	output(logrus.Panic, logrus.PanicLevel, format, a...)
}

func Trace(a ...interface{}) {
	output(logrus.Trace, logrus.TraceLevel, "", a...)
}

func Tracef(format string, a ...interface{}) {
	output(logrus.Trace, logrus.TraceLevel, format, a...)
}

func WithField(key string, value interface{}) *logrus.Entry {
	return logrus.WithField(key, value)
}

func WithFields(fields Fields) *logrus.Entry {
	return logrus.WithFields(logrus.Fields(fields))
}

// Formatter - logrus formatter, implements logrus.Formatter
type Formatter struct {
	// FieldsOrder - default: fields sorted alphabetically
	FieldsOrder []string

	// TimestampFormat - default: time.StampMilli = "2006-01-02 15:04:05.000"
	TimestampFormat string

	// HideKeys - show [fieldValue] instead of [fieldKey:fieldValue]
	HideKeys bool

	// NoColors - disable colors
	NoColors bool

	// NoFieldsColors - apply colors only to the level, default is level + fields
	NoFieldsColors bool

	// ShowFullLevel - show a full level [WARNING] instead of [WARN]
	ShowFullLevel bool

	// TrimMessages - trim whitespaces on messages
	TrimMessages bool

	// CallerFirst - print caller info first
	CallerFirst bool

	// CustomCallerFormatter - set custom formatter for caller info
	CustomCallerFormatter func(*runtime.Frame) string
}

// Format an log entry
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	levelColor := getColorByLevel(entry.Level)

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = "2006-01-02 15:04:05.000"
	}

	// output buffer
	b := &bytes.Buffer{}

	// write time
	b.WriteString(entry.Time.Format(timestampFormat))

	// write level
	level := strings.ToUpper(entry.Level.String())

	if f.CallerFirst {
		f.writeCaller(b, entry)
	}

	if !f.NoColors {
		_, _ = fmt.Fprintf(b, "\x1b[%dm", levelColor)
	}

	b.WriteString(" [")
	if f.ShowFullLevel {
		b.WriteString(level)
	} else {
		b.WriteString(level[:4])
	}
	b.WriteString("] ")

	if !f.NoColors && f.NoFieldsColors {
		b.WriteString("\x1b[0m")
	}

	// write fields
	if f.FieldsOrder == nil {
		f.writeFields(b, entry)
	} else {
		f.writeOrderedFields(b, entry)
	}

	if !f.NoColors && !f.NoFieldsColors {
		b.WriteString("\x1b[0m")
	}

	// write message
	if f.TrimMessages {
		b.WriteString(strings.TrimSpace(entry.Message))
	} else {
		b.WriteString(entry.Message)
	}

	if !f.CallerFirst {
		f.writeCaller(b, entry)
	}

	b.WriteByte('\n')

	return b.Bytes(), nil
}

func (f *Formatter) writeCaller(b *bytes.Buffer, entry *logrus.Entry) {
	if entry.HasCaller() {
		if f.CustomCallerFormatter != nil {
			_, _ = fmt.Fprintf(b, f.CustomCallerFormatter(entry.Caller))
		} else {
			_, _ = fmt.Fprintf(
				b,
				" (%s:%d %s)",
				entry.Caller.File,
				entry.Caller.Line,
				entry.Caller.Function,
			)
		}
	}
}

func (f *Formatter) writeFields(b *bytes.Buffer, entry *logrus.Entry) {
	if len(entry.Data) != 0 {
		fields := make([]string, 0, len(entry.Data))
		for field := range entry.Data {
			fields = append(fields, field)
		}

		sort.Strings(fields)

		for _, field := range fields {
			f.writeField(b, entry, field)
		}
	}
}

func (f *Formatter) writeOrderedFields(b *bytes.Buffer, entry *logrus.Entry) {
	length := len(entry.Data)
	foundFieldsMap := map[string]bool{}
	for _, field := range f.FieldsOrder {
		if _, ok := entry.Data[field]; ok {
			foundFieldsMap[field] = true
			length--
			f.writeField(b, entry, field)
		}
	}

	if length > 0 {
		notFoundFields := make([]string, 0, length)
		for field := range entry.Data {
			if foundFieldsMap[field] == false {
				notFoundFields = append(notFoundFields, field)
			}
		}

		sort.Strings(notFoundFields)
		for _, field := range notFoundFields {
			f.writeField(b, entry, field)
		}
	}
}

func (f *Formatter) writeField(b *bytes.Buffer, entry *logrus.Entry, field string) {
	if f.HideKeys {
		_, _ = fmt.Fprintf(b, "[%v] ", entry.Data[field])
	} else {
		_, _ = fmt.Fprintf(b, "[%s:%v] ", field, entry.Data[field])
	}
}

const (
	red        = 31
	lightgreen = 32
	yellow     = 33
	blue       = 34
	purple     = 35
	green      = 36
	gray       = 37
)

func getColorByLevel(level logrus.Level) int {
	switch level {
	case logrus.InfoLevel:
		return green
	case logrus.DebugLevel:
		return blue
	case logrus.WarnLevel:
		return yellow
	case logrus.ErrorLevel:
		return red
	case logrus.FatalLevel, logrus.PanicLevel:
		return purple
	default:
		return gray
	}
}

func GetLevel(logLevel string) (level logrus.Level) {
	switch strings.ToLower(logLevel) {
	case "debug":
		level = logrus.DebugLevel
	case "info":
		level = logrus.InfoLevel
	case "warn":
		level = logrus.WarnLevel
	case "error":
		level = logrus.ErrorLevel
	default:
		level = logrus.ErrorLevel
	}

	return level
}
