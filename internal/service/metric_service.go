package service

import (
	"context"
	"encoding/json"
	"os"
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
	RestoreMetrics() error
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
		err := s.saveMetricsInFile()
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

func (s *MetricServiceImpl) saveMetricsInFile() error {
	if s.conf.FileStoragePath == "" {
		return nil
	}

	metricsMap := s.GetAll(context.Background())

	var items []dto.MetricDto
	items = append(items, metricsMap...)

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.OpenFile(s.conf.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func (s *MetricServiceImpl) RestoreMetrics() error {

	if !s.conf.Restore {
		return nil
	}
	log.Info().Msg("Начинаем восстановление из файла" + s.conf.FileStoragePath)

	data, err := os.ReadFile(s.conf.FileStoragePath)
	if os.IsNotExist(err) {
		log.Info().Msg("Файл хранения метрик отсутствует, загрузка пропущена")
		return nil
	}
	if err != nil {
		return err
	}

	var items []dto.MetricDto
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	log.Info().Any("Содержимое файла", items).Msg("Содержимое")

	s.repo.Clean(context.Background())
	for _, item := range items {
		s.UpdateMetric(context.Background(), &item)
	}
	return nil
}

func (s *MetricServiceImpl) StartStoreTicker(ctx context.Context) {
	if s.conf.StoreInterval > 0 && s.conf.FileStoragePath != "" {
		go func() {
			ticker := time.NewTicker(time.Duration(s.conf.StoreInterval) * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if err := s.saveMetricsInFile(); err != nil {
					log.Error().Err(err).Msg("Ошибка при периодическом сохранении метрик")
				}
			}
		}()
	}
}
