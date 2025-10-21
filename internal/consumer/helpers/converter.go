package helpers

import (
	"time"

	commonHelpers "github.com/therehabstreet/podoai/internal/common/helpers"
	"github.com/therehabstreet/podoai/internal/consumer/models"
	pb "github.com/therehabstreet/podoai/proto/consumer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// User Proto <-> Model conversion
func UserProtoToModel(proto *pb.User) *models.User {
	return &models.User{
		ID:        proto.GetId(),
		Name:      proto.GetName(),
		Phone:     proto.GetPhone(),
		Email:     proto.GetEmail(),
		Age:       proto.GetAge(),
		Gender:    commonHelpers.GenderProtoToString(proto.GetGender()),
		CreatedAt: timestampPBToTime(proto.GetCreatedAt()),
		UpdatedAt: timestampPBToTime(proto.GetUpdatedAt()),
	}
}

func UserModelToProto(model *models.User) *pb.User {
	return &pb.User{
		Id:        model.ID,
		Name:      model.Name,
		Phone:     model.Phone,
		Email:     model.Email,
		Age:       model.Age,
		Gender:    commonHelpers.GenderStringToProto(model.Gender),
		CreatedAt: timestamppb.New(model.CreatedAt),
		UpdatedAt: timestamppb.New(model.UpdatedAt),
	}
}

// Helper for timestamp conversion
func timestampPBToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}
