/*
 Copyright 2025 The Exto Project Solutions, Inc.
 All rights reserved.

 Author: Vimalraj Arumugam

 This software is the confidential and proprietary product of The Exto Project Solutions, Inc.
 and is protected by copyright and trade secret law.
 Use, reproduction, and distribution of this software is strictly forbidden.

 For more details, please refer to the LICENSE file in the root directory of this project.
*/

package routes

import (
	"log"
	"strings"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/app_di"
	"github.com/gaeaglobal/exto/server/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func AddScanHistoryRoutes(router *gin.RouterGroup) {
	router.GET("/scan-history", getScanHistory)
	router.GET("/scan-history/data/:scanHistoryID", getScannedDocumentDataEndpoint)
	router.GET("/scan-history/document/:scanHistoryID", getScannedDocumentEndpoint)
}

func getScannedDocumentEndpoint(c *gin.Context) {
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

	scanHistoryID := c.Param("scanHistoryID")
	if scanHistoryID == "" {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Scan History ID is required"))
		return
	}

	scanHistoryIDObj, err := bson.ObjectIDFromHex(scanHistoryID)
	if err != nil {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Invalid scan history ID format"))
		return
	}

	scanHistory, err := di.ScanHistoryService.GetScanHistoryByID(reqCtx, scanHistoryIDObj)

	if err != nil {
		log.Printf("Error retrieving scan history: %v", err)
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Failed to retrieve scan history"))
		return
	}
	if scanHistory == nil {
		c.AbortWithStatusJSON(404, utils.NewErrorResponse("Scan history not found"))
		return
	}

	categoryIDObj := scanHistory.CategoryID
	dataIDObj := scanHistory.CategoryDataID

	categoryData, err := di.CategoryDataService.GetCategoryDataByID(reqCtx, categoryIDObj, dataIDObj)
	if err != nil || categoryData == nil {
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Failed to retrieve category data"))
		return
	}
	if len(categoryData.DocumentPaths) == 0 || categoryData.DocumentPaths[0] == "" {
		c.AbortWithStatusJSON(404, utils.NewErrorResponse("Document not found"))
		return
	}

	documentPath := categoryData.DocumentPaths[0]
	documentName := documentPath[strings.LastIndex(documentPath, "/")+1:]
	c.Header("Content-Disposition", "attachment; filename="+documentName)
	c.File(documentPath)
}

func getScannedDocumentDataEndpoint(c *gin.Context) {
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

	scanHistoryID := c.Param("scanHistoryID")
	if scanHistoryID == "" {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Scan History ID is required"))
		return
	}

	scanHistoryIDObj, err := bson.ObjectIDFromHex(scanHistoryID)
	if err != nil {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Invalid scan history ID format"))
		return
	}

	scanHistory, err := di.ScanHistoryService.GetScanHistoryByID(reqCtx, scanHistoryIDObj)

	if err != nil {
		log.Printf("Error retrieving scan history: %v", err)
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Failed to retrieve scan history"))
		return
	}
	if scanHistory == nil {
		c.AbortWithStatusJSON(404, utils.NewErrorResponse("Scan history not found"))
		return
	}

	categoryIDObj := scanHistory.CategoryID
	dataIDObj := scanHistory.CategoryDataID

	categoryData, err := di.CategoryDataService.GetCategoryDataByID(reqCtx, categoryIDObj, dataIDObj)
	if err != nil {
		log.Printf("Error retrieving category data: %v", err)
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Failed to retrieve category data"))
		return
	}
	if categoryData == nil {
		c.AbortWithStatusJSON(404, utils.NewErrorResponse("Category data not found"))
		return
	}
	c.JSON(200, utils.NewOkResponse(categoryData))
}

func getScanHistory(c *gin.Context) {
	reqCtx := app.GetRequestCtx(c)
	if reqCtx.IsZero() {
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Failed to get request context"))
		return
	}

	di, found := app_di.GetAppDI(c)
	if !found {
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Failed to get application dependencies"))
		return
	}

	// Retrieve scan history from the service
	scanHistory, err := di.ScanHistoryService.ListScanHistories(reqCtx, app.NewPageRequest(c))
	if err != nil {
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Failed to retrieve scan history"))
		return
	}

	c.JSON(200, utils.NewOkResponse(scanHistory))
}
