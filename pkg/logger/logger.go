package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Level      string
	TimeFormat string
	Pretty     bool
}

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, err error)
	Fatal(msg string, err error)
	Warn(msg string, err error)
	Debug(msg string, args ...interface{})
}

type zerologLogger struct {
	logger *zerolog.Logger
}

func New(cfg Config) Logger {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}
	zerolog.TimeFieldFormat = cfg.TimeFormat

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

func (l *zerologLogger) Info(msg string, args ...interface{}) {
	l.logger.Info().Msgf(msg, args...)
}

func (l *zerologLogger) Error(msg string, err error) {
	l.logger.Error().Err(err).Msg(msg)
}

func (l *zerologLogger) Fatal(msg string, err error) {
	l.logger.Fatal().Err(err).Msg(msg)
}

func (l *zerologLogger) Warn(msg string, err error) {
	l.logger.Warn().Err(err).Msg(msg)
}

func (l *zerologLogger) Debug(msg string, args ...interface{}) {
	l.logger.Debug().Msgf(msg, args...)
}


func Get() *zerolog.Logger {
	return &log.Logger
}


func WithContext(ctx interface{}) zerolog.Logger {
	return log.With().Interface("context", ctx).Logger()
}

		
func WithError(err error) zerolog.Logger {
	return log.With().Err(err).Logger()
}
