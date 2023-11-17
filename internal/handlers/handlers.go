package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/NRKA/gRPC-Server/internal/kafka"
	"github.com/NRKA/gRPC-Server/internal/repository"
	"github.com/NRKA/gRPC-Server/pkg/grpcServer"
	"github.com/NRKA/gRPC-Server/pkg/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"os"
	"testing"
	"time"
)

const topic = "TOPIC"

type articleInterface interface {
	Create(ctx context.Context, article repository.Article) (int64, error)
	GetByID(ctx context.Context, id int64) (repository.Article, error)
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, article repository.Article) error
}

type GrpcArticleHandler struct {
	repo        articleInterface
	producer    kafka.KafkaInterface
	currentTime func() time.Time
	grpcServer.UnimplementedArticleServiceServer
}

func NewGrpcArticleHandler(repo articleInterface, producer kafka.KafkaInterface) *GrpcArticleHandler {
	return &GrpcArticleHandler{
		repo:        repo,
		producer:    producer,
		currentTime: time.Now,
	}
}

func (handler *GrpcArticleHandler) SetCustomTimeFunc(timeFunc func() time.Time) {
	handler.currentTime = timeFunc
}

func DataConvertationСreate(article *grpcServer.CreateArticleRequest) repository.Article {
	return repository.Article{
		Name:   article.Name,
		Rating: article.Rating,
	}
}

func DataConvertationUpdate(article *grpcServer.UpdateArticleRequest) repository.Article {
	return repository.Article{
		ID:     article.Id,
		Name:   article.Name,
		Rating: article.Rating,
	}
}

func (handler *GrpcArticleHandler) CreateArticle(ctx context.Context, article *grpcServer.CreateArticleRequest) (*grpcServer.CreateArticleResponse, error) {
	l := logger.FromContext(ctx)
	ctx = logger.ToContext(ctx, l.With(zap.String("method", "CreateArticle")))

	span, ctx := opentracing.StartSpanFromContext(ctx, "GrpcArticleHandler: CreateArticle")
	defer span.Finish()

	articleData := DataConvertationСreate(article)
	if articleData.Name == "" || articleData.Rating < 1 {
		span.SetTag("error", true)
		span.LogFields(log.Error(fmt.Errorf(errInvalidData)))
		return &grpcServer.CreateArticleResponse{}, status.Error(codes.InvalidArgument, errInvalidData)
	}
	id, err := handler.repo.Create(ctx, articleData)
	if err != nil {
		span.SetTag("error", true)
		span.LogFields(log.Error(err))
		return &grpcServer.CreateArticleResponse{}, status.Error(codes.Internal, errArticleCreate+err.Error())
	}
	articleData.ID = id

	method, _ := grpc.Method(ctx)
	err = handler.producer.SendEvent(os.Getenv(topic), kafka.Event{
		TimeStamp:   handler.currentTime(),
		Type:        method,
		RequestBody: article.String(),
	})
	if err != nil {
		span.SetTag("error", true)
		span.LogFields(log.Error(err))
	}

	return &grpcServer.CreateArticleResponse{
		Id:     articleData.ID,
		Name:   articleData.Name,
		Rating: articleData.Rating,
	}, nil
}

