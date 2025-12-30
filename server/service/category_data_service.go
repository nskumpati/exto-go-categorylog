package service

import (
	"errors"
	"fmt"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/model"
	"github.com/gaeaglobal/exto/server/repo"
	"github.com/gaeaglobal/exto/server/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CategoryDataService struct {
	r                  repo.CategoryDataRepository
	categoryService    *CategoryService
	orgService         *OrganizationService
	scanHistoryService *ScanHistoryService
}

func NewCategoryDataService(r repo.CategoryDataRepository, categoryService *CategoryService, orgService *OrganizationService, scanHistoryService *ScanHistoryService) *CategoryDataService {
	return &CategoryDataService{
		r:                  r,
		categoryService:    categoryService,
		orgService:         orgService,
		scanHistoryService: scanHistoryService,
	}
}

func (s *CategoryDataService) getCategorySlug(reqCtx *app.RequestContext, categoryID bson.ObjectID) (string, error) {
	category, err := s.categoryService.GetCategoryByID(reqCtx, categoryID)
	if err != nil {
		return "", fmt.Errorf("failed to get category: %w", err)
	}
	if category == nil {
		return "", errors.New("category not found")
	}
	return category.Slug, nil
}

func (s *CategoryDataService) getCategoryPrimaryField(reqCtx *app.RequestContext, categoryID bson.ObjectID) (string, error) {
	category, err := s.categoryService.GetCategoryByID(reqCtx, categoryID)
	if err != nil {
		return "", fmt.Errorf("failed to get category: %w", err)
	}
	if category == nil {
		return "", errors.New("category not found")
	}
	return category.PrimaryField, nil
}

func (s *CategoryDataService) GetCategoryDataByID(reqCtx *app.RequestContext, categoryID bson.ObjectID, id bson.ObjectID) (*model.CategoryData, error) {
	categorySlug, err := s.getCategorySlug(reqCtx, categoryID)
	if err != nil {
		return nil, err
	}
	return s.r.GetCategoryDataByID(reqCtx, categorySlug, id)
}

func (s *CategoryDataService) ListCategoryData(reqCtx *app.RequestContext, categoryID bson.ObjectID, pageReq *app.PageRequest) (*app.PageResponse[*model.CategoryData], error) {
	categorySlug, err := s.getCategorySlug(reqCtx, categoryID)
	if err != nil {
		return nil, err
	}
	return s.r.ListCategoryData(reqCtx, categorySlug, pageReq)
}

type CreateCategoryDataResult struct {
	CategoryData *model.CategoryData
	ScanHistory  *model.ScanHistory
}

func (s *CategoryDataService) CreateCategoryData(
	reqCtx *app.RequestContext,
	categoryID bson.ObjectID,
	data *map[string]any,
	rawData *map[string]any,
	filePath string,
	batchID bson.ObjectID,
) (*CreateCategoryDataResult, error) {
	categorySlug, err := s.getCategorySlug(reqCtx, categoryID)
	if err != nil {
		return nil, err
	}
	thumbnail, err := utils.GetThumbnailImage(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create thumbnail: %w", err)
	}
	newCategoryData := &model.CreateCategoryDataRequest{
		CategoryID:    categoryID,
		MetaData:      *data,
		RawData:       *rawData,
		DocumentPaths: []string{filePath},
	}

	primaryField, err := s.getCategoryPrimaryField(reqCtx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get primary field: %w", err)
	}

	//get primary field value from data map
	scanCode, _ := (*data)[primaryField].(string)
	if scanCode == "" {
		scanCode, err = s.orgService.GenerateNextScanCode(reqCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate scan code: %w", err)
		}
	}

	catData, err := s.r.CreateCategoryData(reqCtx, categorySlug, newCategoryData)
	if err != nil {
		return nil, err
	}

	scanHistory := &model.CreateScanHistoryRequest{
		CategoryID:      categoryID,
		CategoryDataID:  catData.ID,
		BatchID:         batchID,
		CategoryDataCol: categorySlug + "_data",
		ScanCode:        scanCode,
		Thumbnails:      []string{thumbnail},
	}

	scanHistoryRes, err := s.scanHistoryService.CreateScanHistory(reqCtx, scanHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to create scan history: %w", err)
	}

	return &CreateCategoryDataResult{
		CategoryData: catData,
		ScanHistory:  scanHistoryRes,
	}, nil
}

func (s *CategoryDataService) UpdateCategoryData(reqCtx *app.RequestContext, categoryID bson.ObjectID, dataID bson.ObjectID, updatedExtractedData *map[string]any) (*model.CategoryData, error) {
	categorySlug, err := s.getCategorySlug(reqCtx, categoryID)
	if err != nil {
		return nil, err
	}
	return s.r.UpdateCategoryData(reqCtx, categorySlug, dataID, &model.UpdateCategoryDataRequest{
		MetaData: *updatedExtractedData,
	})
}
