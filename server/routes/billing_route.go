package routes

import (
	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/service"
	"github.com/gaeaglobal/exto/server/utils"
	"github.com/gin-gonic/gin"
)

func AddBillingRoutes(r *gin.RouterGroup) {
	r.POST("/billing/payment", paymentRoute)
}

func paymentRoute(c *gin.Context) {
	appCtx, exists := app.GetAppContext(c)
	if !exists {
		c.AbortWithStatusJSON(500, gin.H{"error": "Application context not found"})
		return
	}

	resp, err := service.CreatePaymentIntent(appCtx, 1000, "usd")
	if err != nil {
		c.JSON(500, utils.NewErrorResponse("Failed to create payment intent: "+err.Error()))
		return
	}
	c.JSON(200, resp)
}
