package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App   AppConfig
	Redis RedisConfig
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

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	viper.SetDefault("app.port", 8080)
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	_ = viper.BindEnv("app.name", "APP_NAME")
	_ = viper.BindEnv("app.env", "APP_ENV")
	_ = viper.BindEnv("app.port", "APP_PORT")
	_ = viper.BindEnv("redis.host", "APP_REDIS_HOST")
	_ = viper.BindEnv("redis.port", "APP_REDIS_PORT")
	_ = viper.BindEnv("redis.password", "APP_REDIS_PASSWORD")
	_ = viper.BindEnv("redis.db", "APP_REDIS_DB")
	_ = viper.BindEnv("redis.pool_size", "APP_REDIS_POOL_SIZE")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
