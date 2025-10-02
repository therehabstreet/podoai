package handlers

import (
	"context"

	"github.com/therehabstreet/podoai/internal/common/helpers"
	pb "github.com/therehabstreet/podoai/proto/common"
)

// GetScans handles the GetScans gRPC request
func (cs *CommonServer) GetScans(ctx context.Context, req *pb.GetScansRequest) (*pb.GetScansResponse, error) {
	userID := req.GetUserId()
	page := req.GetPage()
	pageSize := req.GetPageSize()
	sortBy := req.GetSortBy()
	sortOrder := req.GetSortOrder()

	scans, total, err := cs.DBClient.FetchScans(ctx, userID, page, pageSize, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}

	var protoScans []*pb.Scan
	for _, scan := range scans {
		protoScans = append(protoScans, helpers.ScanModelToProto(scan))
	}

	return &pb.GetScansResponse{
		Scans:      protoScans,
		TotalCount: int32(total),
	}, nil
}

// GetScan handles the GetScan gRPC request
func (cs *CommonServer) GetScan(ctx context.Context, req *pb.GetScanRequest) (*pb.GetScanResponse, error) {
	scanID := req.GetScanId()

	scan, err := cs.DBClient.FetchScanByID(ctx, scanID)
	if err != nil {
		return nil, err
	}

	return &pb.GetScanResponse{
		Scan: helpers.ScanModelToProto(scan),
	}, nil
}

// CreateScan handles the CreateScan gRPC request
func (cs *CommonServer) CreateScan(ctx context.Context, req *pb.CreateScanRequest) (*pb.CreateScanResponse, error) {
	scanModel := helpers.ScanProtoToModel(req.GetScan())

	createdScan, err := cs.DBClient.CreateScan(ctx, scanModel)
	if err != nil {
		return nil, err
	}

	return &pb.CreateScanResponse{
		Scan: helpers.ScanModelToProto(createdScan),
	}, nil
}

// DeleteScan handles the DeleteScan gRPC request
func (cs *CommonServer) DeleteScan(ctx context.Context, req *pb.DeleteScanRequest) (*pb.DeleteScanResponse, error) {
	scanID := req.GetScanId()

	err := cs.DBClient.DeleteScanByID(ctx, scanID)
	if err != nil {
		return &pb.DeleteScanResponse{Success: false}, err
	}

	return &pb.DeleteScanResponse{Success: true}, nil
}


