package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/middleware/request_response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func CompressMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		log.Debug().Msg("CompressMiddleware start")

		// Пытаемся привести c.Request.Body к нашему типу *RequestMultiple
		reqMulti, ok := c.Request.Body.(*request_response.RequestMultiple)

		if !ok {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		reqMulti.IsRequestCompress = c.GetHeader("Content-Encoding") == "gzip"
		if reqMulti.IsRequestCompress {
			reader, err := gzip.NewReader(bytes.NewReader(*reqMulti.RequestBody))
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			uncompress, err := io.ReadAll(reader)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			reqMulti.RequestBody = &uncompress
		}

		c.Next()

		// Пытаемся привести c.Writer к нашему типу *ResponseMultiple
		resMulti, ok := c.Writer.(*request_response.ResponseMultiple)
		if !ok {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		contentType := resMulti.Header().Get("Content-Type")

		if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
			resMulti.Header().Set("Content-Encoding", "gzip")
			resMulti.Header().Add("Vary", "Accept-Encoding")
			compressBytes, err := CompressBytes(resMulti.BodyResponse)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			resMulti.BodyResponse = compressBytes
		}

		resMulti.ResponseWriter.WriteHeader(resMulti.Status())
		_, _ = resMulti.ResponseWriter.Write(resMulti.BodyResponse)

		log.Debug().Msg("CompressMiddleware end")

	}
}

func CompressBytes(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	// 1. Создаем буфер в памяти, куда gzip будет записывать сжатые данные
	var buf bytes.Buffer

	// 2. Инициализируем gzip-писатель поверх нашего буфера
	gzWriter := gzip.NewWriter(&buf)

	// 3. Записываем несжатые байты в gzip-писатель
	_, err := gzWriter.Write(data)
	if err != nil {
		return nil, err
	}

	// 4. Обязательно (!) закрываем gzip-писатель прямо сейчас.
	// Если этого не сделать, gzip не запишет финальные контрольные суммы (checksum),
	// и архив останется поврежденным/неполным.
	err = gzWriter.Close()
	if err != nil {
		return nil, err
	}

	// 5. Возвращаем итоговый массив сжатых байт
	return buf.Bytes(), nil
}
