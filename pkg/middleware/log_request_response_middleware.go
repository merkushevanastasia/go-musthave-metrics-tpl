package middleware

import (
	"strings"
	"time"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs/rest"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// LoggingRequestResponseCommonMiddleware фиксирует в журнале общие данные, не обращаясь к телу запроса и ответа, проставляет MDC
func LoggingRequestResponseCommonMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		//Для дальнейшего разбора обработки запроса добавляем MDC uuid и добавляем его в логи на протяжении всей обработки запроса
		logger := log.With().Str("requestId", uuid.New().String()).Logger()
		reqCtx := logger.WithContext(ctx.Request.Context())
		ctx.Request = ctx.Request.WithContext(reqCtx)

		rest.LogRequest(&logger, ctx.Request.Host, ctx.Request.Method, ctx.Request.URL.Path, ctx.GetHeader("Content-Type"), ctx.Request.ContentLength, requestHeadersStr(ctx))

		ctx.Next()

		duration := time.Since(start)
		rest.LogResponse(&logger, ctx.Writer.Status(), ctx.Writer.Header().Get("Content-Type"), duration, ctx.Writer.Size(), responseHeadersStr(ctx))
	}
}

// Получение header-ов виде строки для запроса
func requestHeadersStr(c *gin.Context) string {
	var reqBuf strings.Builder
	_ = c.Request.Header.Write(&reqBuf)
	return reqBuf.String()
}

// Получение header-ов виде строки для ответа
func responseHeadersStr(c *gin.Context) string {
	var resBuf strings.Builder
	_ = c.Writer.Header().Write(&resBuf)
	return resBuf.String()
}
