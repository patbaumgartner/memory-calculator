package logger

import (
	"log"
	"os"
)

// Logger provides a simple logging interface to replace bard.Logger
type Logger struct {
	logger *log.Logger
	quiet  bool
}

// Create creates a new logger instance
func Create(quiet bool) *Logger {
	return &Logger{
		logger: log.New(os.Stderr, "", log.LstdFlags),
		quiet:  quiet,
	}
}

// Info logs an informational message
func (l *Logger) Info(v ...interface{}) {
	if !l.quiet {
		l.logger.Print(v...)
	}
}

// Infof logs a formatted informational message
func (l *Logger) Infof(format string, v ...interface{}) {
	if !l.quiet {
		l.logger.Printf(format, v...)
	}
}

// Debug logs a debug message (currently same as Info)
func (l *Logger) Debug(v ...interface{}) {
	if !l.quiet {
		l.logger.Print(v...)
	}
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, v ...interface{}) {
	if !l.quiet {
		l.logger.Printf(format, v...)
	}
}
