package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/agent"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils"
)

var pollInterval = 2 * time.Second
var reportInterval = 10 * time.Second
var serverUrl = "http://localhost:8080"
var level = "debug"

func main() {

	utils.InitBaseLogger(utils.LoggerConfig{Level: level})
	slog.Info("Запуск мониторинга...")

	values := agent.NewCurrentValues()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	collectTicker := time.NewTicker(pollInterval)
	sendTicker := time.NewTicker(reportInterval)
	defer collectTicker.Stop()
	defer sendTicker.Stop()

	go func() {
		for {
			select {
			case <-collectTicker.C:
				agent.Update(values)
			case <-sendTicker.C:
				agent.SendAll(values, serverUrl)
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
