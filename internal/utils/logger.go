package utils

import (
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	once   sync.Once
)

// LogLevel represents the severity level of logs
type LogLevel string

const (
	// DebugLevel logs messages for debugging
	DebugLevel LogLevel = "debug"
	// InfoLevel logs informational messages
	InfoLevel LogLevel = "info"
	// WarnLevel logs warning messages
	WarnLevel LogLevel = "warn"
	// ErrorLevel logs error messages
	ErrorLevel LogLevel = "error"
)

// GetLogger returns a singleton zap logger instance
func GetLogger() *zap.Logger {
	once.Do(func() {
		logger = initLogger(InfoLevel)
	})
	return logger
}

// GetLoggerWithLevel returns a logger with the specified log level
func GetLoggerWithLevel(level LogLevel) *zap.Logger {
	return initLogger(level)
}

// initLogger initializes a new zap logger
func initLogger(level LogLevel) *zap.Logger {
	// Create log directory if it doesn't exist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fall back to console-only logging
		return createConsoleLogger(level)
	}

	logDir := filepath.Join(homeDir, ".jenkins-tui", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return createConsoleLogger(level)
	}

	// Configure logging
	logFile := filepath.Join(logDir, "jenkins-tui.log")

	// Create encoders for console and file logging
	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	fileEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	// Open log file
	logFileWriter, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return createConsoleLogger(level)
	}

	// Convert LogLevel to zapcore.Level
	var zapLevel zapcore.Level
	switch level {
	case DebugLevel:
		zapLevel = zapcore.DebugLevel
	case InfoLevel:
		zapLevel = zapcore.InfoLevel
	case WarnLevel:
		zapLevel = zapcore.WarnLevel
	case ErrorLevel:
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Create core for both console and file
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logFileWriter), zapLevel),
	)

	// Create logger
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// createConsoleLogger creates a logger that only logs to the console
func createConsoleLogger(level LogLevel) *zap.Logger {
	var zapLevel zapcore.Level
	switch level {
	case DebugLevel:
		zapLevel = zapcore.DebugLevel
	case InfoLevel:
		zapLevel = zapcore.InfoLevel
	case WarnLevel:
		zapLevel = zapcore.WarnLevel
	case ErrorLevel:
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	logger, _ := config.Build()
	return logger
}

// Sugar returns a sugared logger for more convenient logging
func Sugar() *zap.SugaredLogger {
	return GetLogger().Sugar()
}
