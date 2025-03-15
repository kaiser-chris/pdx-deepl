package logging

import (
	"log"
	"os"
)

const (
	LevelTrace = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelOff
)

const (
	PrefixTrace = AnsiFgBlue + "TRACE " + AnsiFgDefault
	PrefixDebug = AnsiFgBlue + "DEBUG " + AnsiFgDefault
	PrefixInfo  = AnsiFgGreen + "INFO " + AnsiFgDefault
	PrefixWarn  = AnsiFgYellow + "WARN " + AnsiFgDefault
	PrefixError = AnsiFgRed + "ERROR " + AnsiFgDefault
	PrefixFatal = AnsiFgRed + "FATAL " + AnsiFgDefault
)

var globalMinLogLevel = LevelInfo

func SetGlobalLogLevel(minLevel int) {
	globalMinLogLevel = minLevel
	GlobalLogger.MinLogLevel = minLevel
}

var GlobalLogger = New()

// Extremely simple logger
// that includes:
//   - log levels
//   - ansi colors
type Logger struct {
	MinLogLevel int
	*log.Logger
}

func New() *Logger {
	logger := Logger{
		MinLogLevel: globalMinLogLevel,
		Logger:      log.New(os.Stdout, PrefixInfo, log.Ldate|log.Ltime),
	}
	return &logger
}

func (logger *Logger) Trace(v ...any) {
	if logger.MinLogLevel > LevelTrace {
		return
	}
	logger.SetPrefix(PrefixTrace)
	logger.Println(v...)
}

func (logger *Logger) Tracef(format string, v ...any) {
	if logger.MinLogLevel > LevelTrace {
		return
	}
	logger.SetPrefix(PrefixTrace)
	logger.Printf(format, v...)
}

func (logger *Logger) Debug(v ...any) {
	if logger.MinLogLevel > LevelDebug {
		return
	}
	logger.SetPrefix(PrefixDebug)
	logger.Println(v...)
}

func (logger *Logger) Debugf(format string, v ...any) {
	if logger.MinLogLevel > LevelDebug {
		return
	}
	logger.SetPrefix(PrefixDebug)
	logger.Printf(format, v...)
}

func (logger *Logger) Info(v ...any) {
	if logger.MinLogLevel > LevelInfo {
		return
	}
	logger.SetPrefix(PrefixInfo)
	logger.Println(v...)
}

func (logger *Logger) Infof(format string, v ...any) {
	if logger.MinLogLevel > LevelInfo {
		return
	}
	logger.SetPrefix(PrefixInfo)
	logger.Printf(format, v...)
}

func (logger *Logger) Warn(v ...any) {
	if logger.MinLogLevel > LevelWarn {
		return
	}
	logger.SetPrefix(PrefixWarn)
	logger.Println(v...)
}

func (logger *Logger) Warnf(format string, v ...any) {
	if logger.MinLogLevel > LevelWarn {
		return
	}
	logger.SetPrefix(PrefixWarn)
	logger.Printf(format, v...)
}

func (logger *Logger) Error(v ...any) {
	if logger.MinLogLevel > LevelError {
		return
	}
	logger.SetPrefix(PrefixError)
	logger.Println(v...)
}

func (logger *Logger) Errorf(format string, v ...any) {
	if logger.MinLogLevel > LevelError {
		return
	}
	logger.SetPrefix(PrefixError)
	logger.Printf(format, v...)
}

func (logger *Logger) Fatal(v ...any) {
	logger.SetPrefix(PrefixFatal)
	logger.Logger.Fatal(v...)
}

func (logger *Logger) Fatalf(format string, v ...any) {
	logger.SetPrefix(PrefixFatal)
	logger.Logger.Fatalf(format, v...)
}

func Trace(v ...any) {
	GlobalLogger.Trace(v...)
}

func Tracef(format string, v ...any) {
	GlobalLogger.Tracef(format, v...)
}

func Debug(v ...any) {
	GlobalLogger.Debug(v...)
}

func Debugf(format string, v ...any) {
	GlobalLogger.Debugf(format, v...)
}

func Info(v ...any) {
	GlobalLogger.Info(v...)
}

func Infof(format string, v ...any) {
	GlobalLogger.Infof(format, v...)
}

func Warn(v ...any) {
	GlobalLogger.Warn(v...)
}

func Warnf(format string, v ...any) {
	GlobalLogger.Warnf(format, v...)
}

func Error(v ...any) {
	GlobalLogger.Error(v...)
}

func Errorf(format string, v ...any) {
	GlobalLogger.Errorf(format, v...)
}

func Fatal(v ...any) {
	GlobalLogger.Fatal(v...)
}

func Fatalf(format string, v ...any) {
	GlobalLogger.Fatalf(format, v...)
}
