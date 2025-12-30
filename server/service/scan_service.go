package service

import (
	"errors"
	"fmt"
	"log"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/file_utils"
	"github.com/gaeaglobal/exto/server/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ScanService struct {
	appCtx              *app.AppContext
	orgService          *OrganizationService
	openAIService       *OpenAIService
	batchService        *BatchService
	scanHistoryService  *ScanHistoryService
	categoryDataService *CategoryDataService
	meterService        *MeterService
}

func NewScanService(appCtx *app.AppContext, orgService *OrganizationService, batchService *BatchService, openAIService *OpenAIService, scanHistoryService *ScanHistoryService, categoryDataService *CategoryDataService, meterService *MeterService) *ScanService {
	return &ScanService{
		appCtx:              appCtx,
		orgService:          orgService,
		batchService:        batchService,
		openAIService:       openAIService,
		scanHistoryService:  scanHistoryService,
		categoryDataService: categoryDataService,
		meterService:        meterService,
	}
}

type ScanResult struct {
	BatchID        string
	CategoryDataID string
	ScanHistoryID  string
	ScanCode       string
	Data           map[string]any
}

// Example method to perform the scan and return ScanResult
func (s *ScanService) PerformScan(appCtx *app.AppContext, reqCtx *app.RequestContext, c *gin.Context, categoryObjID bson.ObjectID, batchObjID bson.ObjectID) (*ScanResult, error) {

	// Save the uploaded file to disk
	filePath, fileErr := SaveScanFileToDisk(appCtx, reqCtx, c)
	if fileErr != nil {
		return nil, fmt.Errorf("failed to save uploaded file: %w", fileErr)
	}

	// Get base64 string from file path
	base64Image, base64Err := utils.GetBase64FromFilePath(filePath)
	if base64Err != nil {
		return nil, fmt.Errorf("failed to get base64 image: %w", base64Err)
	}

	return s.PerformOpenAIScan(reqCtx, categoryObjID, base64Image, batchObjID, filePath)

}

func SaveScanFileToDisk(appCtx *app.AppContext, reqCtx *app.RequestContext, c *gin.Context) (string, error) {
	// Retrieve the file from the form data
	file, err := c.FormFile("file")
	if err != nil {
		return "", errors.New("Uploaded file not found in form data (expected key: 'file')")
	}

	dst := file_utils.GetFilePath(appCtx, reqCtx, appCtx.Config.UPLOAD_DIR, file.Filename, 7, 0)
	if dst == "" {
		return "", errors.New("failed to generate unique file path for upload (path generation issue)")
	}

	// Save the file
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}
	return dst, nil
}

func (s *ScanService) PerformOpenAIScan(reqCtx *app.RequestContext, categoryObjID bson.ObjectID, base64Image string, batchID bson.ObjectID, filePath string) (*ScanResult, error) {
	evt, err := s.meterService.IncrementMeterEvent(reqCtx, "scan", 1)
	if err != nil {
		log.Printf("Failed to create meter event: %v", err)
	}
	if evt != nil {
		log.Printf("Created meter event")
	} else {
		log.Printf("Failed to create meter event: event is nil")
	}
	result, err := s.openAIService.ExtractDocumentData(reqCtx, categoryObjID, base64Image)
	if err != nil {
		return nil, errors.New("OpenAI extraction failed")
	}

	resultStr, ok := result.(string)
	if !ok {
		return nil, errors.New("failed to assert result as string")
	}

	extractedMap, err := s.openAIService.ConvertKeyValueToMap(resultStr)
	if err != nil {
		return nil, errors.New("failed to convert result to map")
	}

	confidenceMap, err := s.openAIService.ConvertConfidenceScoresToMap(resultStr)
	if err != nil {
		return nil, errors.New("failed to convert confidence scores to map")
	}

	// Convert map[string]int to map[string]float64
	confidenceMapFloat := make(map[string]float64, len(confidenceMap))
	for k, v := range confidenceMap {
		confidenceMapFloat[k] = float64(v)
	}

	avg := s.openAIService.CalculateAverageConfidence(confidenceMap)
	fmt.Println("Average confidence:", avg)

	rawData := map[string]any{
		"extractedData":     extractedMap,
		"confidenceScores":  confidenceMapFloat,
		"averageConfidence": avg,
	}

	categoryDataRes, err := s.categoryDataService.CreateCategoryData(reqCtx, categoryObjID, &extractedMap, &rawData, filePath, batchID)
	if err != nil {
		return nil, errors.New("failed to save category data")
	}

	log.Printf("Created Category Data: %+v\n", categoryDataRes)

	scanResult := &ScanResult{
		BatchID:        batchID.Hex(),
		CategoryDataID: categoryDataRes.CategoryData.ID.Hex(),
		ScanHistoryID:  categoryDataRes.ScanHistory.ID.Hex(),
		ScanCode:       categoryDataRes.ScanHistory.ScanCode,
		Data:           rawData,
	}

	return scanResult, nil
}
