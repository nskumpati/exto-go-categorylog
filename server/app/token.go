package app

import (
	"errors"
	"log"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/gin-gonic/gin"
)

type GoogleTokenUser struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Exp           int64  `json:"exp"`
	Hd            string `json:"hd"`
}

func GetUserFromToken(c *gin.Context) (*GoogleTokenUser, error) {
	appCtx, exists := GetAppContext(c)
	if !exists {
		return nil, errors.New("appContext not found")
	}
	token := c.GetHeader("Authorization")
	if token == "" {
		return nil, errors.New("authorization token is required")
	}

	// Remove "Bearer " prefix if it exists
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	// token will be provided in Request body
	id_token := token
	// Handle sign-up logic here
	tokenPayload, err := idtoken.Validate(c.Request.Context(), id_token, appCtx.Config.GOOGLE_MOBILE_CLIENT_ID)
	if err != nil {
		tokenPayload, err = idtoken.Validate(c.Request.Context(), id_token, appCtx.Config.GOOGLE_CLIENT_ID)
		if err != nil {
			log.Printf("Error validating ID token: %v", err)
			return nil, errors.New(`invalid ID token`)
		}
	}
	// You can access the token payload to get user information
	email, _ := tokenPayload.Claims["email"].(string)
	emailVerified, _ := tokenPayload.Claims["email_verified"].(bool)
	expFloat, _ := tokenPayload.Claims["exp"].(float64)
	hd, _ := tokenPayload.Claims["hd"].(string)

	return &GoogleTokenUser{
		Email:         email,
		EmailVerified: emailVerified,
		Exp:           int64(expFloat),
		Hd:            hd,
	}, nil
}
