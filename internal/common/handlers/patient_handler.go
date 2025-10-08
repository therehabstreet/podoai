package handlers

import (
	"context"

	"github.com/therehabstreet/podoai/internal/common/helpers"
	pb "github.com/therehabstreet/podoai/proto/common"
)

// GetPatients handler
func (cs *CommonServer) GetPatients(ctx context.Context, req *pb.GetPatientsRequest) (*pb.GetPatientsResponse, error) {
	page := req.GetPage()
	pageSize := req.GetPageSize()
	sortBy := req.GetSortBy()
	sortOrder := req.GetSortOrder()
	ownerEntityID := req.GetOwnerEntityId()

	mongoPatients, totalCount, err := cs.DBClient.FetchPatients(ctx, ownerEntityID, page, pageSize, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}

	var patients []*pb.Patient
	for _, patient := range mongoPatients {
		patients = append(patients, helpers.PatientModelToProto(*patient))
	}

	return &pb.GetPatientsResponse{
		Patients:   patients,
		TotalCount: int32(totalCount),
	}, nil
}

// GetPatient handler
func (cs *CommonServer) GetPatient(ctx context.Context, req *pb.GetPatientRequest) (*pb.GetPatientResponse, error) {
	patientID := req.GetPatientId()
	ownerEntityID := req.GetOwnerEntityId()

	patient, err := cs.DBClient.FetchPatientByID(ctx, patientID, ownerEntityID)
	if err != nil {
		return nil, err
	}

	return &pb.GetPatientResponse{
		Patient: helpers.PatientModelToProto(*patient),
	}, nil
}

// SearchPatient handler
func (cs *CommonServer) SearchPatient(ctx context.Context, req *pb.SearchPatientRequest) (*pb.SearchPatientResponse, error) {
	searchTerm := req.GetSearchTerm()
	ownerEntityID := req.GetOwnerEntityId()
	page := req.GetPage()
	pageSize := req.GetPageSize()

	// Default values if not provided
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// TODO: Implement patient search in database client
	mongoPatients, totalCount, err := cs.DBClient.SearchPatients(ctx, searchTerm, ownerEntityID, page, pageSize)
	if err != nil {
		return nil, err
	}

	var patients []*pb.Patient
	for _, patient := range mongoPatients {
		patients = append(patients, helpers.PatientModelToProto(*patient))
	}

	return &pb.SearchPatientResponse{
		Patients:   patients,
		TotalCount: int32(totalCount),
	}, nil
}

// CreatePatient handler
func (cs *CommonServer) CreatePatient(ctx context.Context, req *pb.CreatePatientRequest) (*pb.CreatePatientResponse, error) {
	patientModel := helpers.PatientProtoToModel(req.GetPatient())
	createdPatient, err := cs.DBClient.CreatePatient(ctx, patientModel)
	if err != nil {
		return nil, err
	}
	return &pb.CreatePatientResponse{
		Patient: helpers.PatientModelToProto(*createdPatient),
	}, nil
}

// DeletePatient handler
func (cs *CommonServer) DeletePatient(ctx context.Context, req *pb.DeletePatientRequest) (*pb.DeletePatientResponse, error) {
	patientID := req.GetPatientId()
	ownerEntityID := req.GetOwnerEntityId()

	err := cs.DBClient.DeletePatientByID(ctx, patientID, ownerEntityID)
	if err != nil {
		return nil, err
	}

	return &pb.DeletePatientResponse{
		Success: true,
	}, nil
}
