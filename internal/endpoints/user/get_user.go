// Code generated. DO NOT EDIT.

package user

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/oulabla/go-base/gen/go/user/v1"
)

func (c *Controller) GetUser(
	ctx context.Context,
	req *pb.GetUserRequest,
) (*pb.GetUserResponse, error) {
	// TODO: implement GetUser

	// Пример хорошей заготовки:
	// if err := req.Validate(); err != nil {
	// 	return nil, status.Error(codes.InvalidArgument, err.Error())
	// }

	// result, err := c.usecase.GetUser(ctx, req)
	// if err != nil {
	// 	return nil, status.Error(codes.Internal, "internal error")
	// }

	return nil, status.Errorf(codes.Unimplemented, "GetUser not implemented")
}
