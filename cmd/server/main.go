package main

import (
	"context"
	"errors"
	"github.com/NRKA/gRPC-Server/internal/db"
	"github.com/NRKA/gRPC-Server/internal/handlers"
	"github.com/NRKA/gRPC-Server/internal/kafka"
	"github.com/NRKA/gRPC-Server/internal/repository/postgresql"
	"github.com/NRKA/gRPC-Server/pkg/grpcServer"
	"github.com/NRKA/gRPC-Server/pkg/logger"
	"github.com/joho/godotenv"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

const (
	port       = "PORT"
	dbHost     = "DB_HOST"
	dbPort     = "DB_PORT"
	dbUser     = "DB_USER"
	dbPassword = "DB_PASSWORD"
	dbName     = "DB_NAME"
	brokerAddr = "BROKER_ADDRESS"
	topic      = "TOPIC"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	brokerAddress := os.Getenv(brokerAddr)

	producer, err := kafka.NewKafkaProducer(brokerAddress)
	if err != nil {
		logger.Fatalf(ctx, "failed to create producer: %v", err)
	}
	defer func() {
		err := producer.Close()
		if err != nil {
			logger.Errorf(ctx, "Failed to close producer: %v", err)
		}
	}()

	consumer, err := kafka.NewKafkaConsumer(brokerAddress)
	if err != nil {
		logger.Fatalf(ctx, "failed to create consumer: %v", err)
	}
	defer func() {
		err := consumer.Close()
		if err != nil {
			logger.Errorf(ctx, "Failed to close consumer: %v", err)
		}
	}()

	port := os.Getenv(port)

	dbConfig := db.DatabaseConfig{
		Host:     os.Getenv(dbHost),
		Port:     os.Getenv(dbPort),
		User:     os.Getenv(dbUser),
		Password: os.Getenv(dbPassword),
		DBName:   os.Getenv(dbName),
	}
	database, err := db.NewDB(ctx, dbConfig)
	if err != nil {
		logger.Fatalf(ctx, "Failed to connect to database: %v", err)
	}
	defer database.GetPool().Close()
	articleRepo := postgresql.NewArticleRepo(database)

	zapLogger, err := zap.NewProduction()
	if err != nil {
		logger.Fatalf(ctx, "Error creating zap logger: %v", err)
	}

	logger.SetGlobal(
		zapLogger.With(zap.String("component", "server")),
	)

	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            false,
			BufferFlushInterval: 1 * time.Second,
		},
	}
	tracer, closer, err := cfg.New(
		"messages-service",
	)
	if err != nil {
		logger.Fatalf(ctx, "cannot create tracer: %v\n", err)
	}
	defer closer.Close()

	opentracing.SetGlobalTracer(tracer)
	server := grpc.NewServer()

	service := handlers.NewGrpcArticleHandler(articleRepo, producer)
	grpcServer.RegisterArticleServiceServer(server, service)

	go func() {
		err := consumer.Consume(ctx, os.Getenv(topic))
		if err != nil {
			logger.Errorf(ctx, "failed to consume: %v", err)
			return
		}

		server.Stop()
		return
	}()

	logger.Infof(ctx, "server listening on %q", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Fatalf(ctx, "failed to create listener: %v", err)
	}
	err = server.Serve(lis)
	if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		logger.Fatalf(ctx, "failed to serve: %v", err)
	}
}
