package log

import (
	"fmt"
	"log"
	"os"
	"sync"
)

const (
	ErrorLevel LogLevel = iota + 1
	WarnLevel
	InfoLevel
	DebugLevel
)

const (
	callDepth = 4
)

type ILogger interface {
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Debugf(fmt string, args ...any)
	Infof(fmt string, args ...any)
	Warnf(fmt string, args ...any)
	Errorf(fmt string, args ...any)
}

type LogLevel int

type Logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	level       LogLevel // 日志等级：xlog.DebugLevel、xlog.InfoLevel、xlog.WarnLevel、xlog.ErrorLevel
	once        sync.Once
}

func NewLogger() *Logger {
	return &Logger{level: InfoLevel}
}

func (l *Logger) Debug(args ...any) {
	l.logout(DebugLevel, nil, args...)
}

func (l *Logger) Info(args ...any) {
	l.logout(InfoLevel, nil, args...)
}

func (l *Logger) Warn(args ...any) {
	l.logout(WarnLevel, nil, args...)
}

func (l *Logger) Error(args ...any) {
	l.logout(ErrorLevel, nil, args...)
}

func (l *Logger) Debugf(fmt string, args ...any) {
	l.logout(DebugLevel, &fmt, args...)
}

func (l *Logger) Infof(fmt string, args ...any) {
	l.logout(InfoLevel, &fmt, args...)
}

func (l *Logger) Warnf(fmt string, args ...any) {
	l.logout(WarnLevel, &fmt, args...)
}

func (l *Logger) Errorf(fmt string, args ...any) {
	l.logout(ErrorLevel, &fmt, args...)
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *Logger) logout(level LogLevel, format *string, args ...any) {
	l.once.Do(func() {
		l.init()
	})
	var msg string
	if format != nil {
		msg = fmt.Sprintf(*format, args...)
	} else {
		msg = fmt.Sprintln(args...)
	}
	if l.level >= level {
		switch level {
		case ErrorLevel:
			l.errorLogger.Output(callDepth, msg)
		case WarnLevel:
			l.warnLogger.Output(callDepth, msg)
		case InfoLevel:
			l.infoLogger.Output(callDepth, msg)
		case DebugLevel:
			l.debugLogger.Output(callDepth, msg)
		}
	}
}

func (l *Logger) init() {
	l.debugLogger = log.New(os.Stdout, "[DEBUG] >> ", log.Lmsgprefix|log.Lshortfile|log.Ldate|log.Lmicroseconds)
	l.infoLogger = log.New(os.Stdout, "[INFO] >> ", log.Lmsgprefix|log.Lshortfile|log.Ldate|log.Lmicroseconds)
	l.warnLogger = log.New(os.Stdout, "[WARN] >> ", log.Lmsgprefix|log.Lshortfile|log.Ldate|log.Lmicroseconds)
	l.errorLogger = log.New(os.Stdout, "[ERROR] >> ", log.Lmsgprefix|log.Lshortfile|log.Ldate|log.Lmicroseconds)
}
