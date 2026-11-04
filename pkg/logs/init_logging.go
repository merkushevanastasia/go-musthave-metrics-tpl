package logs

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LoggerConfig struct {
	Level string
}

// InitBaseLogger Инициализируем и настраиваем базовый логер
func InitBaseLogger(config LoggerConfig) {
	log.Info().Msg("Инициализация настроек базового логера...")
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logLevel := parseLogLevel(config.Level)
	zerolog.SetGlobalLevel(logLevel)
	logger := zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	log.Logger = logger
	log.Info().Msg("Базовый логгер успешно проинициализирован")
}

// parseLogLevel парсим из строки уровень логирования
func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

func GetLoggerFromContext(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

func GetLoggerFromGin(ctx *gin.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx.Request.Context())
}
