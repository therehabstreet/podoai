package workers

import (
	"context"
	"fmt"
	"log"

	"github.com/therehabstreet/podoai/internal/common/clients"
	"github.com/therehabstreet/podoai/internal/common/models"
	pb "github.com/therehabstreet/podoai/proto/common"
)

// ScanResultWorkflowRun implements WorkflowRun for scan processing
type ScanResultWorkflowRun struct {
	Scan        *models.Scan
	MongoClient clients.DBClient
	AIClient    clients.AIClient
	Steps       []WorkflowStep
}

// GetID returns the scan ID
func (st *ScanResultWorkflowRun) GetID() string {
	return st.Scan.ID
}

// GetType returns the workflow type
func (st *ScanResultWorkflowRun) GetType() string {
	return "scan_result_workflow"
}

func (st *ScanResultWorkflowRun) GetSteps() []WorkflowStep {
	return st.Steps
}

func (st *ScanResultWorkflowRun) GetInput() map[string]any {
	return map[string]any{
		"scan": st.Scan,
	}
}

/****************************** Steps ***************************************/

type RunAIAnalysisStep struct{}

func (s *RunAIAnalysisStep) GetName() string {
	return "RunAIAnalysis"
}

func (s *RunAIAnalysisStep) Execute(ctx context.Context, input map[string]any) error {
	scanRun, ok := input["scan"].(*ScanResultWorkflowRun)
	if !ok {
		return fmt.Errorf("invalid run type for RunAIAnalysisStep")
	}

	// Check if already completed this step
	if scanRun.Scan.Status != pb.ScanStatus_MEDIA_UPLOADED.String() {
		log.Printf("Scan %s already past AI analysis stage (status: %s), skipping", scanRun.Scan.ID, scanRun.Scan.Status)
		return nil
	}

	// Update status to AI_PROCESSING
	scanRun.Scan.Status = pb.ScanStatus_AI_PROCESSING.String()
	if _, err := scanRun.MongoClient.UpdateScan(ctx, scanRun.Scan); err != nil {
		return fmt.Errorf("failed to update scan status to AI_PROCESSING: %w", err)
	}
	log.Printf("Scan %s status updated to AI_PROCESSING", scanRun.Scan.ID)

	// Call Python AI service to analyze scan
	analysisResult, err := scanRun.AIClient.AnalyzeScan(ctx, scanRun.Scan.ID, scanRun.Scan.Images, scanRun.Scan.Videos)
	if err != nil {
		return fmt.Errorf("failed to analyze scan: %w", err)
	}

	// Initialize ScanAIResult if not exists
	if scanRun.Scan.ScanAIResult == nil {
		scanRun.Scan.ScanAIResult = &models.ScanAIResult{}
	}

	// Update scan with AI analysis results
	scanRun.Scan.ScanAIResult.FootScore = analysisResult.FootScore
	scanRun.Scan.ScanAIResult.GaitScore = analysisResult.GaitScore
	scanRun.Scan.ScanAIResult.LeftPronationAngle = analysisResult.LeftPronationAngle
	scanRun.Scan.ScanAIResult.RightPronationAngle = analysisResult.RightPronationAngle
	scanRun.Scan.ScanAIResult.LeftArchHeightIndex = analysisResult.LeftArchHeightIndex
	scanRun.Scan.ScanAIResult.RightArchHeightIndex = analysisResult.RightArchHeightIndex
	scanRun.Scan.ScanAIResult.LeftHalluxValgusAngle = analysisResult.LeftHalluxValgusAngle
	scanRun.Scan.ScanAIResult.RightHalluxValgusAngle = analysisResult.RightHalluxValgusAngle

	// Save updated scan with AI results
	if _, err := scanRun.MongoClient.UpdateScan(ctx, scanRun.Scan); err != nil {
		return fmt.Errorf("failed to update scan with AI results: %w", err)
	}
	log.Printf("Scan %s AI analysis completed and saved", scanRun.Scan.ID)

	return nil
}

