package routes

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/app_di"
	"github.com/gaeaglobal/exto/server/model"
	"github.com/gaeaglobal/exto/server/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func AddBatchRoutes(router *gin.RouterGroup) {
	router.POST("/batch", newBatchHandler)
	router.PATCH("/batch/:batchID", updateBatchStatusHandler)
}

func newBatchHandler(c *gin.Context) {
	reqCtx := app.GetRequestCtx(c)
	if reqCtx.IsZero() {
		c.JSON(500, utils.NewErrorResponse("failed to get request context"))
		return
	}

	di, found := app_di.GetAppDI(c)
	if !found {
		c.JSON(500, utils.NewErrorResponse("failed to get app DI"))
		return
	}

	//generate a batch name with 6 digit random alphanumeric string
	batchName := fmt.Sprintf("Batch-%06d", rand.Intn(1000000))
	newBatch, err := di.BatchService.CreateBatch(reqCtx, &model.CreateBatchRequest{Name: batchName})
	if err != nil {
		log.Printf("failed to create batch: %v", err)
		c.JSON(500, utils.NewErrorResponse("failed to create new batch"))
		return
	}

	c.JSON(http.StatusOK, utils.NewOkResponse(newBatch))
}

//implement update batch status

func updateBatchStatusHandler(c *gin.Context) {
	reqCtx := app.GetRequestCtx(c)
	if reqCtx.IsZero() {
		c.JSON(500, utils.NewErrorResponse("failed to get request context"))
		return
	}

	di, found := app_di.GetAppDI(c)
	if !found {
		c.JSON(500, utils.NewErrorResponse("failed to get app DI"))
		return
	}

	batchID := c.Param("batchID")
	if batchID == "" {
		c.JSON(400, utils.NewErrorResponse("batch ID is required"))
		return
	}

	batchObjID, err := bson.ObjectIDFromHex(batchID)
	if err != nil {
		c.JSON(400, utils.NewErrorResponse("invalid batch ID format"))
		return
	}

	updatedBatch, err := di.BatchService.UpdateBatchStatus(reqCtx, batchObjID, "Closed")
	if err != nil {
		log.Printf("failed to update batch status: %v", err)
		c.JSON(500, utils.NewErrorResponse("failed to update batch status"))
		return
	}

	c.JSON(http.StatusOK, utils.NewOkResponse(updatedBatch))
}
