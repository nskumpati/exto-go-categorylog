package app

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gaeaglobal/exto/server/model"
	"github.com/gaeaglobal/exto/server/utils"
)

const appRequestKey = "app_request"

type RequestUser struct {
	ID             bson.ObjectID  `json:"id"`
	Email          string         `json:"email"`
	FirstName      string         `json:"first_name"`
	LastName       string         `json:"last_name"`
	Role           model.UserRole `json:"role"`
	OrganizationID bson.ObjectID  `json:"organization_id"`
	IdentityID     bson.ObjectID  `json:"identity_id"`
	IsActive       bool           `json:"is_active"`
}

type RequestOrg struct {
	ID   bson.ObjectID `json:"id"`
	Name string        `json:"name"`
	Slug string        `json:"slug"`
}

func (r *RequestUser) IsZero() bool {
	return r == nil || r.Email == ""
}

type RequestContext struct {
	ctx  context.Context `json:"-"`
	User RequestUser     `json:"user"`
	Org  RequestOrg      `json:"organization"`
}

func (r *RequestContext) IsZero() bool {
	return r == nil || r.User.IsZero()
}

func (r *RequestContext) WithContext(ctx context.Context) *RequestContext {
	r.ctx = ctx
	return r
}

func (r *RequestContext) WithUser(user RequestUser) *RequestContext {
	r.User = user
	return r
}

func (r *RequestContext) Context() context.Context {
	if r.ctx == nil {
		return context.Background()
	}
	return r.ctx
}

func NewMockRequestContext() *RequestContext {

	idenID, err := bson.ObjectIDFromHex("64a7f0f4f0f0f0f0f0f0f0f1")
	if err != nil {
		log.Fatal("Failed to create ObjectID:", err)
	}

	userID, err := bson.ObjectIDFromHex("64a7f0f4f0f0f0f0f0f0f0f2")
	if err != nil {
		log.Fatal("Failed to create ObjectID:", err)
	}

	orgID, err := bson.ObjectIDFromHex("64a7f0f4f0f0f0f0f0f0f0f3")
	if err != nil {
		log.Fatal("Failed to create ObjectID:", err)
	}

	user := &RequestUser{
		ID:             userID,
		Email:          "test@example.com",
		FirstName:      "Test",
		LastName:       "User",
		Role:           model.RoleSuperAdmin,
		OrganizationID: orgID,
		IdentityID:     idenID,
		IsActive:       true,
	}
	org := &RequestOrg{
		ID:   orgID,
		Name: "Test Organization",
		Slug: "org_test-1",
	}
	return newRequestContext(user, org)
}

func newRequestContext(user *RequestUser, org *RequestOrg) *RequestContext {
	ctx := context.Background()
	return &RequestContext{
		ctx:  ctx,
		User: *user,
		Org:  *org,
	}
}

func GetRequestCtx(c *gin.Context) *RequestContext {
	val, exists := c.Get(appRequestKey)
	if !exists {
		return nil
	}
	ctx, ok := val.(*RequestContext)
	if !ok {
		return nil
	}
	return ctx
}

func AppAuthzMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.GetHeader("x-auth-provider")
		if provider == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, utils.NewErrorResponse("X-Auth-Provider header is required"))
			return
		}
		var tokenUserEmail string
		if provider == "google" {
			tokenUser, err := GetUserFromToken(c)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, utils.NewErrorResponse("Unauthorized"))
				return
			}
			tokenUserEmail = tokenUser.Email
		} else if provider == "microsoft" {
			tokenUser, err := GetUserFromMicrosoftToken(c)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, utils.NewErrorResponse("Unauthorized"))
				return
			}
			tokenUserEmail = tokenUser.Email
		}

		// Implement your authorization logic here.
		// For example, you might check if the user is logged in and has the right permissions.
		appCtx, exists := GetAppContext(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, utils.NewErrorResponse("AppContext not found"))
			return
		}

		cached, err := appCtx.DB.GetUserByEmail(tokenUserEmail)
		if err != nil {
			log.Println("Error retrieving user:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.NewErrorResponse("User not found or not authorized"))
			return
		}

		org := &RequestOrg{
			ID:   cached.Org.ID,
			Name: cached.Org.Name,
			Slug: cached.Org.Slug,
		}
		c.Set(appRequestKey, newRequestContext(&RequestUser{
			ID:             cached.User.ID,
			Email:          cached.Email,
			FirstName:      cached.User.FirstName,
			LastName:       cached.User.LastName,
			Role:           cached.User.Role,
			OrganizationID: cached.User.OrganizationID,
			IdentityID:     cached.User.IdentityID,
			IsActive:       cached.User.IsActive,
		}, org))
		c.Next()
	}
}
