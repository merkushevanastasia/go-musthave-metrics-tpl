package logging

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
type (
	// ResponseData создаём структуру для хранения сведений об ответе
	ResponseData struct {
		Status int
		Size   int
	}

	// LoggingResponseWriter добавляем реализацию gin.ResponseWriter
	LoggingResponseWriter struct {
		gin.ResponseWriter
		ResponseData *ResponseData
	}
)

func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.ResponseData.Size += size
	return size, err
}

func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.Status = statusCode
}

// InitBaseLogger Инициализируем и настраиваем базовый логер
func InitBaseLogger(config LoggerConfig) {
	log.Info().Msg("Инициализация настроек базового логера...")
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logLevel := parseLogLevel(config.Level)
	zerolog.SetGlobalLevel(logLevel)
	baseLogger := zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	log.Logger = baseLogger
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

func FromContext(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}
