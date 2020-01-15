package log

import (
	"errors"
	"fmt"
	utils "github.com/mzky/utils/files"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// levels
const (
	debugLevel = 0
	infoLevel  = 1
	warnLevel  = 2
	errorLevel = 3
	fatalLevel = 4

	printDebugLevel = "[DEBUG]"
	printInfoLevel  = "[INFO ]"
	printWarnLevel  = "[WARN ]"
	printErrorLevel = "[ERROR]"
	printFatalLevel = "[FATAL]" //此级别打印日志后自动退出

	logMaxLiveAge  = 24 * 60 * time.Hour //日志保留时间,单位小时,默认60天
	logSplitTime   = 24 * time.Hour      //日志分割时间,单位小时,默认每天分割
	logMaxFileSize = 1024 * 1024 * 10    //日志分割大小,最大10M,优先触发
)

type Logger struct {
	level      int
	baseLogger *log.Logger
	fileLogger *log.Logger
	baseFile   *os.File
	count      int32
}

var onDebugEvent func(logLevel int, errStr string)
var srvName = "sentry"      //服务名称
var logDir = ""             //log目录
var logLevel = "debug"      //默认日志级别
var logFullPath = ""        //带后缀.log全路径地址
var logFlag = log.LstdFlags //默认值3为彩色输出
var lastTimeLog = time.Now()

func New(serviceName string, strLevel string, logPath string) (*Logger, error) {
	srvName = serviceName
	logLevel = strLevel
	logDir = logPath
	// level
	var level int
	switch strings.ToLower(logLevel) {
	case "debug":
		level = debugLevel
	case "info":
		level = infoLevel
	case "warn":
		level = warnLevel
	case "error":
		level = errorLevel
	case "fatal":
		level = fatalLevel
	default:
		return nil, errors.New("日志级别设置错误: " + logLevel)
	}

	// logger
	var baseLogger *log.Logger
	var baseFile *os.File

	if logPath != "" {
		logFullPath = path.Join(logDir, srvName+".log")
		_, err := os.Stat(logFullPath)
		if !os.IsNotExist(err) { //已存在日志文件（每次执行重新创建日志,已有日志文件改名）
			err := os.Rename(logFullPath, path.Join(logDir, newLogName()))
			if err != nil {
				return nil, errors.New("检查文件或目录权限：" + logFullPath)
			}
		}

		file, err := os.Create(logFullPath)
		if err != nil {
			return nil, errors.New("检查文件或目录权限：" + logFullPath)
		}

		baseLogger = log.New(file, "", logFlag)
		baseFile = file
	}

	// new
	logger := new(Logger)
	logger.level = level
	logger.fileLogger = baseLogger
	logger.baseLogger = log.New(os.Stdout, "", logFlag)
	logger.baseFile = baseFile

	return logger, nil
}

// It's dangerous to call the method on logging
func (logger *Logger) Close() {
	if logger.baseFile != nil {
		logger.baseFile.Close()
	}
	logger.fileLogger = nil
	logger.baseLogger = nil
	logger.baseFile = nil
}

func (logger *Logger) doPrintf(color func(str string, modifier ...interface{}) string, level int, printLevel string, format string, a ...interface{}) {
	if level < logger.level {
		return
	}
	if logger.baseLogger == nil {
		panic("logger closed")
	}
	//定位执行函数及行号
	pc, _, line, _ := runtime.Caller(2)
	if level > infoLevel {
		//logger.baseLogger.SetFlags(log.Lshortfile | log.LstdFlags)
		format = fmt.Sprintf("%s <%v#%v> %s", printLevel, runtime.FuncForPC(pc).Name(), line, format)
	} else {
		//logger.baseLogger.SetFlags(log.LstdFlags)
		format = fmt.Sprintf("%s %s", printLevel, format)
	}

	str := fmt.Sprintf(format, a...)
	if onDebugEvent != nil {
		onDebugEvent(level, str)
	}

	// 控制台打印
	_ = logger.baseLogger.Output(log.LstdFlags, color(str))
	if logger.fileLogger != nil {
		//写文件
		_, err := os.Stat(logFullPath)
		if os.IsNotExist(err) { //不存在日志文件
			logger = createNewLog(logger, false)
		} else {
			//文件大小超过预设时,将当前日志改名并创建新日志
			if getFileSize(logFullPath) > int64(logMaxFileSize-len(str)) {
				logger = createNewLog(logger, true)
			}
		}
		now := time.Now()
		if now.Unix() >= lastTimeLog.Add(logSplitTime).Unix() { //每天分割日志
			logger = createNewLog(logger, true)
			lastTimeLog = now
		}

		_ = logger.fileLogger.Output(logFlag, str)
	}

	if level == fatalLevel {
		os.Exit(1)
	}
}

// 两种实现方式,此为第一种方式:此种方式适用于功能模块的log初始化
// logger, err := log.new(...)
// if err != nil {
//     log.Fatal("%s", err)
// }
// logger.debug(...)
func (logger *Logger) Debug(format string, a ...interface{}) {
	logger.doPrintf(green, debugLevel, printDebugLevel, format, a...)
}

func (logger *Logger) Info(format string, a ...interface{}) {
	logger.doPrintf(blue, infoLevel, printInfoLevel, format, a...)
}

func (logger *Logger) Warn(format string, a ...interface{}) {
	logger.doPrintf(yellow, warnLevel, printWarnLevel, format, a...)
}

func (logger *Logger) Error(format string, a ...interface{}) {
	logger.doPrintf(red, errorLevel, printErrorLevel, format, a...)
}

func (logger *Logger) Fatal(format string, a ...interface{}) {
	logger.doPrintf(cyan, fatalLevel, printFatalLevel, format, a...)
}

//第二种方式的默认级别
var gLogger, _ = New(srvName, "debug", "")

// It's dangerous to call the method on logging
func Export(logger *Logger) {
	if logger != nil {
		gLogger = logger
	}
}

// 两种实现方式,此为第二种方式:此种方式适用于全局的log初始化
// logger, err := log.new(...)
// if err != nil {
//     log.Fatal("%s", err)
// }
// log.Export(logger)
// log.Debug(...)
func Debug(format string, a ...interface{}) {
	gLogger.doPrintf(green, debugLevel, printDebugLevel, format, a...)
}

func Info(format string, a ...interface{}) {
	gLogger.doPrintf(blue, infoLevel, printInfoLevel, format, a...)
}

func Warn(format string, a ...interface{}) {
	gLogger.doPrintf(yellow, warnLevel, printWarnLevel, format, a...)
}

func Error(format string, a ...interface{}) {
	gLogger.doPrintf(red, errorLevel, printErrorLevel, format, a...)
}

func Fatal(format string, a ...interface{}) {
	gLogger.doPrintf(cyan, fatalLevel, printFatalLevel, format, a...)
}

//捕获错误行
func Recover(r interface{}) {
	buf := make([]byte, 4096)
	l := runtime.Stack(buf, false)
	Error("%+v: \r\n%s", r, string(buf[:l]))
}

func ReceiveMsg(format string, a ...interface{}) {
	gLogger.doPrintf(blue, debugLevel, printDebugLevel, format, a...)
}

func SendMsg(format string, a ...interface{}) {
	gLogger.doPrintf(purple, debugLevel, printDebugLevel, format, a...)
}

func FightValueChange(format string, a ...interface{}) {
	gLogger.doPrintf(green, debugLevel, printDebugLevel, format, a...)
}

func Close() {
	gLogger.Close()
}

//绿色字体，modifier里，第一个控制闪烁，第二个控制下划线
func green(str string, modifier ...interface{}) string {
	return cliColorRender(str, 32, 1, modifier...)
}

//青色/蓝绿色
func cyan(str string, modifier ...interface{}) string {
	return cliColorRender(str, 36, 1, modifier...)
}

//红字体
func red(str string, modifier ...interface{}) string {
	return cliColorRender(str, 31, 1, modifier...)
}

//黄色字体
func yellow(str string, modifier ...interface{}) string {
	return cliColorRender(str, 33, 1, modifier...)
}

//黑色
func black(str string, modifier ...interface{}) string {
	return cliColorRender(str, 30, 1, modifier...)
}

//深灰色
func darkGray(str string, modifier ...interface{}) string {
	return cliColorRender(str, 30, 1, modifier...)
}

//浅灰色
func lightGray(str string, modifier ...interface{}) string {
	return cliColorRender(str, 37, 1, modifier...)
}

//白色
func white(str string, modifier ...interface{}) string {
	return cliColorRender(str, 37, 1, modifier...)
}

//蓝色
func blue(str string, modifier ...interface{}) string {
	return cliColorRender(str, 34, 1, modifier...)
}

//紫色
func purple(str string, modifier ...interface{}) string {
	return cliColorRender(str, 35, 1, modifier...)
}

//棕色
func brown(str string, modifier ...interface{}) string {
	return cliColorRender(str, 33, 1, modifier...)
}

func cliColorRender(str string, color int, weight int, extraArgs ...interface{}) string {
	var mo []string

	if weight > 0 {
		mo = append(mo, fmt.Sprintf("%d", weight))
	}
	if len(mo) <= 0 {
		mo = append(mo, "0")
	}
	return "\033[" + strings.Join(mo, ";") + ";" + strconv.Itoa(color) + "m" + str + "\033[0m"
}

func getFileSize(filename string) int64 {
	var result int64
	filepath.Walk(filename, func(path string, f os.FileInfo, err error) error {
		result = f.Size()
		return nil
	})
	return result
}

func newLogName() string {
	now := time.Now()
	return fmt.Sprintf("%v%d%02d%02d_%02d_%02d_%02d.log",
		srvName,
		now.Year(),
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second())
}

func timeContrast() {
	now := time.Now()
	fileArr, _ := utils.GetAllFiles(logDir, ".log")
	for _, fi := range fileArr {
		if now.Unix() > utils.GetFileModTime(fi).Add(logMaxLiveAge).Unix() {
			os.Remove(fi)
		}
	}
}

func createNewLog(logger *Logger, state bool) *Logger {
	if state {
		os.Rename(logFullPath, path.Join(logDir, newLogName()))
	}
	file, _ := os.Create(logFullPath)
	logger.fileLogger = log.New(file, "", logFlag)
	logger.baseFile = file
	go timeContrast() //清理过期日志
	return logger
}
