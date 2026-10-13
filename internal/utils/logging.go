package utils

import (
	"context"
	"log/slog"
	"os"
)

type ctxKey struct{}

var LoggerKey = ctxKey{}

type LoggerConfig struct {
	Level string
}

// InitBaseLogger Инициализируем и настраиваем базовый логер
func InitBaseLogger(config LoggerConfig) {
	slog.Info("Инициализация настроек базового логера...")
	logLevel := parseLogLevel(config.Level)
	slog.Info("Базовый логгер успешно проинициализировн")
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	baseLogger := slog.New(handler)
	slog.SetDefault(baseLogger)
}

// parseLogLevel парсим из строки уровень логирования
func parseLogLevel(levelStr string) slog.Level {
	switch levelStr {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}

}

func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