func (handler *GrpcArticleHandler) GetArticle(ctx context.Context, id *grpcServer.GetArticleIDRequest) (*grpcServer.GetArticleResponse, error) {
	l := logger.FromContext(ctx)
	ctx = logger.ToContext(ctx, l.With(zap.String("method", "GetArticleByID")))

	span, ctx := opentracing.StartSpanFromContext(ctx, "GrpcArticleHandler: GetArticleByID")
	defer span.Finish()

	article, err := handler.repo.GetByID(ctx, id.Id)
	if err != nil {
		if errors.Is(err, repository.ErrArticalNotFound) {
			span.SetTag("error", true)
			span.LogFields(log.Error(err))
			return &grpcServer.GetArticleResponse{}, status.Error(codes.NotFound, err.Error())
		}
		span.SetTag("error", true)
		span.LogFields(log.Error(err))
		return &grpcServer.GetArticleResponse{}, status.Error(codes.Internal, errArticleGetById+err.Error())
	}
	method, _ := grpc.Method(ctx)
	err = handler.producer.SendEvent(os.Getenv(topic), kafka.Event{
		TimeStamp:   handler.currentTime(),
		Type:        method,
		RequestBody: "",
	})
	if err != nil {
		span.SetTag("error", true)
		span.LogFields(log.Error(err))
	}

	return &grpcServer.GetArticleResponse{
		Id:     article.ID,
		Name:   article.Name,
		Rating: article.Rating,
	}, nil
}

func (handler *GrpcArticleHandler) DeleteArticle(ctx context.Context, id *grpcServer.DeleteArticleIDRequest) (*emptypb.Empty, error) {
	l := logger.FromContext(ctx)
	ctx = logger.ToContext(ctx, l.With(zap.String("method", "DeleteArticle")))

	span, ctx := opentracing.StartSpanFromContext(ctx, "GrpcArticleHandler: DeleteArticle")
	defer span.Finish()

	err := handler.repo.Delete(ctx, id.Id)
	if err != nil {
		if errors.Is(err, repository.ErrArticalNotFound) {
			span.SetTag("error", true)
			span.LogFields(log.Error(err))
			return nil, status.Error(codes.NotFound, err.Error())
		}
		span.SetTag("error", true)
		span.LogFields(log.Error(err))
		return nil, status.Error(codes.Internal, errArticleDelete+err.Error())
	}

	method, _ := grpc.Method(ctx)
	err = handler.producer.SendEvent(os.Getenv(topic), kafka.Event{
		TimeStamp:   handler.currentTime(),
		Type:        method,
		RequestBody: "",
	})
	if err != nil {
		span.SetTag("error", true)
		span.LogFields(log.Error(err))
	}

	return new(emptypb.Empty), nil
}

func (handler *GrpcArticleHandler) UpdateArticle(ctx context.Context, article *grpcServer.UpdateArticleRequest) (*emptypb.Empty, error) {
	l := logger.FromContext(ctx)
	ctx = logger.ToContext(ctx, l.With(zap.String("method", "UpdateArticle")))

	span, ctx := opentracing.StartSpanFromContext(ctx, "GrpcArticleHandler: UpdateArticle")
	defer span.Finish()

	articleData := DataConvertationUpdate(article)
	if article.Name == "" || article.Rating < 1 {
		span.SetTag("error", true)
		span.LogFields(log.Error(fmt.Errorf(errInvalidData)))
		return nil, status.Error(codes.InvalidArgument, errInvalidData)
	}

	err := handler.repo.Update(ctx, articleData)
	if err != nil {
		if errors.Is(err, repository.ErrArticalNotFound) {
			span.SetTag("error", true)
			span.LogFields(log.Error(err))
			return nil, status.Error(codes.NotFound, err.Error())
		}
		span.SetTag("error", true)
		span.LogFields(log.Error(err))
		return nil, status.Error(codes.Internal, errArticleUpdate+err.Error())
	}

	method, _ := grpc.Method(ctx)
	err = handler.producer.SendEvent(os.Getenv(topic), kafka.Event{
		TimeStamp:   handler.currentTime(),
		Type:        method,
		RequestBody: article.String(),
	})
	if err != nil {
		span.SetTag("error", true)
		span.LogFields(log.Error(err))
	}

	return new(emptypb.Empty), nil
}

func setupGRPCConnection(t *testing.T, server *grpc.Server) (*grpc.ClientConn, func()) {
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	go server.Serve(listener)

	conn, err := grpc.Dial(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	closeConnAndServer := func() {
		conn.Close()
		server.Stop()
	}

	return conn, closeConnAndServer
}
