package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config holds the logger configuration
type Config struct {
	Level      string
	TimeFormat string
	Pretty     bool
}

// Logger defines the interface for logging
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, err error)
	Fatal(msg string, err error)
	Warn(msg string, err error)
	Debug(msg string, args ...interface{})
}

// zerologLogger implements the Logger interface using zerolog
type zerologLogger struct {
	logger *zerolog.Logger
}

// New creates a new logger instance
func New(cfg Config) Logger {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Set time format
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}
	zerolog.TimeFieldFormat = cfg.TimeFormat

	// Configure output
	if cfg.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: cfg.TimeFormat,
		})
	} else {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	return &zerologLogger{logger: &log.Logger}
}

// Info logs an info message
func (l *zerologLogger) Info(msg string, args ...interface{}) {
	l.logger.Info().Msgf(msg, args...)
}

// Error logs an error message
func (l *zerologLogger) Error(msg string, err error) {
	l.logger.Error().Err(err).Msg(msg)
}

// Fatal logs a fatal message and exits
func (l *zerologLogger) Fatal(msg string, err error) {
	l.logger.Fatal().Err(err).Msg(msg)
}

// Warn logs a warning message
func (l *zerologLogger) Warn(msg string, err error) {
	l.logger.Warn().Err(err).Msg(msg)
}

// Debug logs a debug message
func (l *zerologLogger) Debug(msg string, args ...interface{}) {
	l.logger.Debug().Msgf(msg, args...)
}

// Get returns the global logger instance
func Get() *zerolog.Logger {
	return &log.Logger
}

// WithContext adds context to the logger
func WithContext(ctx interface{}) zerolog.Logger {
	return log.With().Interface("context", ctx).Logger()
}

// WithError adds an error to the logger
func WithError(err error) zerolog.Logger {
	return log.With().Err(err).Logger()
}
