package handlers

import (
	"context"
	"fmt"
	"github.com/NRKA/gRPC-Server/internal/kafka"
	mock_kafka_interface "github.com/NRKA/gRPC-Server/internal/kafka/mocks"
	"github.com/NRKA/gRPC-Server/internal/repository"
	mock_repository "github.com/NRKA/gRPC-Server/internal/repository/mocks"
	"github.com/NRKA/gRPC-Server/pkg/grpcServer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"testing"
	"time"
)

func TestArticleHandler_Create(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		request          grpcServer.CreateArticleRequest
		mockReturnValue  int64
		mockError        error
		expectedCode     codes.Code
		expectedResponse *grpcServer.CreateArticleResponse
		mockKafka        func(*gomock.Controller, kafka.Event) kafka.KafkaInterface
	}{{
		name:             "success",
		request:          grpcServer.CreateArticleRequest{Name: "name", Rating: 10},
		mockReturnValue:  1,
		mockError:        nil,
		expectedCode:     codes.OK,
		expectedResponse: &grpcServer.CreateArticleResponse{Id: 1, Name: "name", Rating: 10},
		mockKafka: func(controller *gomock.Controller, event kafka.Event) kafka.KafkaInterface {
			mockProducer := mock_kafka_interface.NewMockKafkaInterface(controller)
			mockProducer.EXPECT().SendEvent(os.Getenv(topic), event).Return(nil)
			return mockProducer
		},
	},
		{
			name:             "internal server error",
			request:          grpcServer.CreateArticleRequest{Name: "name", Rating: 10},
			mockReturnValue:  1,
			mockError:        fmt.Errorf("failed to create article: internal server error"),
			expectedCode:     codes.Internal,
			expectedResponse: &grpcServer.CreateArticleResponse{},
			mockKafka: func(controller *gomock.Controller, event kafka.Event) kafka.KafkaInterface {
				return mock_kafka_interface.NewMockKafkaInterface(controller)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			ctrl := gomock.NewController(t)
			mockRepo := mock_repository.NewMockArticleInterface(ctrl)
			mockKafka := tc.mockKafka(ctrl, kafka.Event{
				TimeStamp:   time.Date(2023, 10, 22, 22, 22, 22, 22, time.Local),
				Type:        "/ArticleService/CreateArticle",
				RequestBody: tc.request.String(),
			})

			server := grpc.NewServer()
			handler := NewGrpcArticleHandler(mockRepo, mockKafka)
			grpcServer.RegisterArticleServiceServer(server, handler)
			mockRepo.EXPECT().Create(gomock.Any(), repository.Article{
				Name:   tc.request.Name,
				Rating: tc.request.Rating,
			}).Return(tc.mockReturnValue, tc.mockError)
			defer ctrl.Finish()

			handler.SetCustomTimeFunc(func() time.Time {
				return time.Date(2023, 10, 22, 22, 22, 22, 22, time.Local)
			})

			conn, closeConnAndServer := setupGRPCConnection(t, server)
			defer closeConnAndServer()

			client := grpcServer.NewArticleServiceClient(conn)
			response, err := client.CreateArticle(context.Background(), &tc.request)

			if err != nil {
				st, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, st.Code())
				return
			}
			assert.Equal(t, tc.expectedResponse.Id, response.Id)
			assert.Equal(t, tc.expectedResponse.Name, response.Name)
			assert.Equal(t, tc.expectedResponse.Rating, response.Rating)
		})
	}
}

func TestArticleHandler_GetByID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		request           grpcServer.GetArticleIDRequest
		mockReturnArticle repository.Article
		mockError         error
		expectedCode      codes.Code
		expectedResponse  *grpcServer.GetArticleResponse
		mockKafka         func(*gomock.Controller, kafka.Event) kafka.KafkaInterface
	}{{
		name:              "success",
		request:           grpcServer.GetArticleIDRequest{Id: 1},
		mockReturnArticle: repository.Article{ID: 1, Name: "name", Rating: 10},
		mockError:         nil,
		expectedCode:      codes.OK,
		expectedResponse:  &grpcServer.GetArticleResponse{Id: 1, Name: "name", Rating: 10},
		mockKafka: func(controller *gomock.Controller, event kafka.Event) kafka.KafkaInterface {
			mockProducer := mock_kafka_interface.NewMockKafkaInterface(controller)
			mockProducer.EXPECT().SendEvent(os.Getenv(topic), event).Return(nil)
			return mockProducer
		}}, {
		name:              "article not found",
		request:           grpcServer.GetArticleIDRequest{Id: 99},
		mockReturnArticle: repository.Article{},
		mockError:         repository.ErrArticalNotFound,
		expectedCode:      codes.NotFound,
		expectedResponse:  &grpcServer.GetArticleResponse{},
		mockKafka: func(controller *gomock.Controller, event kafka.Event) kafka.KafkaInterface {
			return mock_kafka_interface.NewMockKafkaInterface(controller)
		}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := mock_repository.NewMockArticleInterface(ctrl)
			mockKafka := tc.mockKafka(ctrl, kafka.Event{
				TimeStamp:   time.Date(2023, 10, 22, 22, 22, 22, 22, time.Local),
				Type:        "/ArticleService/GetArticle",
				RequestBody: "",
			})

			server := grpc.NewServer()
			handler := NewGrpcArticleHandler(mockRepo, mockKafka)
			grpcServer.RegisterArticleServiceServer(server, handler)
			mockRepo.EXPECT().GetByID(gomock.Any(), tc.request.Id).Return(tc.mockReturnArticle, tc.mockError)

			handler.SetCustomTimeFunc(func() time.Time {
				return time.Date(2023, 10, 22, 22, 22, 22, 22, time.Local)
			})

			conn, closeConnAndServer := setupGRPCConnection(t, server)
			defer closeConnAndServer()

			client := grpcServer.NewArticleServiceClient(conn)
			response, err := client.GetArticle(context.Background(), &tc.request)

			if err != nil {
				st, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, st.Code())
				return
			}
			assert.Equal(t, tc.expectedResponse.Id, response.Id)
			assert.Equal(t, tc.expectedResponse.Name, response.Name)
			assert.Equal(t, tc.expectedResponse.Rating, response.Rating)
		})
	}
}

