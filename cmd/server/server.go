package main

import (
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

// Внутренний враппер для логгера, который гарантирует правильный перехват WriteString для Gin
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

// LoggingMiddleware теперь корректно работает в связке с gzip
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

type gzipResponseWriter struct {
	gin.ResponseWriter
	gzWriter           *gzip.Writer
	shouldCompress     bool
	headerWritten      bool
	clientSupportsGzip bool
}

func (g *gzipResponseWriter) checkContentType() {
	contentType := g.Header().Get("Content-Type")
	if g.clientSupportsGzip && (strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")) {
		g.shouldCompress = true
		g.Header().Set("Content-Encoding", "gzip")
		g.Header().Add("Vary", "Accept-Encoding")

		if g.gzWriter == nil {
			g.gzWriter = gzip.NewWriter(g.ResponseWriter)
		}
	}
}

func (g *gzipResponseWriter) Close() error {
	if g.shouldCompress && g.gzWriter != nil {
		return g.gzWriter.Close()
	}
	return nil
}

func (g *gzipResponseWriter) WriteHeader(statusCode int) {
	if g.headerWritten {
		return
	}

	g.checkContentType()
	g.ResponseWriter.WriteHeader(statusCode)
	g.headerWritten = true
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if !g.headerWritten {
		g.checkContentType()
		status := g.ResponseWriter.Status()
		if status == 0 {
			status = http.StatusOK
		}
		g.ResponseWriter.WriteHeader(status)
		g.headerWritten = true
	}

	if g.shouldCompress && g.gzWriter != nil {
		return g.gzWriter.Write(b)
	}
	return g.ResponseWriter.Write(b)
}

func (g *gzipResponseWriter) WriteString(s string) (int, error) {
	return g.Write([]byte(s))
}

func CustomGzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid gzip body"})
				return
			}
			c.Request.Body = &gzipReadCloser{Reader: gzipReader, gzipReader: gzipReader, bodyReader: c.Request.Body}
		}

		acceptEncoding := c.GetHeader("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		customWriter := &gzipResponseWriter{
			ResponseWriter:     c.Writer,
			gzWriter:           nil,
			clientSupportsGzip: supportsGzip,
		}

		c.Writer = customWriter
		c.Next()
		customWriter.Close()
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

	// Включаем оба middleware в строгом порядке работы цепочки
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
