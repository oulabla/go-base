// Code generated. DO NOT EDIT.

package user

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	pb "github.com/oulabla/go-base/gen/go/user/v1"
	"github.com/oulabla/go-base/internal/server"
)

func init() {
	server.RegisterGRPC(func(srv *grpc.Server) {
		pb.RegisterUserServiceServer(srv, NewController())
	})

	server.RegisterGateway(func(ctx context.Context, mux *runtime.ServeMux, grpcAddr string, opts []grpc.DialOption) error {
		return pb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts)
	})

	server.RegisterSwagger(func() server.SwaggerConfig {
		return server.SwaggerConfig{
			FileName: "all-apis.swagger.json", // фиксированное имя
		}
	})
}
