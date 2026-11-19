package rest

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func LogRequestBody(logger *zerolog.Logger, contentType string, body []byte) {
	logger.Debug().
		Str("body", getBodyToLog(contentType, body)).
		Msg("Тело входящего HTTP-запроса")
}

func LogRequest(logger *zerolog.Logger, host string, method string, path string, contentType string, bodySize int64, headers string) {
	logger.Info().
		Str("host", host).
		Str("method", method).
		Str("path", path).
		Msg("Входящий HTTP-апрос")
	logger.Debug().
		Str("host", host).
		Str("method", method).
		Str("path", path).
		Str("contentType", contentType).
		Int64("body size", bodySize).
		Str("headers", headers).
		Msg("Входящий HTTP-запрос")
}

func LogError(logger *zerolog.Logger, err error, message string) {
	logger.Error().Err(err).Msg(message)
}
func LogResponse(logger *zerolog.Logger, status int, contentType string, duration time.Duration, bodySize int, headers string) {
	logger.Info().
		Int("status", status).
		Msg("Исходящий  HTTP-ответ")
	logger.Debug().
		Dur("duration", duration).
		Int("status", status).
		Str("contentType", contentType).
		Str("headers", headers).
		Int("body size", bodySize).
		Msg("Исходящий  HTTP-ответ")
}

func LogResponseBody(logger *zerolog.Logger, contentType string, body []byte) {
	logger.Debug().
		Str("body", getBodyToLog(contentType, body)).
		Msg("Исходящее тело ответа")
}
func getBodyToLog(contentType string, body []byte) string {
	ct := strings.ToLower(contentType)

	if strings.Contains(ct, "text/") ||
		strings.Contains(ct, "json") ||
		strings.Contains(ct, "xml") {
		return string(body)
	}

	return fmt.Sprintf("%d bytes", len(body))
}
