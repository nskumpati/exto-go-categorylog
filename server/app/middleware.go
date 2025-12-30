package app

import (
	"context"
	"log"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var activeOrgCache *ristretto.Cache[string, time.Time]

func init() {
	cache, err := ristretto.NewCache(&ristretto.Config[string, time.Time]{
		NumCounters: 100, // number of keys to track frequency of (10x MaxCost).
		MaxCost:     100, // 100 MB
		BufferItems: 5,   // Number of keys per Get buffer
	})
	if err != nil {
		log.Fatalf("failed to initialize activeOrgCache: %v", err)
	}
	activeOrgCache = cache
}

func LastActiveMiddleware(appCtx *AppContext) gin.HandlerFunc {
	// Define a time threshold (e.g., 10 minutes)
	const updateThreshold = 10 * time.Minute

	return func(c *gin.Context) {
		reqCtx := GetRequestCtx(c)
		if reqCtx == nil || reqCtx.User.IsZero() {
			c.Next()
			return
		}

		// Check the in-memory cache first.
		orgID := reqCtx.Org.ID.Hex()
		if lastUpdated, ok := activeOrgCache.Get(orgID); ok {
			if time.Since(lastUpdated) < updateThreshold {
				// The timestamp in cache is still fresh, do nothing and proceed.
				c.Next()
				return
			}
		}

		// If the cache is expired or the organization is not in it, perform the database update.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filter := bson.M{"_id": reqCtx.Org.ID}
		update := bson.M{"$set": bson.M{"last_active_at": time.Now()}}

		organizationsCollection := appCtx.DB.GetCoreDatabase().Collection("organizations")
		result, err := organizationsCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("Failed to perform update for organization %s: %v", orgID, err)
		} else if result.ModifiedCount > 0 {
			// Update was successful, update the in-memory cache.
			activeOrgCache.Set(orgID, time.Now(), 1)
			activeOrgCache.Wait()
			log.Printf("Successfully updated lastActiveAt for organization %s", orgID)
		} else {
			// The document wasn't modified, likely because it didn't exist or was updated by another process.
			log.Printf("Document for organization %s not modified.", orgID)
		}

		// Continue to the next handler.
		c.Next()
	}
}
