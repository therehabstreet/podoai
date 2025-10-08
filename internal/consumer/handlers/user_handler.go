package handlers

import (
	"context"

	"github.com/therehabstreet/podoai/internal/consumer/helpers"
	pb "github.com/therehabstreet/podoai/proto/consumer"
)

// GetUser handler
func (cs *ConsumerServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	userModel, err := cs.DBClient.FetchUserByID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &pb.GetUserResponse{
		User: helpers.UserModelToProto(*userModel),
	}, nil
}

// UpdateUser handler
func (cs *ConsumerServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	userModel := helpers.UserProtoToModel(req.GetUser())
	updatedUser, err := cs.DBClient.UpdateUser(ctx, userModel)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateUserResponse{
		User: helpers.UserModelToProto(*updatedUser),
	}, nil
}
