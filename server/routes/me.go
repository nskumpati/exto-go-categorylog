package routes

import (
	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/app_di"
	"github.com/gaeaglobal/exto/server/utils"
	"github.com/gin-gonic/gin"
)

func AddMeRoutes(router *gin.RouterGroup) {
	router.GET("/me", meHandler)
	router.DELETE("/me", deleteMeHandler)
}

func meHandler(c *gin.Context) {
	reqCtx := app.GetRequestCtx(c)
	if reqCtx.IsZero() {
		c.JSON(401, utils.NewErrorResponse("Unauthorized"))
		return
	}
	c.JSON(200, utils.NewOkResponse(reqCtx))
}

func deleteMeHandler(c *gin.Context) {
	reqCtx := app.GetRequestCtx(c)
	if reqCtx.IsZero() {
		c.JSON(401, utils.NewErrorResponse("Unauthorized"))
		return
	}

	di, found := app_di.GetAppDI(c)
	if !found {
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Failed to get application dependencies"))
		return
	}

	if err := di.UserService.DeleteUser(reqCtx); err != nil {
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Failed to delete user: "+err.Error()))
		return
	}
	c.JSON(200, utils.NewOkResponse("User deleted successfully"))
}
