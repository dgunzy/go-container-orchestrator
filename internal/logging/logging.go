package logging

import (
	"io"
	"log"
	"os"
	"sync"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger is our custom logger type
type Logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	mu          sync.Mutex
}

// New creates a new Logger instance
func New(out io.Writer) *Logger {
	return &Logger{
		debugLogger: log.New(out, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		infoLogger:  log.New(out, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLogger:  log.New(out, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(out, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// SetOutput sets the output destination for the logger
func (l *Logger) SetOutput(out io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debugLogger.SetOutput(out)
	l.infoLogger.SetOutput(out)
	l.warnLogger.SetOutput(out)
	l.errorLogger.SetOutput(out)
}

// Log logs a message with the specified level
func (l *Logger) Log(level LogLevel, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	switch level {
	case DEBUG:
		l.debugLogger.Print(v...)
	case INFO:
		l.infoLogger.Print(v...)
	case WARN:
		l.warnLogger.Print(v...)
	case ERROR:
		l.errorLogger.Print(v...)
	}
}

// Logf logs a formatted message with the specified level
func (l *Logger) Logf(level LogLevel, format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	switch level {
	case DEBUG:
		l.debugLogger.Printf(format, v...)
	case INFO:
		l.infoLogger.Printf(format, v...)
	case WARN:
		l.warnLogger.Printf(format, v...)
	case ERROR:
		l.errorLogger.Printf(format, v...)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(v ...interface{}) {
	l.Log(DEBUG, v...)
}

// Info logs an info message
func (l *Logger) Info(v ...interface{}) {
	l.Log(INFO, v...)
}

// Warn logs a warning message
func (l *Logger) Warn(v ...interface{}) {
	l.Log(WARN, v...)
}

// Error logs an error message
func (l *Logger) Error(v ...interface{}) {
	l.Log(ERROR, v...)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Logf(DEBUG, format, v...)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, v ...interface{}) {
	l.Logf(INFO, format, v...)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Logf(WARN, format, v...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Logf(ERROR, format, v...)
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
