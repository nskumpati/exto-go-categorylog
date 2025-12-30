package app

import (
	"github.com/gaeaglobal/exto/server/db"
	"github.com/gin-gonic/gin"
)

type Org struct {
	ID   string
	Name string
}

type CurrentUser struct {
	ID           string
	Email        string
	FirstName    string
	LastName     string
	Organization []Org
}

type Services struct {
	// User *service.UserService
}

// AppContext holds dependencies that handlers might need, including the configuration.
// This is a common pattern to inject dependencies into handlers.
type AppContext struct {
	Config *Config
	DB     *db.AppDB
	Count  int64
}

func NewMockAppContext() *AppContext {
	return &AppContext{
		Config: NewMockConfig(),
		DB:     db.NewMockAppDB(),
	}
}

// appCtxKey is the key used to store and retrieve AppContext from Gin's context.
const appCtxKey = "app_context"

// AppContextMiddleware creates a Gin middleware that injects the AppContext into the request context.
func AppContextMiddleware(appCtx *AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(appCtxKey, appCtx)
		c.Next() // Proceed to the next handler
	}
}

// GetAppContext is a helper function to retrieve the AppContext from Gin's context.
// It returns the AppContext and a boolean indicating if it was found.
func GetAppContext(c *gin.Context) (*AppContext, bool) {
	// Retrieve the AppContext from the request's context.
	val, exists := c.Get(appCtxKey)
	if !exists {
		return nil, false
	}
	appCtx, ok := val.(*AppContext)
	if !ok {
		return nil, false
	}
	return appCtx, true
}
