package middleware

import (
	"io"
	"net/http"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/middleware/request_response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func MultipleReadMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debug().Msg("Multiple read middleware start")
		rawBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		_ = c.Request.Body.Close()

		// 2. Инициализируем наш объект
		reqMulti := &request_response.RequestMultiple{
			RequestBody: &rawBytes,
			ContentType: c.GetHeader("Content-Type"),
		}

		c.Request.Body = reqMulti

		resMulti := &request_response.ResponseMultiple{
			ResponseWriter: c.Writer,
			BodyResponse:   make([]byte, 0, 1024),
		}
		originalWriter := c.Writer
		c.Writer = resMulti
		c.Next()
		c.Writer = originalWriter

		log.Debug().Msg("Multiple read middleware end")

	}

}
