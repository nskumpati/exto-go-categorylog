package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/app_di"
	"github.com/gaeaglobal/exto/server/model"
	"github.com/gaeaglobal/exto/server/utils"
)

func AddPaymentRoutes(r *gin.RouterGroup) {
	r.GET("/subscription", getMySubscription)
	r.POST("/payment/setup", setupPayment)
	r.POST("/payment/subscribe", createSubscription)
	r.POST("/payment/cancel", cancelSubscription)
	r.GET("/payment/free-trial", getFreeTrialInfo)
}

func setupPayment(c *gin.Context) {
	reqCtx := app.GetRequestCtx(c)
	if reqCtx == nil {
		c.JSON(500, utils.NewErrorResponse("missing request context"))
		return
	}

	di, found := app_di.GetAppDI(c)
	if !found {
		c.JSON(500, utils.NewErrorResponse("failed to get app DI"))
		return
	}

	billing := &model.Billing{}
	if err := c.ShouldBindJSON(billing); err != nil {
		c.JSON(400, utils.NewErrorResponse("invalid request payload"))
		return
	}

	resp, err := di.PaymentService.CreateSetupIntent(reqCtx, billing)
	if err != nil {
		c.JSON(500, utils.NewErrorResponse("failed to create setup intent"))
		return
	}

	c.JSON(200, utils.NewOkResponse(resp))
}

func createSubscription(c *gin.Context) {
	reqCtx := app.GetRequestCtx(c)
	if reqCtx == nil {
		c.JSON(500, utils.NewErrorResponse("missing request context"))
		return
	}

	di, found := app_di.GetAppDI(c)
	if !found {
		c.JSON(500, utils.NewErrorResponse("failed to get app DI"))
		return
	}

	sub, err := di.PaymentService.CreateSubscription(reqCtx)
	if err != nil {
		c.JSON(500, utils.NewErrorResponse("failed to create subscription"))
		return
	}

	c.JSON(200, utils.NewOkResponse(sub))
}

func cancelSubscription(c *gin.Context) {
	reqCtx := app.GetRequestCtx(c)
	if reqCtx == nil {
		c.JSON(500, utils.NewErrorResponse("missing request context"))
		return
	}

	di, found := app_di.GetAppDI(c)
	if !found {
		c.JSON(500, utils.NewErrorResponse("failed to get app DI"))
		return
	}

	err := di.PaymentService.CancelSubscription(reqCtx)
	if err != nil {
		c.JSON(500, utils.NewErrorResponse("failed to cancel subscription"))
		return
	}

	c.JSON(200, utils.NewOkResponse("subscription cancelled successfully"))
}

func getMySubscription(c *gin.Context) {
	reqCtx := app.GetRequestCtx(c)
	if reqCtx == nil {
		c.JSON(500, utils.NewErrorResponse("missing request context"))
		return
	}

	di, found := app_di.GetAppDI(c)
	if !found {
		c.JSON(500, utils.NewErrorResponse("failed to get app DI"))
		return
	}

	sub, err := di.SubscriptionService.GetMySubscription(reqCtx)
	if err != nil {
		if err.Error() == "no current subscription found" {
			c.JSON(404, utils.NewErrorResponse("no current subscription found"))
			return
		}
		c.JSON(500, utils.NewErrorResponse("failed to get subscription"))
		return
	}

	c.JSON(200, utils.NewOkResponse(sub))
}

func getFreeTrialInfo(c *gin.Context) {
	reqCtx := app.GetRequestCtx(c)
	if reqCtx == nil {
		c.JSON(500, utils.NewErrorResponse("missing request context"))
		return
	}

	di, found := app_di.GetAppDI(c)
	if !found {
		c.JSON(500, utils.NewErrorResponse("failed to get app DI"))
		return
	}

	info, err := di.PaymentService.GetFreeTrialInfo(reqCtx)
	if err != nil {
		c.JSON(500, utils.NewErrorResponse("failed to get free trial info"))
		return
	}

	c.JSON(200, utils.NewOkResponse(info))
}
