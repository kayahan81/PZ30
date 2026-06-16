package logger

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func NewLogger(serviceName string) *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "["+serviceName+"] ", log.LstdFlags|log.Lshortfile),
	}
}

func (l *Logger) Info(msg string) {
	l.Printf("[INFO] %s", msg)
}

func (l *Logger) Warn(msg string) {
	l.Printf("[WARN] %s", msg)
}

func (l *Logger) Error(msg string) {
	l.Printf("[ERROR] %s", msg)
}

func (l *Logger) Debug(msg string) {
	l.Printf("[DEBUG] %s", msg)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Printf("[INFO] "+format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Printf("[WARN] "+format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Printf("[ERROR] "+format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Printf("[DEBUG] "+format, args...)
}
