package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

const (
	ERROR = iota
	WARN
	INFO
	DEBUG
	SKIP
)

var (
	DefaultLogger = NewLogger(os.Stderr, "Debug")
)

type Logger struct {
	l     *log.Logger
	lock  sync.Mutex
	level int
}

func Lock() {
	DefaultLogger.lock.Lock()
}

func Unlock() {
	DefaultLogger.lock.Unlock()
}

func SetLevel(level string) {
	DefaultLogger.SetLevel(level)
}

func GetLevel() string {
	return DefaultLogger.GetLevel()
}

func Error(format string, v ...interface{}) {
	DefaultLogger.Error(format, v...)
}

func Info(format string, v ...interface{}) {
	DefaultLogger.Info(format, v...)
}

func Warn(format string, v ...interface{}) {
	DefaultLogger.Warn(format, v...)
}

func Debug(format string, v ...interface{}) {
	DefaultLogger.Debug(format, v...)
}

func Skip(format string, v ...interface{}) {
	DefaultLogger.Skip(format, v...)
}

func NewLogger(out io.Writer, level string) *Logger {
	logger := new(Logger)
	logger.l = log.New(out, "", log.Lmsgprefix)
	logger.SetLevel(level)
	return logger
}

func (logger *Logger) SetLevel(level string) {
	logger.lock.Lock()
	defer logger.lock.Unlock()
	switch strings.ToUpper(level) {
	case "ERROR":
		logger.level = ERROR
	case "WARN":
		logger.level = WARN
	case "INFO":
		logger.level = INFO
	case "DEBUG":
		logger.level = DEBUG
	case "SKIP":
		logger.level = SKIP
	}
}

func (logger *Logger) GetLevel() string {
	switch logger.level {
	case ERROR:
		return "ERROR"
	case WARN:
		return "WARN"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	case SKIP:
		return "SKIP"
	default:
		panic(fmt.Errorf("unknown log level"))
	}
}

func (logger *Logger) Error(format string, v ...interface{}) {
	logger.lock.Lock()
	defer logger.lock.Unlock()
	logger.l.Fatalf("\x1b[31m[x] %s\x1b[0m", fmt.Sprintf(format, v...))
}

func (logger *Logger) Info(format string, v ...interface{}) {
	logger.lock.Lock()
	defer logger.lock.Unlock()
	if logger.level >= INFO {
		logger.l.Printf("\x1b[36m[*] %s\x1b[0m", fmt.Sprintf(format, v...))
	}
}

func (logger *Logger) Warn(format string, v ...interface{}) {
	logger.lock.Lock()
	defer logger.lock.Unlock()
	if logger.level >= WARN {
		logger.l.Printf("\x1b[33m[!] %s\x1b[0m", fmt.Sprintf(format, v...))
	}
}

func (logger *Logger) Debug(format string, v ...interface{}) {
	logger.lock.Lock()
	defer logger.lock.Unlock()
	if logger.level >= DEBUG {
		logger.l.Printf("\x1b[95m[-] %s\x1b[0m", fmt.Sprintf(format, v...))
	}
}

func (logger *Logger) Skip(format string, v ...interface{}) {
	logger.lock.Lock()
	defer logger.lock.Unlock()
	if logger.level >= SKIP {
		logger.l.Printf("\x1b[90m[^] %s\x1b[0m", fmt.Sprintf(format, v...))
	}
}