type GenerateLLMReportStep struct{}

func (s *GenerateLLMReportStep) GetName() string {
	return "GenerateLLMReport"
}

func (s *GenerateLLMReportStep) Execute(ctx context.Context, input map[string]any) error {
	scanRun, ok := input["scan"].(*ScanResultWorkflowRun)
	if !ok {
		return fmt.Errorf("invalid run type for GenerateLLMReportStep")
	}

	// Check if already completed this step
	if scanRun.Scan.Status != pb.ScanStatus_AI_PROCESSING.String() {
		log.Printf("Scan %s already past LLM report stage (status: %s), skipping", scanRun.Scan.ID, scanRun.Scan.Status)
		return nil
	}

	// Ensure we have AI analysis results
	if scanRun.Scan.ScanAIResult == nil {
		return fmt.Errorf("scan %s has no AI analysis results", scanRun.Scan.ID)
	}

	// Prepare analysis result for LLM
	analysisResult := &clients.AIAnalysisResult{
		FootScore:              scanRun.Scan.ScanAIResult.FootScore,
		GaitScore:              scanRun.Scan.ScanAIResult.GaitScore,
		LeftPronationAngle:     scanRun.Scan.ScanAIResult.LeftPronationAngle,
		RightPronationAngle:    scanRun.Scan.ScanAIResult.RightPronationAngle,
		LeftArchHeightIndex:    scanRun.Scan.ScanAIResult.LeftArchHeightIndex,
		RightArchHeightIndex:   scanRun.Scan.ScanAIResult.RightArchHeightIndex,
		LeftHalluxValgusAngle:  scanRun.Scan.ScanAIResult.LeftHalluxValgusAngle,
		RightHalluxValgusAngle: scanRun.Scan.ScanAIResult.RightHalluxValgusAngle,
	}

	// Call Python AI service to generate LLM report
	llmResult, err := scanRun.AIClient.GenerateLLMReport(ctx, scanRun.Scan.ID, analysisResult)
	if err != nil {
		return fmt.Errorf("failed to generate LLM report: %w", err)
	}

	// Update scan with LLM report
	scanRun.Scan.ScanAIResult.LLMResult = &models.ScanLLMResult{
		Summary: llmResult.Summary,
	}

	// Update status to REPORT_GENERATED
	scanRun.Scan.Status = pb.ScanStatus_REPORT_GENERATED.String()
	if _, err := scanRun.MongoClient.UpdateScan(ctx, scanRun.Scan); err != nil {
		return fmt.Errorf("failed to update scan with LLM report: %w", err)
	}
	log.Printf("Scan %s LLM report generated and saved", scanRun.Scan.ID)

	return nil
}

type GenerateRecommendationsStep struct{}

func (s *GenerateRecommendationsStep) GetName() string {
	return "GenerateRecommendations"
}

func (s *GenerateRecommendationsStep) Execute(ctx context.Context, input map[string]any) error {
	scanRun, ok := input["scan"].(*ScanResultWorkflowRun)
	if !ok {
		return fmt.Errorf("invalid run type for GenerateRecommendationsStep")
	}

	// Check if already completed this step
	if scanRun.Scan.Status == pb.ScanStatus_RECOMMENDATIONS_GENERATED.String() {
		log.Printf("Scan %s already has recommendations (status: %s), skipping", scanRun.Scan.ID, scanRun.Scan.Status)
		return nil
	}

	// TODO: Implement recommendations logic
	// - Based on AI results, recommend products
	// - Recommend exercises
	// - Recommend therapies
	// - Update scan with recommendations

	// Update status to RECOMMENDATIONS_GENERATED
	scanRun.Scan.Status = pb.ScanStatus_RECOMMENDATIONS_GENERATED.String()
	if _, err := scanRun.MongoClient.UpdateScan(ctx, scanRun.Scan); err != nil {
		return fmt.Errorf("failed to update scan %s status to RECOMMENDATIONS_GENERATED: %w", scanRun.Scan.ID, err)
	}

	return nil
}
