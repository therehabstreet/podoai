package handlers

import (
	"context"

	"github.com/therehabstreet/podoai/internal/clinical/helpers"
	pb "github.com/therehabstreet/podoai/proto/clinical"
)

// GetClinic handler
func (cs *ClinicalServer) GetClinic(ctx context.Context, req *pb.GetClinicRequest) (*pb.GetClinicResponse, error) {
	clinic, err := cs.DBClient.FetchClinicByID(ctx, req.GetClinicId())
	if err != nil {
		return nil, err
	}
	return &pb.GetClinicResponse{Clinic: helpers.ClinicModelToProto(clinic)}, nil
}

// ClinicUser CRUDL handlers
func (cs *ClinicalServer) CreateClinicUser(ctx context.Context, req *pb.CreateClinicUserRequest) (*pb.CreateClinicUserResponse, error) {
	userModel := helpers.ClinicUserProtoToModel(req.GetUser())
	createdUser, err := cs.DBClient.CreateClinicUser(ctx, userModel)
	if err != nil {
		return nil, err
	}
	return &pb.CreateClinicUserResponse{User: helpers.ClinicUserModelToProto(createdUser)}, nil
}

func (cs *ClinicalServer) GetClinicUser(ctx context.Context, req *pb.GetClinicUserRequest) (*pb.GetClinicUserResponse, error) {
	user, err := cs.DBClient.FetchClinicUserByID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	return &pb.GetClinicUserResponse{User: helpers.ClinicUserModelToProto(user)}, nil
}

func (cs *ClinicalServer) UpdateClinicUser(ctx context.Context, req *pb.UpdateClinicUserRequest) (*pb.UpdateClinicUserResponse, error) {
	userModel := helpers.ClinicUserProtoToModel(req.GetUser())
	updatedUser, err := cs.DBClient.UpdateClinicUser(ctx, userModel)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateClinicUserResponse{User: helpers.ClinicUserModelToProto(updatedUser)}, nil
}

func (cs *ClinicalServer) DeleteClinicUser(ctx context.Context, req *pb.DeleteClinicUserRequest) (*pb.DeleteClinicUserResponse, error) {
	err := cs.DBClient.DeleteClinicUserByID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	return &pb.DeleteClinicUserResponse{Success: true}, nil
}

func (cs *ClinicalServer) ListClinicUsers(ctx context.Context, req *pb.ListClinicUsersRequest) (*pb.ListClinicUsersResponse, error) {
	users, total, err := cs.DBClient.ListClinicUsers(ctx, req.GetClinicId(), req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	var protoUsers []*pb.ClinicUser
	for _, user := range users {
		protoUsers = append(protoUsers, helpers.ClinicUserModelToProto(user))
	}
	return &pb.ListClinicUsersResponse{Users: protoUsers, TotalCount: int32(total)}, nil
}
