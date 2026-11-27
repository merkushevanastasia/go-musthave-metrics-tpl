package middleware

import (
	"net/http"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs/rest"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/middleware/requestresponse"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/utils"
	"github.com/gin-gonic/gin"
)

func SignMiddleware(cryptoKey string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := logs.GetLoggerFromGin(ctx)

		isSignPresent := false
		if cryptoKey != "" {
			isSignPresent = true
			signature := ctx.GetHeader(utils.AlgSignHeader)
			if signature != "" {

				reqMulti, ok := ctx.Request.Body.(*requestresponse.RequestMultiple)
				if !ok {
					rest.LogError(logger, nil, "Не удалось преобразовать запрос в requestresponse.ResponseMultiple")
					ctx.AbortWithStatus(http.StatusInternalServerError)
					return
				}

				isValid, err := utils.CheckSign(signature, reqMulti.RequestBody, cryptoKey)

				if err != nil || !isValid {
					rest.LogError(logger, err, "Во время проверки подписи произошла ошибка")
					ctx.AbortWithStatus(http.StatusBadRequest)
					return
				}
			}

		}

		ctx.Next()

		if isSignPresent {
			resMulti, ok := ctx.Writer.(*requestresponse.ResponseMultiple)
			if !ok {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				rest.LogError(logger, nil, "Не удалось преобразовать ответ в requestresponse.ResponseMultiple")
				return
			}

			resMulti.Header().Set(utils.AlgSignHeader, utils.Sign(resMulti.BodyResponse, cryptoKey))

		}

	}
}
