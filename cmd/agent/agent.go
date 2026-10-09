package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	config "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/agent"
)

func main() {

	slog.Info("Запуск агента по сбору метрик ...")

	// Получаем настройки, переданные в аргументах запуска
	conf := config.Config{LogLevel: "debug"}
	flag.StringVar(&conf.ServerURL, "a", "localhost:8080", "server url")
	flag.IntVar(&conf.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&conf.PollInterval, "p", 2, "report interval")
	flag.Parse()
	checkServerURL(&conf)

	utils.InitBaseLogger(utils.LoggerConfig{Level: conf.LogLevel})

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

	slog.Info("Завершение работы мониторинга...")
}

func checkServerURL(conf *config.Config) {
	if !strings.HasPrefix(conf.ServerURL, "http://") {
		conf.ServerURL = "http://" + conf.ServerURL
	}
}
