package handlers

import (
	"context"

	"github.com/therehabstreet/podoai/internal/common/helpers"
	pb "github.com/therehabstreet/podoai/proto/common"
)

// GetProduct handles the GetProduct gRPC request
func (cs *CommonServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	productID := req.GetProductId()

	product, err := cs.DBClient.FetchProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	return &pb.GetProductResponse{
		Product: helpers.ProductModelToProto(*product),
	}, nil
}

// GetExercise handles the GetExercise gRPC request
func (cs *CommonServer) GetExercise(ctx context.Context, req *pb.GetExerciseRequest) (*pb.GetExerciseResponse, error) {
	exerciseID := req.GetExerciseId()

	exercise, err := cs.DBClient.FetchExerciseByID(ctx, exerciseID)
	if err != nil {
		return nil, err
	}

	return &pb.GetExerciseResponse{
		Exercise: helpers.ExerciseModelToProto(*exercise),
	}, nil
}

// GetTherapy handles the GetTherapy gRPC request
func (cs *CommonServer) GetTherapy(ctx context.Context, req *pb.GetTherapyRequest) (*pb.GetTherapyResponse, error) {
	therapyID := req.GetTherapyId()

	therapy, err := cs.DBClient.FetchTherapyByID(ctx, therapyID)
	if err != nil {
		return nil, err
	}

	return &pb.GetTherapyResponse{
		Therapy: helpers.TherapyModelToProto(*therapy),
	}, nil
}
