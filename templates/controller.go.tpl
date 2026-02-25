// Code generated. DO NOT EDIT.

package ${SERVICE_PKG_NAME}

import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"

    ${PROTO_IMPORT}
)

type Controller struct {
    pb.Unimplemented${SERVICE_NAME}Server
    // сюда внедряем зависимости (usecases, logger, repositories и т.д.)
    // usecase usecase.UserServiceUsecase
    // logger  *zap.Logger
}

func NewController(/* usecase usecase.UserServiceUsecase */) *Controller {
    return &Controller{
        // usecase: usecase,
    }
}

// Register регистрирует сервис в grpc-сервере
func (c *Controller) Register(srv *grpc.Server) {
    pb.Register${SERVICE_NAME}Server(srv, c)
    reflection.Register(srv)
}
