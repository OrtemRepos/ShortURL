package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strconv"

	"go.uber.org/zap"
)

const maxInt64 = 1<<63 - 1

type URL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (u *URL) GenerateShortURL() string {
	randomInt, err := rand.Int(rand.Reader, big.NewInt(maxInt64))
	if err != nil {
		zap.L().Fatal("random number generation error", zap.Error(err))
	}

	randomStr := strconv.FormatInt(randomInt.Int64(), 10)
	hash := sha256.Sum256([]byte(u.OriginalURL + randomStr))
	u.ShortURL = hex.EncodeToString(hash[:])[:8]
	return u.ShortURL
}

func NewURL(longURL string) *URL {
	return &URL{
		OriginalURL: longURL,
		ShortURL:    "",
	}
}
