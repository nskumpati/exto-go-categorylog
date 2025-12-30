package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/app_di"
	"github.com/gaeaglobal/exto/server/file_utils"
	"github.com/gaeaglobal/exto/server/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func AddCategoryDataRoutes(router *gin.RouterGroup) {
	//Not needed create will happen via scan route
	router.POST("/categories/:categoryID/data", createCategoryDataEndpoint)
	router.PATCH("/categories/:categoryID/data/:dataID", updateCategoryDataEndpoint)
}

func createCategoryDataEndpoint(c *gin.Context) {
	// User uploads the file along with category data
	appCtx, exists := app.GetAppContext(c)
	if !exists {
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Application context not found"))
		return
	}

	di, found := app_di.GetAppDI(c)
	if !found {
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Failed to get application DI"))
		return
	}

	reqCtx := app.GetRequestCtx(c)
	if reqCtx.IsZero() {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Current user not found"))
		return
	}

	categoryID := c.Param("categoryID")
	if categoryID == "" {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Category ID is required"))
		return
	}

	categoryIDObj, err := bson.ObjectIDFromHex(categoryID)
	if err != nil {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Invalid category ID format"))
		return
	}

	// Save the uploaded file to disk
	filePath, err := SaveFileToDisk(appCtx, reqCtx, c)
	if err != nil {
		c.AbortWithStatusJSON(500, utils.NewErrorResponse(fmt.Sprintf("Failed to save file: %s", err.Error())))
		return
	}

	data, err := GetCategoryData(c)

	if err != nil {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse(fmt.Sprintf("Invalid category data: %s", err.Error())))
		return
	}

	//dummy batchID
	batchID := bson.NewObjectID()

	createdData, err := di.CategoryDataService.CreateCategoryData(reqCtx, categoryIDObj, data, data, filePath, batchID)

	if err != nil {
		log.Printf("Error creating category data: %v", err)
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Failed to create category data"))
		return
	}
	c.JSON(201, createdData)
}

func GetCategoryData(c *gin.Context) (*map[string]any, error) {
	jsonData := c.PostForm("data")
	if jsonData == "" {
		return nil, errors.New("data not found in form data")
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, errors.New("invalid JSON format")
	}
	return &data, nil
}

func GetCategoryRawData(c *gin.Context) (*map[string]any, error) {
	var jsonData map[string]any
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		return nil, errors.New("invalid JSON format")
	}

	if jsonData == nil {
		return nil, errors.New("data not found in form data")
	}
	data := jsonData
	return &data, nil
}

func SaveFileToDisk(appCtx *app.AppContext, reqCtx *app.RequestContext, c *gin.Context) (string, error) {
	// Retrieve the file from the form data
	file, err := c.FormFile("file")
	if err != nil {
		return "", errors.New("file not found in form data")
	}

	dst := file_utils.GetFilePath(appCtx, reqCtx, appCtx.Config.UPLOAD_DIR, file.Filename, 7, 0)
	if dst == "" {
		return "", errors.New("failed to upload file, unique file path not found")
	}

	// Save the file
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}
	return dst, nil
}

func updateCategoryDataEndpoint(c *gin.Context) {
	// User uploads the file along with category data
	_, exists := app.GetAppContext(c)
	if !exists {
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Application context not found"))
		return
	}

	di, found := app_di.GetAppDI(c)
	if !found {
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Failed to get application DI"))
		return
	}

	reqCtx := app.GetRequestCtx(c)
	if reqCtx.IsZero() {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Current user not found"))
		return
	}

	categoryID := c.Param("categoryID")
	if categoryID == "" {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Category ID is required"))
		return
	}

	categoryIDObj, err := bson.ObjectIDFromHex(categoryID)
	if err != nil {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Invalid category ID format"))
		return
	}
	dataID := c.Param("dataID")
	if dataID == "" {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Data ID is required"))
		return
	}

	dataIDObj, err := bson.ObjectIDFromHex(dataID)
	if err != nil {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Invalid data ID format"))
		return
	}
	//get json request body not form dat
	//GetRawData(c)

	if err != nil {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse(fmt.Sprintf("Invalid category data: %s", err.Error())))
		return
	}

	data, err := GetCategoryRawData(c)
	if err != nil {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse(fmt.Sprintf("Invalid category data: %s", err.Error())))
		return
	}

	updatedData, err := di.CategoryDataService.UpdateCategoryData(reqCtx, categoryIDObj, dataIDObj, data)

	if err != nil {
		log.Printf("Error updating category data: %v", err)
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Failed to update category data"))
		return
	}
	c.JSON(200, utils.NewOkResponse(updatedData))
}
