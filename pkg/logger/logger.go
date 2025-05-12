package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/swryu/telegpt/pkg/config"
)

// Level represents the severity level of a log message
type Level int

const (
	// LevelDebug for detailed diagnostic information
	LevelDebug Level = iota
	// LevelInfo for general operational information
	LevelInfo
	// LevelWarn for potentially harmful situations
	LevelWarn
	// LevelError for error events that might still allow the application to continue
	LevelError
	// LevelFatal for severe error events that will presumably lead the application to abort
	LevelFatal
)

var (
	logger *Logger
	once   sync.Once
)

// Logger is a simple logging utility
type Logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
	level       Level
	output      io.Writer
	fileHandle  *os.File
}

// Initialize initializes the logger with configuration
func Initialize(cfg *config.Config) error {
	// Parse log level
	var level Level
	switch strings.ToLower(cfg.Logging.Level) {
	case "debug":
		level = LevelDebug
	case "info":
		level = LevelInfo
	case "warn", "warning":
		level = LevelWarn
	case "error":
		level = LevelError
	case "fatal":
		level = LevelFatal
	default:
		level = LevelInfo
	}

	// Configure outputs
	var outputs []io.Writer

	// Add console output if enabled
	if cfg.Logging.Console {
		outputs = append(outputs, os.Stdout)
	}

	// Add file output if specified
	var fileHandle *os.File
	if cfg.Logging.File != "" {
		// Create directory if it doesn't exist
		dir := filepath.Dir(cfg.Logging.File)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		file, err := os.OpenFile(cfg.Logging.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		outputs = append(outputs, file)
		fileHandle = file
	}

	// Create multi-writer if we have multiple outputs
	var output io.Writer
	if len(outputs) > 0 {
		output = io.MultiWriter(outputs...)
	} else {
		// Default to stdout if no outputs configured
		output = os.Stdout
	}

	// Initialize the logger
	once.Do(func() {
		logger = &Logger{
			debugLogger: log.New(output, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
			infoLogger:  log.New(output, "INFO: ", log.Ldate|log.Ltime),
			warnLogger:  log.New(output, "WARN: ", log.Ldate|log.Ltime),
			errorLogger: log.New(output, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
			fatalLogger: log.New(output, "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile),
			level:       level,
			output:      output,
			fileHandle:  fileHandle,
		}
	})

	return nil
}

// Close closes any open file handles
func Close() {
	if logger != nil && logger.fileHandle != nil {
		logger.fileHandle.Close()
		logger.fileHandle = nil
	}
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	if logger == nil {
		return
	}
	if logger.level <= LevelDebug {
		logger.debugLogger.Printf(format, v...)
	}
}

// Info logs an informational message
func Info(format string, v ...interface{}) {
	if logger == nil {
		return
	}
	if logger.level <= LevelInfo {
		logger.infoLogger.Printf(format, v...)
	}
}

// Warn logs a warning message
func Warn(format string, v ...interface{}) {
	if logger == nil {
		return
	}
	if logger.level <= LevelWarn {
		logger.warnLogger.Printf(format, v...)
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	if logger == nil {
		return
	}
	if logger.level <= LevelError {
		logger.errorLogger.Printf(format, v...)
	}
}

// Fatal logs a fatal message and exits the application
func Fatal(format string, v ...interface{}) {
	if logger == nil {
		log.Fatalf(format, v...)
		return
	}
	if logger.level <= LevelFatal {
		logger.fatalLogger.Printf(format, v...)
		os.Exit(1)
	}
}
