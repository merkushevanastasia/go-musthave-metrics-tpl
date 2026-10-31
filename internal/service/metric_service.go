package service

import (
	"context"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/mapper"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/dto"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
)

type MetricService interface {
	UpdateMetric(ctx context.Context, dto *dto.MetricDto)
	GetMetric(ctx context.Context, metricName string) (dto.MetricDto, error)
	GetAll(ctx context.Context) []dto.MetricDto
}

type MetricServiceImpl struct {
	repo repository.MetricRepository
}

func NewMetricServiceImpl(repo repository.MetricRepository) *MetricServiceImpl {
	return &MetricServiceImpl{
		repo: repo,
	}
}

func (s *MetricServiceImpl) UpdateMetric(ctx context.Context, dto *dto.MetricDto) {
	entity := mapper.MapToEntity(*dto)
	s.repo.Update(ctx, entity)
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
