package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/handler"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/service"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/static"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils"
	"github.com/caarlos0/env/v11"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/rs/zerolog/log"
)

func main() {
	log.Printf("Запуск сервера для сбора рантайм-метрик...")
	run()
}

// LoggingMiddleware Здесь мы логируем начало и конец обработки любого запроса, а так же проставляем requestId в логгер
func LoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		reqID := uuid.New().String()
		logger := log.With().Str("requestId", reqID).Logger()
		reqCtx := logger.WithContext(ctx.Request.Context())
		ctx.Request = ctx.Request.WithContext(reqCtx)

		responseData := &utils.ResponseData{
			Status: http.StatusOK, // По умолчанию 200, если хэндлер не вызовет WriteHeader явным образом
			Size:   0,
		}

		lw := &utils.LoggingResponseWriter{
			ResponseWriter: ctx.Writer,
			ResponseData:   responseData,
		}

		ctx.Writer = lw

		logger.Debug().
			Str("host", ctx.Request.Host).
			Str("method", ctx.Request.Method).
			Str("path", ctx.Request.URL.Path).
			Msg("Входящий HTTP запрос")

		ctx.Next()

		duration := time.Since(start)
		// todo перенесла duration в ответ, потому что это логичнее.
		// По логам важно понять, когда запрос поступил, что происходило,
		// и чем закончился. А если лог о поступлении запроса будет после его обработки, на мой взгляд странно
		logger.Debug().
			Dur("duration", duration).
			Int("size", responseData.Size).
			Int("status", responseData.Status).
			Msg("Обработка метода завершена. Отдаем ответ")

	}
}

func run() {

	conf := server.Config{LogLevel: "debug"}

	flag.StringVar(&conf.ServerURL, "a", "localhost:8080", "server url")
	flag.Parse()
	if err := env.Parse(&conf); err != nil {
		log.Printf("failed to parse env variables: %v", err)
	}

	// Настраиваем дефолтный логгер
	utils.InitBaseLogger(utils.LoggerConfig{Level: conf.LogLevel})
	log.Debug().Any("Конфиг", conf).Msg("Содержимое конфига")

	gin.SetMode(gin.ReleaseMode)

	log.Info().Msg("Инициализация http-server-а")
	// инициализируем http-server
	router := gin.New()
	router.Use(LoggingMiddleware())
	router.Use(gin.Recovery())
	fs := static.SetUpFs()
	router.SetHTMLTemplate(fs)

	repoImpl := repository.NewMetricRepositoryImpl()
	serviceImpl := service.NewMetricServiceImpl(repoImpl)

	handler.SetUpRoutes(router, serviceImpl)

	// Запуск сервера на порту
	log.Info().Msg("Инициализация http-server-а выполнена успешно")

	err := router.Run(conf.ServerURL)
	if err != nil {
		log.Info().Any("error", err.Error()).Msg("Ошибка во время запуска http-сервера на порту ")
	}
}
