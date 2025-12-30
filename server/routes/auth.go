package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/app_di"
	"github.com/gaeaglobal/exto/server/utils"
)

func AddAuthenticationRoutes(auth *gin.RouterGroup) {
	auth.POST("/sign-up", signUpEndpoint)
	auth.POST("/login", loginEndpoint)
}

type SignUpRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Provider  string `json:"provider" binding:"required"`
}

func signUpEndpoint(c *gin.Context) {

	var requestBody SignUpRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Invalid request body"))
		return
	}
	// Check if identity already exists
	appCtx, exists := app.GetAppContext(c)
	if !exists {
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Application context not found"))
		return
	}

	di, found := app_di.GetAppDI(c)
	if !found {
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Failed to get application dependencies"))
		return
	}

	var email string
	if requestBody.Provider == "google" {
		// Use existing Google token validation
		tokenUser, err := app.GetUserFromToken(c)
		if err != nil {
			c.AbortWithStatusJSON(400, utils.NewErrorResponse("Invalid Google ID token"))
			return
		}
		email = tokenUser.Email
	} else if requestBody.Provider == "microsoft" {
		tokenUser, err := app.GetUserFromMicrosoftToken(c)
		if err != nil {
			c.AbortWithStatusJSON(400, utils.NewErrorResponse("Invalid Microsoft ID token"))
			return
		}
		email = tokenUser.Email
	} else {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Unsupported provider"))
		return
	}

	req, err := di.UserService.CreateAndSetupIdentity(appCtx, &app.RequestContext{}, email, requestBody.FirstName, requestBody.LastName)
	if err != nil {
		c.AbortWithStatusJSON(500, utils.NewErrorResponse("Failed to create new identity. cause: "+err.Error()))
		return
	}

	c.JSON(201, utils.NewOkResponse(req))
}

func loginEndpoint(c *gin.Context) {
	reqCtx := app.GetRequestCtx(c)
	if reqCtx.IsZero() {
		c.AbortWithStatusJSON(400, utils.NewErrorResponse("Failed to get current user"))
		return
	}
	c.JSON(200, utils.NewOkResponse(reqCtx.User))
}
