package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs/rest"
	requestresponse2 "github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/middleware/requestresponse"
	"github.com/gin-gonic/gin"
)

func CompressMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		logger := logs.GetLoggerFromGin(ctx)

		reqMulti, ok := ctx.Request.Body.(*requestresponse2.RequestMultiple)
		if !ok {
			rest.LogError(logger, nil, "Не удалось преобразовать запрос в requestresponse.ResponseMultiple")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Если данные были сжаты, то приводим их к изначальному виду
		if ctx.GetHeader("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(bytes.NewReader(*reqMulti.RequestBody))
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			uncompress, err := io.ReadAll(reader)
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			reqMulti.RequestBody = &uncompress
		}

		ctx.Next()

		resMulti, ok := ctx.Writer.(*requestresponse2.ResponseMultiple)
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			rest.LogError(logger, nil, "Не удалось преобразовать ответ в requestresponse.ResponseMultiple")
			return
		}

		if isNeedResultCompress(resMulti) {
			resMulti.Header().Set("Content-Encoding", "gzip")
			resMulti.Header().Add("Vary", "Accept-Encoding")
			compressBytes, err := compressBytes(resMulti.BodyResponse)
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				rest.LogError(logger, nil, "Не удалось сжать ответ")
				return
			}
			resMulti.BodyResponse = compressBytes
			resMulti.ReWrite(compressBytes)
		}

		resMulti.ResponseWriter.WriteHeader(resMulti.Status())
	}
}

// Нужно ли сжимать ответ
func isNeedResultCompress(resMulti *requestresponse2.ResponseMultiple) bool {
	acceptEncoding := resMulti.Header().Get("Accept-Encoding")
	supportsGzip := strings.Contains(acceptEncoding, "gzip")
	contentType := resMulti.Header().Get("Content-Type")
	return supportsGzip && (strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html"))
}

// Сжимаем данные
func compressBytes(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	_, err := gzWriter.Write(data)
	if err != nil {
		return nil, err
	}
	err = gzWriter.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
