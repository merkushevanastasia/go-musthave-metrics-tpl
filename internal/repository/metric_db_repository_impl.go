package repository

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	customerrors "github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/error"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/model/entity"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/pkg/logs"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var pgRetryDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

type MetricDBRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewMetricDBRepositoryImpl(dbPool *pgxpool.Pool) *MetricDBRepositoryImpl {
	return &MetricDBRepositoryImpl{
		DB: dbPool,
	}
}

func (repository *MetricDBRepositoryImpl) Update(ctx context.Context, metricEntity entity.MetricEntity) error {
	log := logs.GetLoggerFromContext(ctx)

	query := `
		INSERT INTO metrics (name, mtype, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name, mtype) DO UPDATE SET
			delta = CASE
				WHEN EXCLUDED.mtype = 'counter' THEN COALESCE(metrics.delta, 0) + COALESCE(EXCLUDED.delta, 0)
			END,
			value = EXCLUDED.value;
	`

	err := retryPG(ctx, func() error {
		_, execErr := repository.DB.Exec(
			ctx,
			query,
			metricEntity.ID,
			metricEntity.MType,
			metricEntity.Delta,
			metricEntity.Value,
		)
		return execErr
	})

	if err != nil {
		log.Error().Err(err).Str("name", metricEntity.ID).Msg("Ошибка при сохранении метрики в PostgreSQL")
		return err
	}
	return nil
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
	err := retryPG(ctx, func() error {
		return repository.DB.QueryRow(ctx, query, metricName).Scan(
			&result.ID,
			&result.MType,
			&result.Delta,
			&result.Value,
		)
	})

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

	var result map[string]entity.MetricEntity
	err := retryPG(ctx, func() error {
		rows, queryErr := repository.DB.Query(ctx, query)
		if queryErr != nil {
			return queryErr
		}
		defer rows.Close()

		temp := make(map[string]entity.MetricEntity)
		for rows.Next() {
			var m entity.MetricEntity
			if scanErr := rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value); scanErr != nil {
				log.Error().Err(scanErr).Msg("Ошибка при сканировании строки метрики из PostgreSQL")
				continue
			}
			temp[m.ID] = m
		}

		if rowsErr := rows.Err(); rowsErr != nil {
			return rowsErr
		}

		result = temp
		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Ошибка при получении всех метрик из PostgreSQL")
		return make(map[string]entity.MetricEntity)
	}

	return result
}

func (repository *MetricDBRepositoryImpl) Clean(ctx context.Context) {
	log := logs.GetLoggerFromContext(ctx)
	log.Debug().Msg("Очищаем таблицу метрик в PostgreSQL")

	err := retryPG(ctx, func() error {
		_, execErr := repository.DB.Exec(ctx, `TRUNCATE TABLE metrics RESTART IDENTITY;`)
		return execErr
	})

	if err != nil {
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

func (repository *MetricDBRepositoryImpl) UpdateBatch(ctx context.Context, entities []entity.MetricEntity) error {
	if len(entities) == 0 {
		return nil
	}

	return retryPG(ctx, func() error {
		tx, err := repository.DB.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		query := `
			INSERT INTO metrics (name, mtype, delta, value)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (name, mtype) DO UPDATE SET
				delta = CASE
					WHEN EXCLUDED.mtype = 'counter' THEN COALESCE(metrics.delta, 0) + COALESCE(EXCLUDED.delta, 0)
				END,
				value = EXCLUDED.value;
		`

		batch := &pgx.Batch{}
		for _, m := range entities {
			batch.Queue(query, m.ID, m.MType, m.Delta, m.Value)
		}

		br := tx.SendBatch(ctx, batch)
		if err := br.Close(); err != nil {
			return err
		}

		return tx.Commit(ctx)
	})
}

func retryPG(ctx context.Context, fn func() error) error {
	err := fn()
	if err == nil || !IsRetriablePGError(err) {
		return err
	}

	log := logs.GetLoggerFromContext(ctx)
	for attempt, delay := range pgRetryDelays {
		log.Warn().
			Err(err).
			Int("attempt", attempt+1).
			Dur("delay", delay).
			Msg("Retriable-ошибка PostgreSQL, повторяем запрос...")

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		err = fn()
		if err == nil || !IsRetriablePGError(err) {
			return err
		}
	}

	return err
}

func IsRetriablePGError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure,
			pgerrcode.SQLClientUnableToEstablishSQLConnection,
			pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
			pgerrcode.TransactionResolutionUnknown,
			pgerrcode.ProtocolViolation:
			return true
		}
		if len(pgErr.Code) >= 2 && pgErr.Code[:2] == "08" {
			return true
		}
		return false
	}

	var connectErr *pgconn.ConnectError
	var netErr net.Error
	if errors.As(err, &connectErr) || errors.As(err, &netErr) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	return false
}
