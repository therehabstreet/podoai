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

	// TODO: Implement AI analysis logic
	// - Call AI service for arch type detection
	// - Call AI service for pronation analysis
	// - Calculate balance score
	// - Update scan with AI results

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

	// TODO: Implement LLM report generation logic
	// - Extract AI analysis results from scan
	// - Call LLM service to generate detailed report
	// - Format report with findings and explanations
	// - Update scan with generated report

	scanRun.Scan.Status = pb.ScanStatus_REPORT_GENERATED.String()
	if _, err := scanRun.MongoClient.UpdateScan(ctx, scanRun.Scan); err != nil {
		return fmt.Errorf("failed to update scan %s with LLM report: %w", scanRun.Scan.ID, err)
	}

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
