package requestresponse

import (
	"bytes"
	"io"
)

type RequestMultiple struct {
	reader *bytes.Reader
	// Это байты запроса для многократного использования
	RequestBody *[]byte
}

// Read реализует интерфейс io.Reader
func (r *RequestMultiple) Read(p []byte) (n int, err error) {

	// Сохраняем уже прочитанные байты для повторной вычитки
	if r.reader == nil {
		r.reader = bytes.NewReader(*r.RequestBody)
	}
	return r.reader.Read(p)
}

func (r *RequestMultiple) Close() error {
	r.reader = nil
	return nil
}

func (r *RequestMultiple) Reset() {
	if r.reader != nil {
		_, _ = r.reader.Seek(0, io.SeekStart)
	}
}
