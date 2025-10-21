package helpers

import (
	"time"

	"github.com/therehabstreet/podoai/internal/clinical/models"
	pb "github.com/therehabstreet/podoai/proto/clinical"
	podoai "github.com/therehabstreet/podoai/proto/common"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mongo Clinic -> Proto Clinic
func ClinicModelToProto(m models.Clinic) *pb.Clinic {
	return &pb.Clinic{
		Id:        m.ID,
		Name:      m.Name,
		Address:   m.Address,
		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
	}
}

// Proto Clinic -> Mongo Clinic
func ClinicProtoToModel(p *pb.Clinic) models.Clinic {
	createdAt := time.Now()
	if p.CreatedAt != nil {
		createdAt = p.CreatedAt.AsTime()
	}
	updatedAt := time.Now()
	if p.UpdatedAt != nil {
		updatedAt = p.UpdatedAt.AsTime()
	}
	return models.Clinic{
		ID:        p.Id,
		Name:      p.Name,
		Address:   p.Address,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

// Mongo ClinicUser -> Proto ClinicUser
func ClinicUserModelToProto(m models.ClinicUser) *pb.ClinicUser {
	var protoRoles []podoai.Role
	for _, r := range m.Roles {
		protoRoles = append(protoRoles, podoai.Role(podoai.Role_value[r]))
	}
	return &pb.ClinicUser{
		Id:          m.ID,
		Name:        m.Name,
		PhoneNumber: m.PhoneNumber,
		Roles:       protoRoles,
		ClinicId:    m.ClinicID,
		CreatedAt:   timestamppb.New(m.CreatedAt),
		UpdatedAt:   timestamppb.New(m.UpdatedAt),
	}
}

// Proto ClinicUser -> Mongo ClinicUser
func ClinicUserProtoToModel(p *pb.ClinicUser) models.ClinicUser {
	var modelRoles []string
	for _, r := range p.Roles {
		modelRoles = append(modelRoles, r.String())
	}
	createdAt := time.Time{}
	updatedAt := time.Time{}
	if p.CreatedAt != nil {
		createdAt = p.CreatedAt.AsTime()
	}
	if p.UpdatedAt != nil {
		updatedAt = p.UpdatedAt.AsTime()
	}
	return models.ClinicUser{
		ID:          p.Id,
		Name:        p.Name,
		PhoneNumber: p.PhoneNumber,
		Roles:       modelRoles,
		ClinicID:    p.ClinicId,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
