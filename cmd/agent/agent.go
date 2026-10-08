package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	config "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/dto"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils"
)

func main() {

	//получаем настройки, переданные в аргументах запуска
	conf := config.Config{LogLevel: "debug"}
	serverURL := flag.String("a", "localhost:8080", "server url")
	flag.IntVar(&conf.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&conf.PollInterval, "p", 2, "report interval")
	flag.Parse()
	//Создаем базовый клиент для запросов на сервер Метрик

	conf.ServerURL = *serverURL
	agent.InitClient(conf)

	utils.InitBaseLogger(utils.LoggerConfig{Level: conf.LogLevel})
	slog.Info("Запуск мониторинга...")

	values := dto.NewMetricCollection()
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
				agent.Update(values)
			case <-sendTicker.C:
				agent.SendAll(values)
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
