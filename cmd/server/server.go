package main

import (
	"context"
	"flag"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/handler"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/middleware"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/service"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/static"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils/logging"
	"github.com/caarlos0/env/v11"
	"github.com/gin-gonic/gin"
	log "github.com/rs/zerolog/log"
)

func main() {
	log.Printf("Запуск сервера для сбора рантайм-метрик...")
	run()
}

type Deps struct {
	MetricRepository repository.MetricRepository
	MetricService    service.MetricService
}

func run() {
	conf := server.Config{LogLevel: "debug"}

	flag.StringVar(&conf.ServerURL, "a", "localhost:8080", "server url")
	flag.IntVar(&conf.StoreInterval, "i", 10, "store interval in seconds")
	flag.StringVar(&conf.FileStoragePath, "f", "/tmp/metrics.json", "file storage path")
	flag.BoolVar(&conf.Restore, "r", false, "restore previous values from file")
	flag.Parse()
	if err := env.Parse(&conf); err != nil {
		log.Printf("failed to parse env variables: %v", err)
	}

	logging.InitBaseLogger(logging.LoggerConfig{Level: conf.LogLevel})
	log.Debug().Any("Конфиг", conf).Msg("Содержимое конфига")

	gin.SetMode(gin.ReleaseMode)

	log.Info().Msg("Инициализация http-server-а")
	router := gin.New()

	router.Use(middleware.MultipleReadMiddleware())
	router.Use(middleware.CompressMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(gin.Recovery())

	fs := static.SetUpFs()
	router.SetHTMLTemplate(fs)

	deps := initDeps(conf)
	err := deps.MetricService.RestoreMetrics()
	if err := env.Parse(&conf); err != nil {
		log.Printf("failed to restore: %v", err)
	}
	deps.MetricService.StartStoreTicker(context.Background())

	handler.SetUpRoutes(router, deps.MetricService)

	log.Info().Msg("Инициализация http-server-а выполнена успешно")

	err = router.Run(conf.ServerURL)
	if err != nil {
		log.Info().Any("error", err.Error()).Msg("Ошибка во время запуска http-сервера на порту ")
	}
}

func initDeps(config server.Config) Deps {
	repoImpl := repository.NewMetricRepositoryImpl()
	serviceImpl := service.NewMetricServiceImpl(repoImpl, config)
	return Deps{
		MetricRepository: repoImpl,
		MetricService:    serviceImpl,
	}
}
