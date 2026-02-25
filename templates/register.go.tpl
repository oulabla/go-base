// Code generated. DO NOT EDIT.

package ${SERVICE_PKG_NAME}

import (
    "context"

    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    "google.golang.org/grpc"

    ${PROTO_IMPORT}
    "github.com/oulabla/go-base/internal/server"
)

func init() {
    server.RegisterGRPC(func(srv *grpc.Server) {
        pb.Register${SERVICE_NAME}Server(srv, NewController())
    })

    server.RegisterGateway(func(ctx context.Context, mux *runtime.ServeMux, grpcAddr string, opts []grpc.DialOption) error {
        return pb.Register${SERVICE_NAME}HandlerFromEndpoint(ctx, mux, grpcAddr, opts)
    })

    server.RegisterSwagger(func() server.SwaggerConfig {
        return server.SwaggerConfig{
            FileName: "all-apis.swagger.json",   // фиксированное имя
        }
    })
}