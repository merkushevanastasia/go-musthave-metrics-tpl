package middleware

import (
	"net/http"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs/rest"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/middleware/requestresponse"
	"github.com/gin-gonic/gin"
)

// LoggingDetailsMiddleware фиксирует в журнале данные о теле запроса и ответа
func LoggingDetailsMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		logger := logs.GetLoggerFromGin(ctx)
		reqRespMulti, ok := ctx.Request.Body.(*requestresponse.RequestMultiple)

		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			rest.LogError(logger, nil, "Не удалось преобразовать запрос в requestresponse.ResponseMultiple")
			return
		}
		rest.LogRequestBody(logger, ctx.Writer.Header().Get("Content-Type"), *reqRespMulti.RequestBody)

		ctx.Next()

		resMulti, ok := ctx.Writer.(*requestresponse.ResponseMultiple)
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			rest.LogError(logger, nil, "Не удалось преобразовать ответ в requestresponse.ResponseMultiple")
			return
		}

		contentType := resMulti.Header().Get("Content-Type")
		rest.LogResponseBody(logger, contentType, resMulti.BodyResponse)

	}

}
