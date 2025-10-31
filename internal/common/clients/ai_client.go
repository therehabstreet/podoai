package clients

import (
	"context"
	"fmt"

	"github.com/therehabstreet/podoai/internal/common/models"
)

type AIClient interface {
	AnalyzeScan(ctx context.Context, scanID string, images []*models.ScanImage, videos []*models.ScanVideo) (*AIAnalysisResult, error)
	GenerateLLMReport(ctx context.Context, scanID string, analysisResult *AIAnalysisResult) (*LLMReportResult, error)
}

type AIAnalysisResult struct {
	FootScore              float64
	GaitScore              float64
	LeftPronationAngle     float64
	RightPronationAngle    float64
	LeftArchHeightIndex    float64
	RightArchHeightIndex   float64
	LeftHalluxValgusAngle  float64
	RightHalluxValgusAngle float64
}

type LLMReportResult struct {
	Summary string
}

type PythonAIClient struct {
	baseURL string
}

func NewPythonAIClient(baseURL string) *PythonAIClient {
	return &PythonAIClient{
		baseURL: baseURL,
	}
}

func (c *PythonAIClient) AnalyzeScan(ctx context.Context, scanID string, images []*models.ScanImage, videos []*models.ScanVideo) (*AIAnalysisResult, error) {
	// TODO: Implement gRPC/HTTP call to Python AI service
	// This will call the Python microservice's AnalyzeScan endpoint
	//
	// Expected proto/API contract:
	// - Input: scan_id, list of image paths, list of video paths
	// - Output: AIAnalysisResult with all biomechanical parameters
	//
	// Example implementation:
	// 1. Create request with scan_id and media paths
	// 2. Call Python service via gRPC: aiService.AnalyzeScan(ctx, req)
	// 3. Parse response and map to AIAnalysisResult
	// 4. Return result

	return nil, fmt.Errorf("ai service not yet implemented")
}

func (c *PythonAIClient) GenerateLLMReport(ctx context.Context, scanID string, analysisResult *AIAnalysisResult) (*LLMReportResult, error) {
	// TODO: Implement gRPC/HTTP call to Python AI service
	// This will call the Python microservice's GenerateLLMReport endpoint
	//
	// Expected proto/API contract:
	// - Input: scan_id, AIAnalysisResult (all biomechanical parameters)
	// - Output: LLMReportResult with generated summary text
	//
	// Example implementation:
	// 1. Create request with scan_id and analysis parameters
	// 2. Call Python service via gRPC: aiService.GenerateLLMReport(ctx, req)
	// 3. Parse response and extract summary text
	// 4. Return LLMReportResult

	return nil, fmt.Errorf("ai service not yet implemented")
}
