package service

import (
	"context"
	"time"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/mapper"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/dto"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
	"github.com/rs/zerolog/log"
)

type MetricService interface {
	UpdateMetric(ctx context.Context, dto *dto.MetricDto)
	GetMetric(ctx context.Context, metricName string) (dto.MetricDto, error)
	GetAll(ctx context.Context) []dto.MetricDto
	StartStoreTicker(ctx context.Context)
	RestoreMetrics(ctx context.Context) error
}
type MetricServiceImpl struct {
	repo repository.MetricRepository
	conf server.Config
}

func NewMetricServiceImpl(repo repository.MetricRepository, config server.Config) *MetricServiceImpl {
	return &MetricServiceImpl{
		repo: repo,
		conf: config,
	}
}

func (s *MetricServiceImpl) UpdateMetric(ctx context.Context, dto *dto.MetricDto) {
	entity := mapper.MapToEntity(*dto)
	s.repo.Update(ctx, entity)
	if s.conf.StoreInterval == 0 {
		err := s.repo.CreateDump(ctx)
		if err != nil {
			log.Err(err)
		}
	}
}

func (s *MetricServiceImpl) GetMetric(ctx context.Context, metricName string) (dto.MetricDto, error) {

	result, err := s.repo.Get(ctx, metricName)
	if err != nil {
		return dto.MetricDto{}, err
	}

	return mapper.MapToDto(result), nil
}

func (s *MetricServiceImpl) GetAll(ctx context.Context) []dto.MetricDto {
	metrics := s.repo.GetAll(ctx)

	result := make([]dto.MetricDto, 0)

	for _, value := range metrics {
		result = append(result, mapper.MapToDto(value))
	}

	return result
}

func (s *MetricServiceImpl) RestoreMetrics(ctx context.Context) error {
	return s.repo.RestoreMetrics(ctx)
}

func (s *MetricServiceImpl) StartStoreTicker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Duration(s.conf.StoreInterval) * time.Second)
		defer ticker.Stop()

		log.Info().Msg("Запущен тикер сохранения метрик в файл")

		for {
			select {
			case <-ticker.C:
				if err := s.repo.CreateDump(ctx); err != nil {
					log.Error().Err(err).Msg("Ошибка при периодическом сохранении метрик")
				}
			case <-ctx.Done():
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				log.Info().Msg("Контекст отменен, останавливаем тикер сохранения метрик")
				if err := s.repo.CreateDump(shutdownCtx); err != nil {
					log.Error().Err(err).Msg("Ошибка при финальном сохранении метрик перед выходом")
				}
				cancel()
				return
			}
		}
	}()
}
