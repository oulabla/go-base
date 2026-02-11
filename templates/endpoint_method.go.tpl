// Code generated. DO NOT EDIT.

package ${SERVICE_PKG_NAME}

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "${MODULE}/gen/go/${SERVICE_PKG_NAME}"
)

func (c *Controller) ${METHOD_NAME}(
	ctx context.Context,
	req *pb.${METHOD_NAME}Request,
) (*pb.${METHOD_NAME}Response, error) {
	// TODO: implement ${METHOD_NAME}

	// Пример хорошей заготовки:
	// if err := req.Validate(); err != nil {
	// 	return nil, status.Error(codes.InvalidArgument, err.Error())
	// }

	// result, err := c.usecase.${METHOD_NAME}(ctx, req)
	// if err != nil {
	// 	return nil, status.Error(codes.Internal, "internal error")
	// }

	return nil, status.Errorf(codes.Unimplemented, "${METHOD_NAME} not implemented")
}