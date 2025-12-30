package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/app_di"
	"github.com/gaeaglobal/exto/server/db"
	"github.com/gaeaglobal/exto/server/routes"
	"github.com/gin-gonic/gin"
)

func main() {

	// --- Load Configuration Once at Startup ---
	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Println("--- Application Configuration Loaded ---")

	// Connect to the database here if needed

	appDB, err := db.ConnectToDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	log.Println("--- Database Connection Established ---")

	// Initialize the AppContext with the loaded configuration.
	appCtx := &app.AppContext{
		Config: cfg,
		DB:     appDB,
	}

	appDI := app_di.NewAppDI(appCtx)

	// Set Gin to release mode if not in debug mode
	if !cfg.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Creates a router without any middleware by default
	router := getRouter(appCtx, appDI)

	addr := fmt.Sprintf(":%d", cfg.AppPort)
	log.Printf("Starting server on %s\n", addr)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Cleanup all caches before shutting down.
	appDI.Close()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

func getRouter(appCtx *app.AppContext, appDI *app_di.AppDI) *gin.Engine {
	// Creates a router without any middleware by default
	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(app.AppContextMiddleware(appCtx))
	router.Use(app_di.AppDIMiddleware(appDI))
	// Auth Routes
	auth := router.Group("/auth")
	routes.AddAuthenticationRoutes(auth)

	// Protected Routes
	protected := router.Group("/v1")
	protected.Use(app.AppAuthzMiddleware())
	protected.Use(app.LastActiveMiddleware(appCtx))
	routes.AddMeRoutes(protected)
	routes.AddOpenAiRoutes(protected)
	routes.AddCategoryRoutes(protected)
	routes.AddCategoryDataRoutes(protected)
	routes.AddScanHistoryRoutes(protected)
	routes.AddExportRoutes(protected)
	routes.AddBillingRoutes(protected)
	routes.AddScanRoutes(protected)
	routes.AddBatchRoutes(protected)
	routes.AddPaymentRoutes(protected)

	return router
}
