package domain_test

import (
	"strings"
	"testing"

	"github.com/OrtemRepos/ShortURL/shortener-service/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestURL_GenerateShortURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		setup    func()
		validate func(t *testing.T, shortURL string, url *domain.URL)
	}{
		{
			name: "correct length",
			url:  "https://example.com",
			validate: func(t *testing.T, shortURL string, _ *domain.URL) {
				assert.Len(t, shortURL, 8, "length must be 8 characters")
			},
		},
		{
			name: "hex format",
			url:  "https://example.com",
			validate: func(t *testing.T, shortURL string, _ *domain.URL) {
				assert.Regexp(t, `^[0-9a-f]{8}$`, shortURL, "must contain only hex characters")
			},
		},
		{
			name: "sets the ShortURL field",
			url:  "https://example.com",
			validate: func(t *testing.T, shortURL string, u *domain.URL) {
				assert.Equal(t, shortURL, u.ShortURL, "the ShortURL field must be set")
			},
		},
		{
			name: "empty URL",
			url:  "",
			validate: func(t *testing.T, shortURL string, _ *domain.URL) {
				assert.Len(t, shortURL, 8, "should work with an empty URL")
			},
		},
		{
			name: "very long URL",
			url:  strings.Repeat("a", 10000),
			validate: func(t *testing.T, shortURL string, _ *domain.URL) {
				assert.Len(t, shortURL, 8, "should handle long URLs")
			},
		},
		{
			name: "wildcards in URL",
			url:  "https://пример.рф/путь/с/юникодом?параметр=значение#якорь",
			validate: func(t *testing.T, shortURL string, _ *domain.URL) {
				assert.Len(t, shortURL, 8, "must handle unicode characters")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			u := domain.NewURL(tt.url)
			shortURL := u.GenerateShortURL()
			tt.validate(t, shortURL, u)
		})
	}
}

func TestURL_Uniqueness(t *testing.T) {
	tests := []struct {
		name  string
		urls  []string
		check func(t *testing.T, results []string)
	}{
		{
			name: "different URLs give different abbreviations",
			urls: []string{
				"https://example.com/1",
				"https://example.com/2",
				"https://example.com/3",
			},
			check: func(t *testing.T, results []string) {
				seen := make(map[string]bool)
				for _, s := range results {
					assert.False(t, seen[s], "must have unique values")
					seen[s] = true
				}
			},
		},
		{
			name: "repeated calls produce different results",
			urls: []string{
				"https://example.com",
				"https://example.com",
				"https://example.com",
			},
			check: func(t *testing.T, results []string) {
				assert.Len(t, results, 3, "3 results were expected")
				assert.NotEqual(t, results[0], results[1], "results should be different")
				assert.NotEqual(t, results[0], results[2], "results should be different")
				assert.NotEqual(t, results[1], results[2], "results should be different")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := make([]string, 0, len(tt.urls))
			for _, url := range tt.urls {
				u := domain.NewURL(url)
				results = append(results, u.GenerateShortURL())
			}
			tt.check(t, results)
		})
	}
}

func TestNewURL(t *testing.T) {
	tests := []struct {
		name     string
		longURL  string
		validate func(t *testing.T, url *domain.URL)
	}{
		{
			name:    "standard case",
			longURL: "https://example.com",
			validate: func(t *testing.T, url *domain.URL) {
				assert.Equal(t, "https://example.com", url.OriginalURL)
				assert.Empty(t, url.ShortURL)
			},
		},
		{
			name:    "empty URL",
			longURL: "",
			validate: func(t *testing.T, url *domain.URL) {
				assert.Empty(t, url.OriginalURL)
				assert.Empty(t, url.ShortURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := domain.NewURL(tt.longURL)
			tt.validate(t, url)
		})
	}
}