func TestArticleHandler_Delete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		request      grpcServer.DeleteArticleIDRequest
		mockError    error
		expectedCode codes.Code
		mockKafka    func(*gomock.Controller, kafka.Event) kafka.KafkaInterface
	}{{
		name:         "success",
		request:      grpcServer.DeleteArticleIDRequest{Id: 1},
		mockError:    nil,
		expectedCode: codes.OK,
		mockKafka: func(controller *gomock.Controller, event kafka.Event) kafka.KafkaInterface {
			mockProducer := mock_kafka_interface.NewMockKafkaInterface(controller)
			mockProducer.EXPECT().SendEvent(os.Getenv(topic), event).Return(nil)
			return mockProducer
		},
	}, {
		name:         "article not found",
		request:      grpcServer.DeleteArticleIDRequest{Id: 9999},
		mockError:    repository.ErrArticalNotFound,
		expectedCode: codes.NotFound,
		mockKafka: func(controller *gomock.Controller, event kafka.Event) kafka.KafkaInterface {
			return mock_kafka_interface.NewMockKafkaInterface(controller)
		},
	}, {
		name:         "internal server error",
		request:      grpcServer.DeleteArticleIDRequest{Id: 9999},
		mockError:    fmt.Errorf("failed to delete article"),
		expectedCode: codes.Internal,
		mockKafka: func(controller *gomock.Controller, event kafka.Event) kafka.KafkaInterface {
			return mock_kafka_interface.NewMockKafkaInterface(controller)
		},
	},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := mock_repository.NewMockArticleInterface(ctrl)
			mockKafka := tc.mockKafka(ctrl, kafka.Event{
				TimeStamp:   time.Date(2023, 10, 22, 22, 22, 22, 22, time.Local),
				Type:        "/ArticleService/DeleteArticle",
				RequestBody: "",
			})
			server := grpc.NewServer()
			handler := NewGrpcArticleHandler(mockRepo, mockKafka)
			grpcServer.RegisterArticleServiceServer(server, handler)
			mockRepo.EXPECT().Delete(gomock.Any(), tc.request.Id).Return(tc.mockError)

			handler.SetCustomTimeFunc(func() time.Time {
				return time.Date(2023, 10, 22, 22, 22, 22, 22, time.Local)
			})

			conn, closeConnAndServer := setupGRPCConnection(t, server)
			defer closeConnAndServer()

			client := grpcServer.NewArticleServiceClient(conn)
			_, err := client.DeleteArticle(context.Background(), &tc.request)

			if err != nil {
				st, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, st.Code())
			}
		})
	}
}

func TestArticleHandler_Update(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		request      grpcServer.UpdateArticleRequest
		mockError    error
		expectedCode codes.Code
		mockKafka    func(*gomock.Controller, kafka.Event) kafka.KafkaInterface
	}{{
		name:         "success",
		request:      grpcServer.UpdateArticleRequest{Id: 123, Name: "name", Rating: 10},
		mockError:    nil,
		expectedCode: codes.OK,
		mockKafka: func(controller *gomock.Controller, event kafka.Event) kafka.KafkaInterface {
			mockProducer := mock_kafka_interface.NewMockKafkaInterface(controller)
			mockProducer.EXPECT().SendEvent(os.Getenv(topic), event).Return(nil)
			return mockProducer
		},
	}, {
		name:         "article not found",
		request:      grpcServer.UpdateArticleRequest{Id: 123123, Name: "name", Rating: 10},
		mockError:    repository.ErrArticalNotFound,
		expectedCode: codes.NotFound,
		mockKafka: func(controller *gomock.Controller, event kafka.Event) kafka.KafkaInterface {
			return mock_kafka_interface.NewMockKafkaInterface(controller)
		},
	},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			ctrl := gomock.NewController(t)
			mockRepo := mock_repository.NewMockArticleInterface(ctrl)
			mockKafka := tc.mockKafka(ctrl, kafka.Event{
				TimeStamp:   time.Date(2023, 10, 22, 22, 22, 22, 22, time.Local),
				Type:        "/ArticleService/UpdateArticle",
				RequestBody: tc.request.String(),
			})
			server := grpc.NewServer()
			handler := NewGrpcArticleHandler(mockRepo, mockKafka)
			grpcServer.RegisterArticleServiceServer(server, handler)
			mockRepo.EXPECT().Update(gomock.Any(), repository.Article{
				ID:     tc.request.Id,
				Name:   tc.request.Name,
				Rating: tc.request.Rating,
			}).Return(tc.mockError)
			defer ctrl.Finish()

			handler.SetCustomTimeFunc(func() time.Time {
				return time.Date(2023, 10, 22, 22, 22, 22, 22, time.Local)
			})

			conn, closeConnAndServer := setupGRPCConnection(t, server)
			defer closeConnAndServer()

			client := grpcServer.NewArticleServiceClient(conn)
			_, err := client.UpdateArticle(context.Background(), &tc.request)
			if err != nil {
				st, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, st.Code())
			}
		})
	}
}
