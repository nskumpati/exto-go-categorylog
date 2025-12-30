package routes

import (
	"log"
	"net/http"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/app_di"
	"github.com/gaeaglobal/exto/server/utils"
	"github.com/gin-gonic/gin"
)

func AddCategoryRoutes(router *gin.RouterGroup) {
	router.GET("/categories", getAllCategoriesEndpoint)
}

func getAllCategoriesEndpoint(c *gin.Context) {

	reqCtx := app.GetRequestCtx(c)
	if reqCtx.IsZero() {
		log.Printf("request context is missing")
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.NewErrorResponse("Internal server error"))
		return
	}

	// Get the application DI container from the context
	di, found := app_di.GetAppDI(c)
	if !found {
		log.Printf("Error retrieving app DI")
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to retrieve app DI"))
		return
	}

	// Call the service function to get all categories
	categories, err := di.CategoryService.ListCategories(reqCtx, app.NewPageRequest(c))
	if err != nil {
		log.Printf("Error retrieving categories: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to retrieve categories"))
		return
	}

	c.JSON(http.StatusOK, categories)
}
