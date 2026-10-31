package request_response

import (
	"compress/gzip"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ResponseMultiple struct {
	gin.ResponseWriter
	//Тут хранятся байты для многоразового использования
	BodyResponse []byte
	//Тут хранятся байты сжатия для многоразового использования
	BodyResponseCompress *[]byte
	// Тут сохраняется статус запроса
	ResponseStatus int64
	//Тип контента
	ContentType string
	// Здесь будет храниться сжатый gzip ответ
	GzWriter *gzip.Writer
	// Это флаг для отслеживания того,установили ли уже хэдеры
	HeaderWritten bool
}

func (w *ResponseMultiple) Write(b []byte) (int, error) {
	log.Debug().Any("ResponseMultipleValue", w).Msg("ResponseMultiple - Write")
	return w.ResponseWriter.Write(b)
}

func (w *ResponseMultiple) ReWrite(b []byte) (int, error) {
	log.Debug().Any("ResponseMultipleValue", w).Msg("ResponseMultiple - Write")
	w.BodyResponse = append(w.BodyResponse[:0], b...)
	return w.ResponseWriter.Write(b)
}
