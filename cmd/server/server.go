package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/handler"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/service"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/static"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs/lifecyrcle"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/middleware"
	"github.com/caarlos0/env/v11"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const moduleName = "Модуль хранения метрик"

func main() {
	lifecyrcle.LogModuleStart(&log.Logger, moduleName)
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

	logs.InitBaseLogger(logs.LoggerConfig{Level: conf.LogLevel})
	log.Debug().Any("Конфиг", conf).Msg("Содержимое конфига")

	gin.SetMode(gin.ReleaseMode)

	log.Info().Msg("Инициализация http-server-а")
	router := gin.New()
	router.Use(middleware.LoggingRequestResponseCommonMiddleware())
	router.Use(middleware.MultipleReadMiddleware())
	router.Use(middleware.CompressMiddleware())
	router.Use(middleware.LoggingDetailsMiddleware())
	router.Use(gin.Recovery())

	fs := static.SetUpFs()
	router.SetHTMLTemplate(fs)

	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	deps := initDeps(conf)
	if err := deps.MetricService.RestoreMetrics(); err != nil {
		log.Printf("failed to restore: %v", err)
	}
	deps.MetricService.StartStoreTicker(mainCtx)

	handler.SetUpRoutes(router, deps.MetricService)

	log.Info().Msg("Инициализация http-server-а выполнена успешно")

	srv := &http.Server{
		Addr:    conf.ServerURL,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("Ошибка во время работы http-сервера")
		}
	}()

	lifecyrcle.LogModuleStartedOk(&log.Logger, moduleName)

	<-mainCtx.Done()

	lifecyrcle.LogModuleDown(&log.Logger, moduleName)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Ошибка при принудительной остановке HTTP-сервера")
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
