// Code generated. DO NOT EDIT.

package user

import (
	pb "github.com/oulabla/go-base/gen/go/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Controller struct {
	pb.UnimplementedUserServiceServer
	// сюда внедряем зависимости (usecases, logger, repositories и т.д.)
	// usecase usecase.UserServiceUsecase
	// logger  *zap.Logger
}

func NewController( /* usecase usecase.UserServiceUsecase */ ) *Controller {
	return &Controller{
		// usecase: usecase,
	}
}

// Register регистрирует сервис в grpc-сервере
func (c *Controller) Register(srv *grpc.Server) {
	pb.RegisterUserServiceServer(srv, c)
	// если нужен reflection (для grpcurl и т.п.)
	reflection.Register(srv)
}
