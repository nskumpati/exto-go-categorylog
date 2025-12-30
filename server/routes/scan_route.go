/*
 Copyright 2025 The Exto Project Solutions, Inc.
 All rights reserved.

 Author: Kalaiselven Moorthy

 This software is the confidential and proprietary product of The Exto Project Solutions, Inc.
 and is protected by copyright and trade secret law.
 Use, reproduction, and distribution of this software is strictly forbidden.

 For more details, please refer to the LICENSE file in the root directory of this project.
*/

package routes

import (
	"log"
	"net/http"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/app_di"
	"github.com/gaeaglobal/exto/server/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ExtractResponseParams struct {
	BatchID        string `json:"batchID"`
	CategoryDataID string `json:"categoryDataID"`
	ScanHistoryID  string `json:"scanHistoryID"`
	ScanCode       string `json:"scanCode"`
	RawData        any    `json:"raw_data"`
}

func AddScanRoutes(router *gin.RouterGroup) {
	router.POST("/scan/document", newScanImageEndpoint)
}

func newScanImageEndpoint(c *gin.Context) {
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

	categoryID := c.PostForm("categoryID")
	log.Printf("Received categoryID: %s\n", categoryID)

	if categoryID == "" {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Category ID is required"))
		return
	}

	categoryObjID, err := bson.ObjectIDFromHex(categoryID)
	if err != nil {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Invalid category ID format"))
		return
	}

	// get batchID from formdata
	batchID := c.PostForm("batchID")
	log.Printf("Received batchID: %s\n", batchID)

	if batchID == "" {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Batch ID is required"))
		return
	}

	batchObjID, err := bson.ObjectIDFromHex(batchID)
	if err != nil {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Invalid batch ID format"))
		return
	}

	// Perform the scan using ScanService
	scanService := di.ScanService
	scanResult, err := scanService.PerformScan(appCtx, reqCtx, c, categoryObjID, batchObjID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to perform scan: "+err.Error()))
		return
	}
	// Return the scan result
	c.JSON(http.StatusOK, utils.NewOkResponse(ExtractResponseParams{
		BatchID:        scanResult.BatchID,
		CategoryDataID: scanResult.CategoryDataID,
		ScanHistoryID:  scanResult.ScanHistoryID,
		ScanCode:       scanResult.ScanCode,
		RawData:        scanResult.Data,
	}))

}
