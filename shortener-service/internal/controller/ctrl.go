package controller

import (
	"context"
	"encoding/json"
	"time"

	"github.com/OrtemRepos/ShortURL/shortener-service/internal/domain"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type URLRepository interface {
	Save(ctx context.Context, url *domain.URL, expTime time.Duration) error
	Get(ctx context.Context, shortURL string) (*domain.URL, error)
	Delete(ctx context.Context, shortURL string) error
}

type Controller struct {
	logger *zap.Logger
	repo   URLRepository
	writer *kafka.Writer
}

func NewController(repo URLRepository, writer *kafka.Writer, logger *zap.Logger) *Controller {
	return &Controller{
		repo:   repo,
		writer: writer,
		logger: logger,
	}
}

func (ctrl *Controller) Save(ctx context.Context, url *domain.URL, expTime time.Duration) error {
	if url.ShortURL == "" {
		url.ShortURL = url.GenerateShortURL()
		ctrl.logger.Debug(
			"short url generated",
			zap.String("short_url", url.ShortURL),
			zap.String("original_url", url.OriginalURL),
		)
	}

	if err := ctrl.repo.Save(ctx, url, expTime); err != nil {
		return err
	}
	go func() {
		msgData, _ := json.Marshal(map[string]string{
			"original_url": url.OriginalURL,
			"short_url":    url.ShortURL,
		})

		kafkaCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := ctrl.writer.WriteMessages(kafkaCtx,
			kafka.Message{
				Key:   []byte("url_created"),
				Value: msgData,
			},
		); err != nil {
			ctrl.logger.Error("kafka write failed", 
				zap.Error(err), 
				zap.String("short_url", url.ShortURL),
			)
		}
	}()

	return nil
}

func (ctrl *Controller) Delete(ctx context.Context, shortURL string) error {
	if err := ctrl.repo.Delete(ctx, shortURL); err != nil {
		return err
	}

	go func() {
		kafkaCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := ctrl.writer.WriteMessages(kafkaCtx,
			kafka.Message{
				Key:   []byte("url_deleted"),
				Value: []byte(shortURL),
			},
		); err != nil {
			ctrl.logger.Error("kafka delete event failed",
				zap.Error(err),
				zap.String("short_url", shortURL),
			)
		}
	}()

	return nil
}

func (ctrl *Controller) Get(ctx context.Context, shortURL string) (*domain.URL, error) {
	return ctrl.repo.Get(ctx, shortURL)
}
