package log

var (
	logger = &Logger{level: InfoLevel}
)

func SetLevel(level LogLevel) {
	logger.level = level
}

func Debug(args ...any) {
	logger.Debug(args...)
}

func Info(args ...any) {
	logger.Info(args...)
}

func Warn(args ...any) {
	logger.Warn(args...)
}

func Error(args ...any) {
	logger.Error(args...)
}

func Debugf(fmt string, args ...any) {
	logger.Debugf(fmt, args...)
}

func Infof(fmt string, args ...any) {
	logger.Infof(fmt, args...)
}

func Warnf(fmt string, args ...any) {
	logger.Warnf(fmt, args...)
}

func Errorf(fmt string, args ...any) {
	logger.Errorf(fmt, args...)
}
