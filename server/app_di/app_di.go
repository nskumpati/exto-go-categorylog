package app_di

import (
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v82"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/db"
	"github.com/gaeaglobal/exto/server/repo"
	"github.com/gaeaglobal/exto/server/service"
)

const appDIKey = "app_di"

type AppDI struct {
	UserService         *service.UserService
	CategoryService     *service.CategoryService
	FormatService       *service.FormatService
	CategoryDataService *service.CategoryDataService
	OrganizationService *service.OrganizationService
	ScanHistoryService  *service.ScanHistoryService
	ExportService       *service.ExportService
	OpenAIService       *service.OpenAIService
	BatchService        *service.BatchService
	ScanService         *service.ScanService
	PaymentService      *service.PaymentService
	SubscriptionService *service.SubscriptionService
}

func NewAppDI(appCtx *app.AppContext) *AppDI {
	// Create Repositories
	identityRepo := repo.NewIdentityRepository(appCtx.DB)
	userRepo := repo.NewMongoUserRepository(appCtx.DB)
	orgRepo := repo.NewOrganizationRepository(appCtx.DB)
	categoryRepo := repo.NewCategoryRepository(appCtx.DB)
	formatRepo := repo.NewFormatRepository(appCtx.DB)

	scanHistoryRepo := repo.NewScanHistoryRepository(appCtx.DB)
	categoryDataRepo := repo.NewCategoryDataRepository(appCtx.DB)
	meterEventRepo := repo.NewMeterEventRepository(appCtx.DB)

	batchRepo := repo.NewBatchRepository(appCtx.DB)
	subscriptionRepo := repo.NewSubscriptionRepository(appCtx.DB)

	dbSessionProvider := db.NewSessionProvider(appCtx.DB.Client)

	// Create Services
	identityService := service.NewIdentityService(identityRepo)
	orgService := service.NewOrganizationService(orgRepo)
	userService := service.NewUserService(dbSessionProvider, userRepo, identityService, orgService, service.NewGoogleSheetService())
	categoryService := service.NewCategoryService(categoryRepo)
	formatService := service.NewFormatService(formatRepo)

	scanHistoryService := service.NewScanHistoryService(dbSessionProvider, scanHistoryRepo)
	categoryDataService := service.NewCategoryDataService(categoryDataRepo, categoryService, orgService, scanHistoryService)

	exportService := service.NewExportService(appCtx, orgService, scanHistoryService, categoryService, categoryDataService)
	openAIService := service.NewOpenAIService(formatService)

	batchService := service.NewBatchService(dbSessionProvider, batchRepo)

	// Stripe Client Initialization
	sc := stripe.NewClient(appCtx.Config.STRIPE_API_KEY)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo)
	paymentService := service.NewPaymentService(sc, orgService, subscriptionService)
	meterService := service.NewMeterService(sc, meterEventRepo, orgService)

	scanService := service.NewScanService(appCtx, orgService, batchService, openAIService, scanHistoryService, categoryDataService, meterService)
	// Return the AppDI instance
	return &AppDI{
		UserService:         userService,
		OrganizationService: orgService,
		CategoryService:     categoryService,
		FormatService:       formatService,
		ScanHistoryService:  scanHistoryService,
		CategoryDataService: categoryDataService,
		ExportService:       exportService,
		OpenAIService:       openAIService,
		BatchService:        batchService,
		ScanService:         scanService,
		PaymentService:      paymentService,
		SubscriptionService: subscriptionService,
	}
}

func (di *AppDI) Close() {

	di.CategoryService.Close()
}

func AppDIMiddleware(appDI *AppDI) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(appDIKey, appDI)
		c.Next()
	}
}

func GetAppDI(c *gin.Context) (*AppDI, bool) {
	val, found := c.Get(appDIKey)
	if !found {
		return nil, false
	}
	appDI, ok := val.(*AppDI)
	return appDI, ok
}
