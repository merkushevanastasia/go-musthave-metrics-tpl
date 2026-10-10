package error

import (
	"errors"
)

var ErrNotValidMetricName = errors.New("в запросе отсутствует имя метрики")
var ErrNotValidMetricValue = errors.New("в запросе невалидное значение метрики")
var ErrNotAllowedMetricType = errors.New("данный тип метрики не поддерживается")
var ErrMetricNotFound = errors.New("метрика не найдена")
