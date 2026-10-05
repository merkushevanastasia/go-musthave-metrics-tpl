package error

import (
	"errors"
)

var ErrNotValidMetricName = errors.New("в запросе отсутствует имя метрики")
var ErrNotValidMetricValue = errors.New("в запросе невалидное значение метрики")
var ErrNotAllowedMethod = errors.New("данный тип запроса не поддерживается")
var ErrNotAllowedMetricType = errors.New("данный тип метрики не поддерживается")
