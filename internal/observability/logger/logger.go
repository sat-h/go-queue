package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Global logger instance
	log  *zap.Logger
	once sync.Once
)

// Init initializes the global logger
// It should be called early in your application startup
func Init() {
	once.Do(func() {
		// Determine log level based on environment (default to info)
		logLevel := zap.InfoLevel
		if os.Getenv("DEBUG") == "true" {
			logLevel = zap.DebugLevel
		}

		// Configure logger
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		config := zap.Config{
			Level:       zap.NewAtomicLevelAt(logLevel),
			Development: false,
			Sampling: &zap.SamplingConfig{
				Initial:    100,
				Thereafter: 100,
			},
			Encoding:         "json",
			EncoderConfig:    encoderConfig,
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}

		var err error
		log, err = config.Build(zap.AddCallerSkip(1))
		if err != nil {
			panic("failed to initialize logger: " + err.Error())
		}

		log.Info("Logger initialized")
	})
}

// Get returns the global logger instance
func Get() *zap.Logger {
	if log == nil {
		// If logger hasn't been initialized, initialize with defaults
		Init()
	}
	return log
}

// Info logs a message at InfoLevel
func Info(message string, fields ...zap.Field) {
	Get().Info(message, fields...)
}

// Debug logs a message at DebugLevel
func Debug(message string, fields ...zap.Field) {
	Get().Debug(message, fields...)
}

// Error logs a message at ErrorLevel
func Error(message string, fields ...zap.Field) {
	Get().Error(message, fields...)
}

// Warn logs a message at WarnLevel
func Warn(message string, fields ...zap.Field) {
	Get().Warn(message, fields...)
}

// Fatal logs a message at FatalLevel
func Fatal(message string, fields ...zap.Field) {
	Get().Fatal(message, fields...)
}

// Sync flushes any buffered log entries
// Important to call this before application exit
func Sync() error {
	if log != nil {
		return log.Sync()
	}
	return nil
}
