package main

import (
	"context"
	"fmt"
	"net"

	"github.com/OrtemRepos/ShortURL/shortener-service/config"
	"github.com/OrtemRepos/ShortURL/shortener-service/gen/url"
	"github.com/OrtemRepos/ShortURL/shortener-service/internal/controller"
	grpcHandler "github.com/OrtemRepos/ShortURL/shortener-service/internal/handler/grpc"
	"github.com/OrtemRepos/ShortURL/shortener-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		logger.Fatal("config reading error", zap.Error(err))
	}

	client := redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
			PoolSize: cfg.Redis.PoolSize,
		},
	)

	ping := client.Ping(context.Background())
	if ping.Err() != nil {
		logger.Error("redis ping error", zap.Error(err))
	}

	repo := repository.NewRedisURLRepo(client, logger.Named("repo_redis"))
	ctrl := controller.NewController(repo, logger.Named("controller"))
	handler := grpcHandler.New(
		ctrl,
		logger.Named("grpc_handler"),
	)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.App.Port))
	if err != nil {
		panic(err)
	}

	srv := grpc.NewServer()

	url.RegisterShortenerServiceServer(srv, handler)

	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
