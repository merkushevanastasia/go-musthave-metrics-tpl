package middleware

import (
	"net/http"
	"time"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/middleware/request_response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		reqID := uuid.New().String()
		logger := log.With().Str("requestId", reqID).Logger()
		reqCtx := logger.WithContext(ctx.Request.Context())
		ctx.Request = ctx.Request.WithContext(reqCtx)

		// 2. Инициализируем наш объект
		reqMulti, ok := ctx.Request.Body.(*request_response.RequestMultiple)

		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		logger.Debug().
			Str("host", ctx.Request.Host).
			Str("method", ctx.Request.Method).
			Str("path", ctx.Request.URL.Path).
			Str("body", string(*reqMulti.RequestBody)).
			Msg("Входящий HTTP запрос")

		ctx.Next()
		resMulti, ok := ctx.Writer.(*request_response.ResponseMultiple)
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		duration := time.Since(start)
		// todo перенесла duration в ответ, потому что это логичнее.
		// По логам важно понять, когда запрос поступил, что происходило,
		// и чем закончился. А если лог о поступлении запроса будет после его обработки, на мой взгляд странно
		logger.Debug().
			Dur("duration", duration).
			Int("size", len(resMulti.BodyResponse)).
			Int("status", resMulti.Status()).
			Str("body", string(resMulti.BodyResponse)).
			Msg("Обработка метода завершена. Отдаем ответ")

	}
}
