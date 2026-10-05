package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/handler"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils"
	"github.com/google/uuid"
)

var serverPort = 8080
var level = "debug"

func main() {
	slog.Info("Запуск сервера для сбора рантайм-метрик...")
	run()
}

// LoggingMiddleware Здесь мы логируем начало и конец обработки любого запроса, а так же проставляем requestId в логгер
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		reqID := uuid.New()
		var logger = slog.Default().With("requestId", reqID)
		logger.Debug("Входящий HTTP запрос",
			"method", r.Method,
			"path", r.URL.Path,
		)
		ctx = context.WithValue(ctx, utils.LoggerKey, logger)

		next.ServeHTTP(w, r.WithContext(ctx))

		logger.Debug("Обработка метода завершена. Отдаем ответ")
	})
}

func run() {

	utils.InitBaseLogger(utils.LoggerConfig{Level: level})

	slog.Info("Инициализация http-server-а...")
	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/update/{%s}/{%s}/{%s}", handler.MetricTypePathKey, handler.MetricNamePathKey, handler.MetricValuePathKey), handler.HandleMetricUpdate)
	// Оборачиваем ВЕСЬ роутер в middleware для логирования каждого запроса и простановки MDC
	wrappedMux := LoggingMiddleware(mux)
	slog.Info("Инициализация http-server-а выполнена успешно. Запуск...")
	err := http.ListenAndServe(fmt.Sprintf(":%d", serverPort), wrappedMux)
	if err != nil {
		slog.Error("Ошибка во время запуска http-сервера на порту ", slog.Any("err", err.Error()))
	}
}
