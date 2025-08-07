package main

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/OrtemRepos/ShortURL/shortener-service/config"
	"github.com/OrtemRepos/ShortURL/shortener-service/gen/url"
	"github.com/OrtemRepos/ShortURL/shortener-service/internal/controller"
	grpcHandler "github.com/OrtemRepos/ShortURL/shortener-service/internal/handler/grpc"
	"github.com/OrtemRepos/ShortURL/shortener-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
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

	writer := &kafka.Writer{
		Addr: kafka.TCP(cfg.Kafka.Brokers...),
		Topic: cfg.Kafka.Topic,
		Balancer: &kafka.LeastBytes{},
		WriteTimeout: cfg.Kafka.WriteTimeout,
		RequiredAcks: kafka.RequiredAcks(cfg.Kafka.RequiredAcks),
		BatchSize: cfg.Kafka.BatchSize,
		BatchBytes: cfg.Kafka.BatchBytes,
		BatchTimeout: cfg.Kafka.BatchTimeout,
		MaxAttempts: cfg.Kafka.MaxAttempts,
		AllowAutoTopicCreation: true,
		Async: true,
	}

	defer func() {
		if err := writer.Close(); err != nil {
			logger.Fatal("failed to close writer", zap.Error(err))
		}
	}()
	
	
	ctrl := controller.NewController(repo, writer, logger.Named("controller"))
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
	
	logger.Info("service started", zap.Any("config", cfg))
	
	if err := ensureTopicExists(context.Background(), writer, cfg.Kafka.Topic, logger); err != nil {
		logger.Fatal("failed to ensure topics exists", zap.Error(err))
	}

	if err := srv.Serve(lis); err != nil {
		panic(err)
	}

}

func ensureTopicExists(ctx context.Context, writer *kafka.Writer, topic string, logger *zap.Logger) error {
	conn, err := kafka.DialContext(ctx, "tcp", strings.Split(writer.Addr.String(), ",")[0])
	if err != nil {
		return fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions(topic)
	if err != nil || len(partitions) == 0 {
		logger.Info("attempting to create topic", zap.String("topic", topic))
		controllerConn, err := kafka.Dial("tcp", strings.Split(writer.Addr.String(), ",")[0])
		if err != nil {
			return fmt.Errorf("failed to dial controller: %w", err)
		}
		defer controllerConn.Close()

		topicConfigs := []kafka.TopicConfig{
			{
				Topic:             topic,
				NumPartitions:     3,
				ReplicationFactor: 2,
			},
		}
		return controllerConn.CreateTopics(topicConfigs...)
	}
	return nil
}