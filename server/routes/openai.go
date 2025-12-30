package routes

import (
	"fmt"
	"net/http"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/app_di"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ExtractRequest struct {
	Base64Image string `json:"base64Image"`
	CategoryID  string `json:"categoryID"`
}

type ExtractResponse struct {
	ExtractedData     any     `json:"extractedData"`
	ConfidenceScores  any     `json:"confidenceScores"`
	AverageConfidence float64 `json:"averageConfidence"`
}

func AddOpenAiRoutes(router *gin.RouterGroup) {
	router.POST("/extract", ExtractHandler)
}

func ExtractHandler(c *gin.Context) {
	var req ExtractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	categoryObjID, err := bson.ObjectIDFromHex(req.CategoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid categoryID"})
		return
	}

	reqCtx := app.GetRequestCtx(c)
	if reqCtx.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current user not found"})
		return
	}

	di, found := app_di.GetAppDI(c)
	if !found {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get application DI"})
		return
	}

	result, err := di.OpenAIService.ExtractDocumentData(reqCtx, categoryObjID, req.Base64Image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAI extraction failed: " + err.Error()})
		return
	}

	resultStr, ok := result.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assert result as string"})
		return
	}
	extractedMap, err := di.OpenAIService.ConvertKeyValueToMap(resultStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert result to map: " + err.Error()})
		return
	}

	confidenceMap, err := di.OpenAIService.ConvertConfidenceScoresToMap(resultStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert confidence scores to map: " + err.Error()})
		return
	}
	avg := di.OpenAIService.CalculateAverageConfidence(confidenceMap)
	fmt.Println("Average confidence:", avg)

	c.JSON(http.StatusOK, ExtractResponse{
		ExtractedData:     extractedMap,
		ConfidenceScores:  confidenceMap,
		AverageConfidence: avg,
	})
}
