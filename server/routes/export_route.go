package routes

import (
	"log"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/app_di"
	"github.com/gaeaglobal/exto/server/utils"
	"github.com/gin-gonic/gin"
)

func AddExportRoutes(router *gin.RouterGroup) {
	router.GET("/export/:scan_history_id", exportHandler)
}

func exportHandler(c *gin.Context) {
	reqCtx := app.GetRequestCtx(c)
	if reqCtx.IsZero() {
		c.JSON(500, utils.NewErrorResponse("failed to get request context"))
		return
	}
	scanHistoryID := c.Param("scan_history_id")

	di, found := app_di.GetAppDI(c)
	if !found {
		c.JSON(500, utils.NewErrorResponse("failed to get app DI"))
		return
	}

	result, err := di.ExportService.ExportAsExcel(reqCtx, scanHistoryID)
	if err != nil {
		log.Printf("failed to export as excel: %v", err)
		c.JSON(500, utils.NewErrorResponse("failed to export as excel"))
		return
	}

	// Downloadable file
	c.Header("Content-Disposition", "attachment; filename="+result.FileName)
	c.File(result.FilePath)
}
