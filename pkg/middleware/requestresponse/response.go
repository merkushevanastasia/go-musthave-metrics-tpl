package requestresponse

import (
	"compress/gzip"

	"github.com/gin-gonic/gin"
)

type ResponseMultiple struct {
	gin.ResponseWriter
	//Тут хранятся байты для многоразового использования
	BodyResponse []byte
	// Здесь будет храниться сжатый gzip ответ
	GzWriter *gzip.Writer
}

func (w *ResponseMultiple) Write(b []byte) (int, error) {
	w.BodyResponse = append(w.BodyResponse, b...)
	return len(b), nil
}

func (w *ResponseMultiple) ReWrite(b []byte) {
	w.BodyResponse = b
}

func (w *ResponseMultiple) FlushCustom() error {
	w.ResponseWriter.WriteHeader(w.Status())
	_, err := w.ResponseWriter.Write(w.BodyResponse)
	return err
}
