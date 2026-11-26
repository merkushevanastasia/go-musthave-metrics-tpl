package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	config "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/dto"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs"
	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog/log"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/agent"
)

func main() {
	log.Info().Msg("Запуск агента по сбору метрик...")

	conf := config.Config{LogLevel: "debug"}

	flag.StringVar(&conf.ServerURL, "a", "localhost:8080", "server url")
	flag.IntVar(&conf.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&conf.PollInterval, "p", 2, "poll interval")
	flag.Parse()

	if err := env.Parse(&conf); err != nil {
		log.Error().Err(err).Msg("Ошибка во время парсинга конфигурации")
	}
	checkServerURL(&conf)

	logs.InitBaseLogger(logs.LoggerConfig{Level: conf.LogLevel})
	log.Debug().Any("Конфиг", conf).Msg("Содержимое конфига")

	agent.InitClient(conf)

	metrics := dto.NewMetricDtoCollection()

	ctx, cancel := context.WithCancel(context.Background())
	collectTicker := time.NewTicker(time.Duration(conf.PollInterval) * time.Second)
	sendTicker := time.NewTicker(time.Duration(conf.ReportInterval) * time.Second)
	defer collectTicker.Stop()
	defer sendTicker.Stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-collectTicker.C:
				agent.Update(metrics)
			case <-sendTicker.C:
				agent.SendAll(ctx, metrics)
			case <-ctx.Done():
				return
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigChan

	log.Info().Msg("Завершение работы мониторинга...")
	cancel()
	wg.Wait()
}

func checkServerURL(conf *config.Config) {
	if !strings.HasPrefix(conf.ServerURL, "http://") && !strings.HasPrefix(conf.ServerURL, "https://") {
		log.Info().Str("Переданный адрес", conf.ServerURL).Msg("Добавляем http:// для resty")
		conf.ServerURL = "http://" + conf.ServerURL
	}
}
