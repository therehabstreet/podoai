package handlers

import (
	"context"

	"github.com/therehabstreet/podoai/internal/clinical/helpers"
	pb "github.com/therehabstreet/podoai/proto/clinical"
	podoai "github.com/therehabstreet/podoai/proto/common"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetPatients handler
func (cs *ClinicalServer) GetPatients(ctx context.Context, req *pb.GetPatientsRequest) (*pb.GetPatientsResponse, error) {
	page := req.GetPage()
	pageSize := req.GetPageSize()
	sortBy := req.GetSortBy()
	sortOrder := req.GetSortOrder()

	mongoPatients, totalCount, err := cs.DBClient.FetchPatients(ctx, page, pageSize, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}

	var patients []*podoai.Patient
	for _, patient := range mongoPatients {
		patients = append(patients, helpers.PatientModelToProto(patient))
	}

	return &pb.GetPatientsResponse{
		Patients:   patients,
		TotalCount: int32(totalCount),
	}, nil
}

// GetPatient handler
func (cs *ClinicalServer) GetPatient(ctx context.Context, req *pb.GetPatientRequest) (*pb.GetPatientResponse, error) {
	id, err := primitive.ObjectIDFromHex(req.GetPatientId())
	if err != nil {
		return nil, err
	}
	patientModel, err := cs.DBClient.FetchPatientByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &pb.GetPatientResponse{
		Patient: helpers.PatientModelToProto(patientModel),
	}, nil
}

// SearchPatient handler
func (cs *ClinicalServer) SearchPatient(ctx context.Context, req *pb.SearchPatientRequest) (*pb.SearchPatientResponse, error) {
	// TODO: Implement patient search logic
	return &pb.SearchPatientResponse{}, nil
}

// CreatePatient handler
func (cs *ClinicalServer) CreatePatient(ctx context.Context, req *pb.CreatePatientRequest) (*pb.CreatePatientResponse, error) {
	patientModel := helpers.PatientProtoToModel(req.GetPatient())
	createdPatient, err := cs.DBClient.CreatePatient(ctx, patientModel)
	if err != nil {
		return nil, err
	}
	return &pb.CreatePatientResponse{
		Patient: helpers.PatientModelToProto(createdPatient),
	}, nil
}

// DeletePatient handler
func (cs *ClinicalServer) DeletePatient(ctx context.Context, req *pb.DeletePatientRequest) (*pb.DeletePatientResponse, error) {
	id, err := primitive.ObjectIDFromHex(req.GetPatientId())
	if err != nil {
		return nil, err
	}
	err = cs.DBClient.DeletePatientByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &pb.DeletePatientResponse{}, nil
}
