package repository

import (
	"context"
	"errors"
	"time"

	"github.com/OrtemRepos/shortener-service/internal/domain"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrURLNotFound = errors.New("url not found")
)

type RedisURLRepo struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRedisURLRepo(client *redis.Client, logger *zap.Logger) *RedisURLRepo {
	return &RedisURLRepo{
		client: client,
		logger: logger,
	}
}

func (r *RedisURLRepo) Save(ctx context.Context, url *domain.URL, expTime time.Duration) error {
	if url == nil {
		return errors.New("url cannot be nil")
	}

	if err := r.client.Set(ctx, url.ShortURL, url.OriginalURL, expTime).Err(); err != nil {
		r.logger.Error("failed to save url",
			zap.String("short_url", url.ShortURL),
			zap.Error(err))
		return err
	}

	r.logger.Debug("url saved successfully",
		zap.String("short_url", url.ShortURL))
	return nil
}

func (r *RedisURLRepo) Get(ctx context.Context, shortURL string) (*domain.URL, error) {
	if shortURL == "" {
		return nil, errors.New("shortURL cannot be empty")
	}

	originalURL, err := r.client.Get(ctx, shortURL).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			r.logger.Warn("url not found",
				zap.String("short_url", shortURL))
			return nil, ErrURLNotFound
		}

		r.logger.Error("failed to get url",
			zap.String("short_url", shortURL),
			zap.Error(err))
		return nil, err
	}

	r.logger.Debug("url retrieved successfully",
		zap.String("short_url", shortURL))
	return &domain.URL{
		OriginalURL: originalURL,
		ShortURL:    shortURL,
	}, nil
}

func (r *RedisURLRepo) Delete(ctx context.Context, shortURL string) error {
	if shortURL == "" {
		return errors.New("shortURL cannot be empty")
	}

	if err := r.client.Del(ctx, shortURL).Err(); err != nil {
		r.logger.Error("failed to delete url",
			zap.String("short_url", shortURL),
			zap.Error(err))
		return err
	}

	r.logger.Debug("url deleted successfully",
		zap.String("short_url", shortURL))
	return nil
}
