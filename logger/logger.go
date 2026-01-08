package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the log level.
type Level int

const (
	// DebugLevel for debug messages
	DebugLevel Level = iota
	// InfoLevel for informational messages
	InfoLevel
	// WarnLevel for warning messages
	WarnLevel
	// ErrorLevel for error messages
	ErrorLevel
)

// String returns the string representation of the log level.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging functionality.
type Logger struct {
	level  Level
	output io.Writer
}

// New creates a new logger that writes to stdout.
func New() *Logger {
	return &Logger{
		level:  InfoLevel,
		output: os.Stdout,
	}
}

// NewWithConfig creates a logger with custom configuration.
func NewWithConfig(level Level, output io.Writer) *Logger {
	return &Logger{
		level:  level,
		output: output,
	}
}

// SetLevel sets the minimum log level.
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// Debug logs a debug message with optional fields.
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.log(DebugLevel, msg, fields...)
}

// Info logs an informational message with optional fields.
func (l *Logger) Info(msg string, fields ...interface{}) {
	l.log(InfoLevel, msg, fields...)
}

// Warn logs a warning message with optional fields.
func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.log(WarnLevel, msg, fields...)
}

// Error logs an error message with optional fields.
func (l *Logger) Error(msg string, fields ...interface{}) {
	l.log(ErrorLevel, msg, fields...)
}

// log is the internal logging method.
func (l *Logger) log(level Level, msg string, fields ...interface{}) {
	if level < l.level {
		return
	}

	// Build log entry
	entry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     level.String(),
		"message":   msg,
	}

	// Add fields as key-value pairs
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fmt.Sprintf("%v", fields[i])
			entry[key] = fields[i+1]
		}
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}

	// Write to output
	fmt.Fprintln(l.output, string(data))
}

// Default logger instance
var defaultLogger = New()

// Debug logs using the default logger.
func Debug(msg string, fields ...interface{}) {
	defaultLogger.Debug(msg, fields...)
}

// Info logs using the default logger.
func Info(msg string, fields ...interface{}) {
	defaultLogger.Info(msg, fields...)
}

// Warn logs using the default logger.
func Warn(msg string, fields ...interface{}) {
	defaultLogger.Warn(msg, fields...)
}

// Error logs using the default logger.
func Error(msg string, fields ...interface{}) {
	defaultLogger.Error(msg, fields...)
}

// SetLevel sets the level for the default logger.
func SetLevel(level Level) {
	defaultLogger.SetLevel(level)
}
