package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	outputs        map[LogLevel]*lumberjack.Logger
	combinedOutput *lumberjack.Logger
	mu             sync.Mutex
}

var globalLogger *Logger

func Setup(logDir string) error {
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logger := &Logger{
		outputs: make(map[LogLevel]*lumberjack.Logger),
	}

	logLevels := []struct {
		level LogLevel
		name  string
	}{
		{DEBUG, "debug"},
		{INFO, "info"},
		{WARN, "warn"},
		{ERROR, "error"},
	}

	for _, ll := range logLevels {
		logger.outputs[ll.level] = &lumberjack.Logger{
			Filename:   filepath.Join(logDir, ll.name+".log"),
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     7, // days
		}
	}

	logger.combinedOutput = &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "combined.log"),
		MaxSize:    50, // megabytes
		MaxBackups: 5,
		MaxAge:     30, // days
	}

	globalLogger = logger

	// Start a goroutine to clean up old logs daily
	go func() {
		for {
			time.Sleep(24 * time.Hour)
			if err := cleanupOldLogs(logDir, 30*24*time.Hour); err != nil {
				fmt.Printf("Error cleaning up old logs: %v\n", err)
			}
		}
	}()

	return nil
}

func GetLogger() *Logger {
	return globalLogger
}

func (l *Logger) log(level LogLevel, format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, file, line, _ := runtime.Caller(2)
	file = filepath.Base(file)
	levelStr := "INFO"
	switch level {
	case DEBUG:
		levelStr = "DEBUG"
	case WARN:
		levelStr = "WARN"
	case ERROR:
		levelStr = "ERROR"
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, v...)
	logLine := fmt.Sprintf("%s %s: %s:%d %s\n", timestamp, levelStr, file, line, msg)

	if writer, ok := l.outputs[level]; ok {
		_, _ = writer.Write([]byte(logLine))
	}
	_, _ = l.combinedOutput.Write([]byte(logLine))
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

func cleanupOldLogs(logDir string, maxAge time.Duration) error {
	return filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && time.Since(info.ModTime()) > maxAge {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
		return nil
	})
}

// Close flushes and closes all log files
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var errs []error

	for _, output := range l.outputs {
		if err := output.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if err := l.combinedOutput.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing log files: %v", errs)
	}

	return nil
}

// CloseGlobalLogger closes the global logger
func CloseGlobalLogger() error {
	if globalLogger != nil {
		return globalLogger.Close()
	}
	return nil
}
