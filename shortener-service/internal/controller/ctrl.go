package controller

import (
	"context"
	"time"

	"github.com/OrtemRepos/shortener-service/internal/domain"
	"go.uber.org/zap"
)

type URLRepositury interface {
	Save(ctx context.Context, url *domain.URL, expTime time.Duration) error
	Get(ctx context.Context, shortURL string) (*domain.URL, error)
	Delete(ctx context.Context, shortURL string) error
}

type Controller struct {
	logger *zap.Logger
	repo   URLRepositury
}

func NewController(repo URLRepositury, logger *zap.Logger) *Controller {
	return &Controller{
		repo:   repo,
		logger: logger,
	}
}

func (ctrl *Controller) Save(ctx context.Context, url *domain.URL, expTime time.Duration) error {
	if url.ShortURL == "" {
		short := url.GenerateShortURL()
		ctrl.logger.Debug(
			"short url generated",
			zap.String("short url", short),
			zap.String("original url", url.OriginalURL),
		)
	}
	return ctrl.repo.Save(ctx, url, expTime)
}

func (ctrl *Controller) Get(ctx context.Context, shortURL string) (*domain.URL, error) {
	return ctrl.repo.Get(ctx, shortURL)
}

func (ctrl *Controller) Delete(ctx context.Context, shortURL string) error {
	return ctrl.repo.Delete(ctx, shortURL)
}
