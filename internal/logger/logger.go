// Package logger provides a simple logging interface.
package logger

import (
	"log"
	"os"
)

// Logger provides a simple logging interface to replace bard.Logger
//
// Everything is written to stderr, never stdout: --quiet exists so that
// `$(memory-calculator --quiet)` captures exactly the JVM options, and a diagnostic on stdout
// would be substituted into JAVA_TOOL_OPTIONS.
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

// Warnf reports a condition that changes the result, and is emitted even under --quiet.
//
// A warning means the options being printed are not the ones the caller asked for: the memory
// limit could not be found, it was clamped, or classes could not be counted. Suppressing that
// under --quiet would hand back a silently mis-sized heap, so only the informational log honors
// the flag.
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.logger.Printf("WARNING: "+format, v...)
}
