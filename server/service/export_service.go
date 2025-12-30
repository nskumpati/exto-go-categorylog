package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/model"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ExportResult struct {
	FilePath string
	FileName string
	Error    error
}

type ExportService struct {
	appCtx              *app.AppContext
	orgService          *OrganizationService
	scanHistoryService  *ScanHistoryService
	categoryService     *CategoryService
	categoryDataService *CategoryDataService
}

func NewExportService(appCtx *app.AppContext, orgService *OrganizationService, scanHistoryService *ScanHistoryService, categoryService *CategoryService, categoryDataService *CategoryDataService) *ExportService {
	return &ExportService{
		appCtx:              appCtx,
		orgService:          orgService,
		scanHistoryService:  scanHistoryService,
		categoryService:     categoryService,
		categoryDataService: categoryDataService,
	}
}

func (s *ExportService) ExportAsExcel(reqCtx *app.RequestContext, scanHistoryID string) (*ExportResult, error) {

	scanHistoryObjectID, err := bson.ObjectIDFromHex(scanHistoryID)
	if err != nil || scanHistoryObjectID.IsZero() {
		return nil, errors.New("invalid scan history ID format")
	}

	scanHistory, err := s.scanHistoryService.GetScanHistoryByID(reqCtx, scanHistoryObjectID)
	if err != nil {
		return nil, err
	}

	if scanHistory == nil {
		return nil, errors.New("scan history not found")
	}

	category, err := s.categoryService.GetCategoryByID(reqCtx, scanHistory.CategoryID)
	if err != nil {
		return nil, err
	}

	if category == nil {
		return nil, errors.New("category not found")
	}

	tableMap := make(map[string][]model.Field, 0)

	for _, field := range category.Fields {
		if field.Type == model.FieldTypeTable {
			tableMap[field.Name] = field.Children
		}
	}

	// Load category data
	catData, err := s.categoryDataService.GetCategoryDataByID(reqCtx, scanHistory.CategoryID, scanHistory.CategoryDataID)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()

	if err := f.SetSheetName("Sheet1", "Header"); err != nil {
		return nil, err
	}

	colIdx := 0
	for _, field := range category.Fields {
		if field.Type != model.FieldTypeTable {
			col := string(rune('A' + colIdx))
			f.SetCellValue("Header", fmt.Sprintf("%s1", col), field.Label)
			colIdx++
		}
	}

	colIdxr := 0
	for _, field := range category.Fields {
		if field.Type != model.FieldTypeTable {
			col := string(rune('A' + colIdxr))
			val := catData.MetaData[field.Name]
			if val != nil {
				f.SetCellValue("Header", fmt.Sprintf("%s%d", col, 2), val)
			}
			colIdxr++
		}
	}

	for tableName, fields := range tableMap {
		_, err := f.NewSheet(tableName)
		if err != nil {
			return nil, err
		}

		// Header
		colIdx := 0
		for _, field := range fields {
			if field.Type != model.FieldTypeTable {
				col := string(rune('A' + colIdx))
				f.SetCellValue(tableName, fmt.Sprintf("%s1", col), field.Label)
				colIdx++
			}
		}

		// Data

		values := catData.MetaData[tableName]
		if values == nil {
			continue
		}
		// assert values to map[string]interface{}
		valuesArray, ok := values.(bson.A)
		if !ok {
			continue
		}

		for rowIdx, row := range valuesArray {
			// Display row type name
			rowMap, ok := row.(bson.D)
			if !ok {
				continue
			}
			interfaceMap := make(map[string]interface{})
			for _, elem := range rowMap {
				interfaceMap[fmt.Sprintf("%v", elem.Key)] = elem.Value
			}
			colIdx := 0
			for _, field := range fields {
				if field.Type != model.FieldTypeTable {
					col := string(rune('A' + colIdx))
					val := interfaceMap[field.Name]
					if val != nil {
						f.SetCellValue(tableName, fmt.Sprintf("%s%d", col, rowIdx+2), val)
					}
					colIdx++
				}
			}
		}

	}

	// Save
	date_string := time.Now().Format("2006_01_02_15_04_05")
	file_name := fmt.Sprintf("export_%s_%s.xlsx", category.Slug, date_string)
	dir := filepath.Join(s.appCtx.Config.EXPORT_DIR, reqCtx.User.OrganizationID.Hex())
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	dst := filepath.Join(dir, file_name)
	if err := f.SaveAs(dst); err != nil {
		return nil, err
	}
	log.Printf("Exported file to: %s", dst)
	return &ExportResult{FilePath: dst, FileName: file_name}, nil
}
