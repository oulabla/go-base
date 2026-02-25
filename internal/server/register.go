// internal/server/register.go
package server

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	"github.com/oulabla/go-base/internal/config"
)

type (
	GRPCRegisterFn    func(*grpc.Server)
	GatewayRegisterFn func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error

	// Теперь возвращаем конфиг с префиксом (чтобы пути не пересекались)
	SwaggerRegisterFn func() SwaggerConfig
)

type SwaggerConfig struct {
	FileName string
}

var (
	grpcRegs    []GRPCRegisterFn
	gatewayRegs []GatewayRegisterFn
	swaggerRegs []SwaggerRegisterFn
)

func RegisterGRPC(fn GRPCRegisterFn) {
	grpcRegs = append(grpcRegs, fn)
}

func RegisterGateway(fn GatewayRegisterFn) {
	gatewayRegs = append(gatewayRegs, fn)
}

func RegisterSwagger(fn SwaggerRegisterFn) {
	swaggerRegs = append(swaggerRegs, fn)
}

func RegisterAllGRPC(srv *grpc.Server) {
	for _, fn := range grpcRegs {
		fn(srv)
	}
}

func RegisterAllGateway(ctx context.Context, mux *runtime.ServeMux, opts []grpc.DialOption) error {
	grpcAddr := config.GetString(ctx, config.K.ServerGrpcPort)
	for _, fn := range gatewayRegs {
		if err := fn(ctx, mux, grpcAddr, opts); err != nil {
			return err
		}
	}
	return nil
}
