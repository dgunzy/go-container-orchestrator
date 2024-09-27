package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	out io.Writer
	mu  sync.Mutex
}

func New(out io.Writer) *Logger {
	return &Logger{
		out: out,
	}
}

// SetOutput sets the output destination for the logger
func (l *Logger) SetOutput(out io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = out
}

func (l *Logger) log(level LogLevel, format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Get caller information
	_, file, line, ok := runtime.Caller(2) // Skip two frames to get the caller of Debug/Info/Warn/Error
	if !ok {
		file = "???"
		line = 0
	}

	// Use short file name
	file = filepath.Base(file)

	// Format the log message
	levelStr := "INFO"
	switch level {
	case DEBUG:
		levelStr = "DEBUG"
	case WARN:
		levelStr = "WARN"
	case ERROR:
		levelStr = "ERROR"
	}

	msg := fmt.Sprintf(format, v...)
	logLine := fmt.Sprintf("%s: %s:%d %s\n", levelStr, file, line, msg)

	// Write to output
	_, _ = l.out.Write([]byte(logLine))
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(DEBUG, format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.log(INFO, format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	l.log(WARN, format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.log(ERROR, format, v...)
}

// Global logger instance
var globalLogger *Logger

// init initializes the global logger
func init() {
	globalLogger = New(os.Stdout)
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	return globalLogger
}
