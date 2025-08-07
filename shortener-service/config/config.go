package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App   AppConfig
	Redis RedisConfig
	Kafka KafkaConfig
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
	Port int    `mapstructure:"port"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type KafkaConfig struct {
	Brokers        []string      `mapstructure:"brokers"`
	Topic          string        `mapstructure:"topic"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	RequiredAcks   int           `mapstructure:"required_acks"`
	BatchSize      int           `mapstructure:"batch_size"`
	BatchBytes     int64         `mapstructure:"batch_bytes"`
	BatchTimeout   time.Duration `mapstructure:"batch_timeout"`
	MaxAttempts    int           `mapstructure:"max_attempts"`
	CommitInterval time.Duration `mapstructure:"commit_interval"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	viper.SetDefault("app.port", 8080)
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	
	viper.SetDefault("kafka.write_timeout", "5s")
	viper.SetDefault("kafka.required_acks", 1)
	viper.SetDefault("kafka.batch_size", 100)
	viper.SetDefault("kafka.batch_timeout", "1s")
	viper.SetDefault("kafka.max_attempts", 3)
	viper.SetDefault("kafka.commit_interval", "1s")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	bindEnvs := []string{
		"app.name", "app.env", "app.port",
		"redis.host", "redis.port", "redis.password", "redis.db", "redis.pool_size",
		"kafka.brokers", "kafka.topic", "kafka.write_timeout", "kafka.required_acks",
		"kafka.batch_size", "kafka.batch_bytes", "kafka.batch_timeout",
		"kafka.max_attempts", "kafka.commit_interval",
	}

	for _, key := range bindEnvs {
		if err := viper.BindEnv(key); err != nil {
			return nil, fmt.Errorf("failed to bind env for %s: %w", key, err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}