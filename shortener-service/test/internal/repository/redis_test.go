package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/OrtemRepos/shortener-service/internal/domain"
	"github.com/OrtemRepos/shortener-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestRedisURLRepo_Save(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	shortURL := "abc123"
	originalURL := "https://example.com"
	expTime := 10 * time.Minute

	tests := []struct {
		name        string
		input       *domain.URL
		expTime     time.Duration
		setupMock   bool
		mockSetup   func(mock redismock.ClientMock)
		expectedErr error
	}{
		{
			name: "successful save",
			input: &domain.URL{
				ShortURL:    shortURL,
				OriginalURL: originalURL,
			},
			expTime: expTime,
			setupMock: true,
			mockSetup: func(mock redismock.ClientMock) {
				mock.ExpectSet(shortURL, originalURL, expTime).SetVal("OK")
			},
		},
		{
			name:        "nil URL",
			input:       nil,
			expTime:     expTime,
			setupMock:   false,
			mockSetup:   func(mock redismock.ClientMock) {},
			expectedErr: repository.ErrURLNil,
		},
		{
			name: "Redis error",
			input: &domain.URL{
				ShortURL:    shortURL,
				OriginalURL: originalURL,
			},
			expTime: expTime,
			setupMock: true,
			mockSetup: func(mock redismock.ClientMock) {
				mock.ExpectSet(shortURL, originalURL, expTime).SetErr(redis.ErrClosed)
			},
			expectedErr: redis.ErrClosed,
		},
		{
			name: "empty shortURL",
			input: &domain.URL{
				ShortURL:    "",
				OriginalURL: originalURL,
			},
			expTime:     expTime,
			setupMock:   false,
			mockSetup:   func(mock redismock.ClientMock) {},
			expectedErr: repository.ErrShortURLEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := redismock.NewClientMock()
			repo := repository.NewRedisURLRepo(db, logger)

			if tt.setupMock {
				tt.mockSetup(mock)
			}

			err := repo.Save(ctx, tt.input, tt.expTime)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr, "unexpected error")
			} else {
				assert.NoError(t, err)
			}

			if tt.setupMock {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

func TestRedisURLRepo_Get(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	shortURL := "abc123"
	originalURL := "https://example.com"

	tests := []struct {
		name          string
		input         string
		mockSetup     func(mock redismock.ClientMock)
		expectedURL   *domain.URL
		expectedError error
	}{
		{
			name:  "successful receipt",
			input: shortURL,
			mockSetup: func(mock redismock.ClientMock) {
				mock.ExpectGet(shortURL).SetVal(originalURL)
			},
			expectedURL: &domain.URL{
				ShortURL:    shortURL,
				OriginalURL: originalURL,
			},
		},
		{
			name:  "URL not found",
			input: shortURL,
			mockSetup: func(mock redismock.ClientMock) {
				mock.ExpectGet(shortURL).RedisNil()
			},
			expectedError: repository.ErrURLNotFound,
		},
		{
			name:  "Redis error",
			input: shortURL,
			mockSetup: func(mock redismock.ClientMock) {
				mock.ExpectGet(shortURL).SetErr(redis.ErrClosed)
			},
			expectedError: redis.ErrClosed,
		},
		{
			name:          "empty shortURL",
			input:         "",
			mockSetup:     func(mock redismock.ClientMock) {},
			expectedError: repository.ErrShortURLEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := redismock.NewClientMock()
			repo := repository.NewRedisURLRepo(db, logger)

			tt.mockSetup(mock)

			result, err := repo.Get(ctx, tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRedisURLRepo_Delete(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	shortURL := "abc123"

	tests := []struct {
		name          string
		input         string
		mockSetup     func(mock redismock.ClientMock)
		expectedError error
	}{
		{
			name:  "successful removal",
			input: shortURL,
			mockSetup: func(mock redismock.ClientMock) {
				mock.ExpectDel(shortURL).SetVal(1)
			},
		},
		{
			name:  "URL not found",
			input: shortURL,
			mockSetup: func(mock redismock.ClientMock) {
				mock.ExpectDel(shortURL).SetVal(0)
			},
		},
		{
			name:  "Redis error",
			input: shortURL,
			mockSetup: func(mock redismock.ClientMock) {
				mock.ExpectDel(shortURL).SetErr(redis.ErrClosed)
			},
			expectedError: redis.ErrClosed,
		},
		{
			name:          "empty shortURL",
			input:         "",
			mockSetup:     func(mock redismock.ClientMock) {},
			expectedError: repository.ErrShortURLEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := redismock.NewClientMock()
			repo := repository.NewRedisURLRepo(db, logger)

			tt.mockSetup(mock)

			err := repo.Delete(ctx, tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}