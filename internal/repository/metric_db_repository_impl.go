package repository

import (
	"context"
	"errors"

	customerrors "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/error"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/entity"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MetricDBRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewMetricDBRepositoryImpl(dbPool *pgxpool.Pool) *MetricDBRepositoryImpl {
	return &MetricDBRepositoryImpl{
		DB: dbPool,
	}
}

func (repository *MetricDBRepositoryImpl) Update(ctx context.Context, metricEntity entity.MetricEntity) {
	log := logs.GetLoggerFromContext(ctx)

	query := `
		INSERT INTO metrics (name, mtype, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name, mtype) DO UPDATE SET
			delta = CASE
				WHEN EXCLUDED.mtype = 'counter' THEN COALESCE(metrics.delta, 0) + COALESCE(EXCLUDED.delta, 0)
				ELSE NULL
			END,
			value = EXCLUDED.value;
	`

	_, err := repository.DB.Exec(
		ctx,
		query,
		metricEntity.ID,
		metricEntity.MType,
		metricEntity.Delta,
		metricEntity.Value,
	)
	if err != nil {
		log.Error().Err(err).Str("name", metricEntity.ID).Msg("Ошибка при сохранении метрики в PostgreSQL")
	}
}

func (repository *MetricDBRepositoryImpl) Get(ctx context.Context, metricName string) (entity.MetricEntity, error) {
	log := logs.GetLoggerFromContext(ctx)
	log.Debug().Str("имя метрики", metricName).Msg("Запрос значения метрики из PostgreSQL")

	query := `
		SELECT name, mtype, delta, value
		FROM metrics
		WHERE name = $1
		LIMIT 1;
	`

	var result entity.MetricEntity
	err := repository.DB.QueryRow(ctx, query, metricName).Scan(
		&result.ID,
		&result.MType,
		&result.Delta,
		&result.Value,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Debug().Str("name", metricName).Msg("Такой метрики не существует в PostgreSQL")
			return entity.MetricEntity{}, customerrors.ErrMetricNotFound
		}
		log.Error().Err(err).Str("name", metricName).Msg("Ошибка при чтении метрики из PostgreSQL")
		return entity.MetricEntity{}, err
	}

	return result, nil
}

func (repository *MetricDBRepositoryImpl) GetAll(ctx context.Context) map[string]entity.MetricEntity {
	log := logs.GetLoggerFromContext(ctx)
	log.Debug().Msg("Получение всех метрик из PostgreSQL")

	query := `SELECT name, mtype, delta, value FROM metrics;`

	rows, err := repository.DB.Query(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("Ошибка при получении всех метрик из PostgreSQL")
		return make(map[string]entity.MetricEntity)
	}
	defer rows.Close()

	result := make(map[string]entity.MetricEntity)
	for rows.Next() {
		var m entity.MetricEntity
		if err := rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value); err != nil {
			log.Error().Err(err).Msg("Ошибка при сканировании строки метрики из PostgreSQL")
			continue
		}
		result[m.ID] = m
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("Ошибка при итерации по строкам PostgreSQL")
	}

	return result
}

func (repository *MetricDBRepositoryImpl) Clean(ctx context.Context) {
	log := logs.GetLoggerFromContext(ctx)
	log.Debug().Msg("Очищаем таблицу метрик в PostgreSQL")

	if _, err := repository.DB.Exec(ctx, `TRUNCATE TABLE metrics RESTART IDENTITY;`); err != nil {
		log.Error().Err(err).Msg("Ошибка при очистке таблицы metrics")
	}
}

func (repository *MetricDBRepositoryImpl) RestoreMetrics(ctx context.Context) error {
	log := logs.GetLoggerFromContext(ctx)
	log.Info().Msg("Для PostgreSQL метод RestoreMetrics не требуется")
	return nil
}

func (repository *MetricDBRepositoryImpl) CreateDump(ctx context.Context) error {
	log := logs.GetLoggerFromContext(ctx)
	log.Info().Msg("Для PostgreSQL метод CreateDump не требуется")
	return nil
}
