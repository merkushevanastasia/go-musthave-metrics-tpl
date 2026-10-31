package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	config "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils/logging"
	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog/log"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/agent"
)

func main() {

	log.Info().Msg("Запуск агента по сбору метрик ...")

	// Получаем настройки, переданные в аргументах запуска
	conf := config.Config{LogLevel: "debug"}

	flag.StringVar(&conf.ServerURL, "a", "localhost:8080", "server url")
	flag.IntVar(&conf.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&conf.PollInterval, "p", 2, "report interval")
	flag.Parse()
	if err := env.Parse(&conf); err != nil {
		log.Error().Any("err", err).Msg("Ошибка во время парсинга конфигурации")
	}
	checkServerURL(&conf)

	logging.InitBaseLogger(logging.LoggerConfig{Level: conf.LogLevel})
	log.Debug().Any("Конфиг", conf).Msg("Содержимое конфига")

	// Инициаоизируем базовый клиент для запросов на сервер Метрик
	agent.InitClient(conf)

	metrics := dto.NewMetricCollection()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	collectTicker := time.NewTicker(time.Duration(conf.PollInterval) * time.Second)
	sendTicker := time.NewTicker(time.Duration(conf.ReportInterval) * time.Second)
	defer collectTicker.Stop()
	defer sendTicker.Stop()

	go func() {
		for {
			select {
			case <-collectTicker.C:
				agent.Update(metrics)
			case <-sendTicker.C:
				agent.SendAll(metrics)
			case <-ctx.Done():
				return
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Info().Msg("Завершение работы мониторинга...")
}

func checkServerURL(conf *config.Config) {
	if !strings.HasPrefix(conf.ServerURL, "http://") {
		log.Info().Any("Переданный адрес", conf.ServerURL).Msg("Добавляем http:// для resty")
		conf.ServerURL = "http://" + conf.ServerURL
	}
}
