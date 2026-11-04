package middleware

import (
	"io"
	"net/http"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/middleware/requestresponse"
	"github.com/gin-gonic/gin"
)

func MultipleReadMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//Получили логгер с MDC
		logger := logs.GetLoggerFromGin(ctx)

		rawBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			logger.Error().Err(err).Msg("Ошибка во время чтения тела запроса")
			return
		}
		err = ctx.Request.Body.Close()
		if err != nil {
			logger.Error().Err(err).Msg("Ошибка во время закрытия  c.Request.Body.Close()")
		}

		// Инициализируем наш объект для многократного чтения данных
		reqMulti := &requestresponse.RequestMultiple{
			RequestBody: &rawBytes,
		}

		ctx.Request.Body = reqMulti

		resMulti := &requestresponse.ResponseMultiple{
			ResponseWriter: ctx.Writer,
			BodyResponse:   make([]byte, 0, 1024),
		}
		originalWriter := ctx.Writer
		ctx.Writer = resMulti
		ctx.Next()
		err = resMulti.FlushCustom()
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			logger.Error().Err(err).Msg("Ошибка во время FlushCustom")
			return
		}
		ctx.Writer = originalWriter
	}
}
