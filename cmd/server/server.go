package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/config/server"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/handler"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/service"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/static"
	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/utils"
	"github.com/caarlos0/env/v11"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/rs/zerolog/log"
)

func main() {
	log.Printf("Запуск сервера для сбора рантайм-метрик...")
	run()
}

// 1. АДАПТЕР ДЛЯ ЛОГГЕРА
type internalLoggingWriter struct {
	gin.ResponseWriter
	responseData *utils.ResponseData
}

func (w *internalLoggingWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.responseData.Status = statusCode
}

func (w *internalLoggingWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.responseData.Size += n
	return n, err
}

func (w *internalLoggingWriter) WriteString(s string) (int, error) {
	n, err := w.ResponseWriter.WriteString(s)
	w.responseData.Size += n
	return n, err
}

func LoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		reqID := uuid.New().String()
		logger := log.With().Str("requestId", reqID).Logger()
		reqCtx := logger.WithContext(ctx.Request.Context())
		ctx.Request = ctx.Request.WithContext(reqCtx)

		responseData := &utils.ResponseData{
			Status: http.StatusOK,
			Size:   0,
		}

		lw := &internalLoggingWriter{
			ResponseWriter: ctx.Writer,
			responseData:   responseData,
		}
		ctx.Writer = lw

		logger.Debug().
			Str("host", ctx.Request.Host).
			Str("method", ctx.Request.Method).
			Str("path", ctx.Request.URL.Path).
			Msg("Входящий HTTP запрос")

		ctx.Next()

		duration := time.Since(start)
		logger.Debug().
			Dur("duration", duration).
			Int("size", responseData.Size).
			Int("status", responseData.Status).
			Msg("Обработка метода завершена. Отдаем ответ")
	}
}

// 2. БУФЕРИЗИРОВАННЫЙ WRITER ДЛЯ GZIP
type gzipBufferWriter struct {
	gin.ResponseWriter
	bodyBuf    *bytes.Buffer
	statusCode int
}

func (w *gzipBufferWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *gzipBufferWriter) Write(b []byte) (int, error) {
	return w.bodyBuf.Write(b)
}

func (w *gzipBufferWriter) WriteString(s string) (int, error) {
	return w.bodyBuf.WriteString(s)
}

func (w *gzipBufferWriter) Status() int {
	if w.statusCode == 0 {
		return w.ResponseWriter.Status()
	}
	return w.statusCode
}

func CustomGzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Декомпрессия входящего контента (остается без изменений)
		if c.GetHeader("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid gzip body"})
				return
			}
			c.Request.Body = &gzipReadCloser{Reader: gzipReader, gzipReader: gzipReader, bodyReader: c.Request.Body}
		}

		// Создаем буферизированный писатель
		bufWriter := &gzipBufferWriter{
			ResponseWriter: c.Writer,
			bodyBuf:        bytes.NewBuffer(make([]byte, 0, 1024)),
			statusCode:     http.StatusOK,
		}

		// Подменяем писатель для хендлеров
		originalWriter := c.Writer
		c.Writer = bufWriter

		c.Next()

		// Возвращаем оригинальный писатель назад перед отправкой данных
		c.Writer = originalWriter

		contentType := bufWriter.Header().Get("Content-Type")
		acceptEncoding := c.GetHeader("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		// Проверяем условия сжатия, когда хендлер завершил работу и заголовки ТОЧНО проставлены
		if supportsGzip && (strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")) {
			bufWriter.Header().Set("Content-Encoding", "gzip")
			bufWriter.Header().Add("Vary", "Accept-Encoding")

			// Пишем статус код
			bufWriter.ResponseWriter.WriteHeader(bufWriter.Status())

			// Сжимаем накопленные байты из буфера в оригинальный ответ
			gz := gzip.NewWriter(bufWriter.ResponseWriter)
			_, _ = gz.Write(bufWriter.bodyBuf.Bytes())
			_ = gz.Close()
		} else {
			// Отправляем данные как есть без сжатия
			bufWriter.ResponseWriter.WriteHeader(bufWriter.Status())
			_, _ = bufWriter.ResponseWriter.Write(bufWriter.bodyBuf.Bytes())
		}
	}
}

func run() {
	conf := server.Config{LogLevel: "debug"}

	flag.StringVar(&conf.ServerURL, "a", "localhost:8080", "server url")
	flag.Parse()
	if err := env.Parse(&conf); err != nil {
		log.Printf("failed to parse env variables: %v", err)
	}

	utils.InitBaseLogger(utils.LoggerConfig{Level: conf.LogLevel})
	log.Debug().Any("Конфиг", conf).Msg("Содержимое конфига")

	gin.SetMode(gin.ReleaseMode)

	log.Info().Msg("Инициализация http-server-а")
	router := gin.New()

	router.Use(LoggingMiddleware())
	router.Use(CustomGzipMiddleware())
	router.Use(gin.Recovery())

	fs := static.SetUpFs()
	router.SetHTMLTemplate(fs)

	repoImpl := repository.NewMetricRepositoryImpl()
	serviceImpl := service.NewMetricServiceImpl(repoImpl)

	handler.SetUpRoutes(router, serviceImpl)

	log.Info().Msg("Инициализация http-server-а выполнена успешно")

	err := router.Run(conf.ServerURL)
	if err != nil {
		log.Info().Any("error", err.Error()).Msg("Ошибка во время запуска http-сервера на порту ")
	}
}

type gzipReadCloser struct {
	io.Reader
	gzipReader *gzip.Reader
	bodyReader io.ReadCloser
}

func (g *gzipReadCloser) Close() error {
	errGz := g.gzipReader.Close()
	errBody := g.bodyReader.Close()
	if errGz != nil {
		return errGz
	}
	return errBody
}
