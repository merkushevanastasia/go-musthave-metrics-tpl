package mapper

import (
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/dto"
	entity "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/entity"
)

func MapToDto(entity entity.MetricEntity) dto.MetricDto {
	return dto.MetricDto{
		ID:    entity.ID,
		MType: entity.MType,
		Delta: entity.Delta,
		Value: entity.Value,
	}
}

func MapToEntity(dto dto.MetricDto) entity.MetricEntity {
	return entity.MetricEntity{
		ID:    dto.ID,
		MType: dto.MType,
		Delta: dto.Delta,
		Value: dto.Value,
	}
}
