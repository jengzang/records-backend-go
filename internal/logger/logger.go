package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// Init initializes the global logger with the specified log level
func Init(logLevel string) {
	Log = logrus.New()

	// Set output to stdout
	Log.SetOutput(os.Stdout)

	// Set JSON formatter for structured logging
	Log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     false,
	})

	// Parse and set log level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
		Log.Warnf("Invalid log level '%s', defaulting to 'info'", logLevel)
	}
	Log.SetLevel(level)

	Log.WithFields(logrus.Fields{
		"level": level.String(),
	}).Info("Logger initialized")
}

// Info logs an informational message with optional fields
func Info(msg string, fields logrus.Fields) {
	if fields == nil {
		Log.Info(msg)
	} else {
		Log.WithFields(fields).Info(msg)
	}
}

// Error logs an error message with optional fields
func Error(msg string, err error, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	Log.WithFields(fields).Error(msg)
}

// Warn logs a warning message with optional fields
func Warn(msg string, fields logrus.Fields) {
	if fields == nil {
		Log.Warn(msg)
	} else {
		Log.WithFields(fields).Warn(msg)
	}
}

// Debug logs a debug message with optional fields
func Debug(msg string, fields logrus.Fields) {
	if fields == nil {
		Log.Debug(msg)
	} else {
		Log.WithFields(fields).Debug(msg)
	}
}

// Fatal logs a fatal error and exits the program
func Fatal(msg string, err error, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	Log.WithFields(fields).Fatal(msg)
}
