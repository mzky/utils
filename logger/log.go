package logger

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"strings"

	files "github.com/mzky/utils/files"
	"github.com/sirupsen/logrus"

	"os"
	"path"
)

type Logger struct {
}

var gLogger Logger

func New(level logrus.Level, logPath string) {
	dir, _ := path.Split(logPath)
	if !files.IsExist(dir) {
		_ = os.MkdirAll(dir, os.ModePerm)
	}

	if !files.IsExist(logPath) {
		logFile, _ := os.Create(logPath)
		logFile.Close()
	}

	src, _ := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

	//logrus.SetLevel(logrus.DebugLevel)
	logrus.SetLevel(level)
	logrus.SetFormatter(&Formatter{
		HideKeys:      true,
		ShowFullLevel: true,
		FieldsOrder:   []string{"component", "category"},
	})
	logrus.SetOutput(src)

}

func (temp *Logger) Print(execute func(...interface{}), level logrus.Level, format string, a ...interface{}) {
	formatMessage := fmt.Sprintf(format, a...)
	pc, _, line, _ := runtime.Caller(2)
	if level <= logrus.WarnLevel { //warn以上级别打印错误位置和行号
		formatMessage = fmt.Sprintf("<%v#%v> %v",
			runtime.FuncForPC(pc).Name(), line, formatMessage)
	}
	execute(formatMessage)
}

/*
仅保留此种方式，适用于全局的log初始化
logger.new(...)
logger.Debug(...)
*/

func Debug(a ...interface{}) {
	gLogger.Print(logrus.Debug, logrus.DebugLevel, "%v", a...)
}

func Debugf(format string, a ...interface{}) {
	gLogger.Print(logrus.Debug, logrus.DebugLevel, format, a...)
}

func Info(a ...interface{}) {
	gLogger.Print(logrus.Info, logrus.InfoLevel, "%v", a...)
}

func Infof(format string, a ...interface{}) {
	gLogger.Print(logrus.Info, logrus.InfoLevel, format, a...)
}

func Warn(a ...interface{}) {
	gLogger.Print(logrus.Warn, logrus.WarnLevel, "%v", a...)
}

func Warnf(format string, a ...interface{}) {
	gLogger.Print(logrus.Warn, logrus.WarnLevel, format, a...)
}

func Error(a ...interface{}) {
	gLogger.Print(logrus.Error, logrus.ErrorLevel, "%v", a...)
}

func Errorf(format string, a ...interface{}) {
	gLogger.Print(logrus.Error, logrus.ErrorLevel, format, a...)
}

func Fatal(a ...interface{}) {
	gLogger.Print(logrus.Fatal, logrus.FatalLevel, "%v", a...)
}

func Fatalf(format string, a ...interface{}) {
	gLogger.Print(logrus.Fatal, logrus.FatalLevel, format, a...)
}

func Panic(a ...interface{}) {
	gLogger.Print(logrus.Panic, logrus.PanicLevel, "%v", a...)
}

func Panicf(format string, a ...interface{}) {
	gLogger.Print(logrus.Panic, logrus.PanicLevel, format, a...)
}

// Formatter - logrus formatter, implements logrus.Formatter
type Formatter struct {
	// FieldsOrder - default: fields sorted alphabetically
	FieldsOrder []string

	// TimestampFormat - default: time.StampMilli = "Jan _2 15:04:05.000"
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
		timestampFormat = "2006-01-02_15:04:05.000"
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
		fmt.Fprintf(b, "\x1b[%dm", levelColor)
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
			fmt.Fprintf(b, f.CustomCallerFormatter(entry.Caller))
		} else {
			fmt.Fprintf(
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
		fmt.Fprintf(b, "[%v] ", entry.Data[field])
	} else {
		fmt.Fprintf(b, "[%s:%v] ", field, entry.Data[field])
	}
}

const (
	colorRed    = 31
	colorYellow = 33
	colorBlue   = 36
	colorGray   = 37
)

func getColorByLevel(level logrus.Level) int {
	switch level {
	case logrus.DebugLevel:
		return colorGray
	case logrus.WarnLevel:
		return colorYellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return colorRed
	default:
		return colorBlue
	}
}
