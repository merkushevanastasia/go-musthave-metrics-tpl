package service

import (
	"context"
	"time"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/mapper"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/dto"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/entity"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
	"github.com/rs/zerolog/log"
)

type MetricService interface {
	UpdateMetric(ctx context.Context, dto *dto.MetricDto) error
	GetMetric(ctx context.Context, metricName string) (dto.MetricDto, error)
	GetAll(ctx context.Context) []dto.MetricDto
	StartStoreTicker(ctx context.Context)
	RestoreMetrics(ctx context.Context) error
	UpdateMetricsBatch(ctx context.Context, dtoList []dto.MetricDto) error
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

func (s *MetricServiceImpl) UpdateMetric(ctx context.Context, dto *dto.MetricDto) error {
	entity := mapper.MapToEntity(*dto)
	err := s.repo.Update(ctx, entity)
	if err != nil {
		return err
	}
	if s.conf.StoreInterval == 0 {
		err := s.repo.CreateDump(ctx)
		if err != nil {
			log.Err(err)
		}
	}
	return nil
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

func (s *MetricServiceImpl) UpdateMetricsBatch(ctx context.Context, dtoList []dto.MetricDto) error {
	if len(dtoList) == 0 {
		return nil
	}

	entities := make([]entity.MetricEntity, 0, len(dtoList))
	for _, d := range dtoList {
		entities = append(entities, mapper.MapToEntity(d))
	}

	if err := s.repo.UpdateBatch(ctx, entities); err != nil {
		return err
	}

	if s.conf.DatabaseDSN == "" && s.conf.StoreInterval == 0 && s.conf.FileStoragePath != "" {
		return s.repo.CreateDump(ctx)
	}

	return nil
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
				log.Info().Msg("Контекст отменен, останавливаем тикер сохранения метрик")
				return
			}
		}
	}()
}
