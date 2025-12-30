package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/stripe/stripe-go/v82"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/model"
)

type PaymentService struct {
	sc                  *stripe.Client
	orgService          *OrganizationService
	subscriptionService *SubscriptionService
}

func NewPaymentService(sc *stripe.Client, orgService *OrganizationService, subscriptionService *SubscriptionService) *PaymentService {
	return &PaymentService{
		sc:                  sc,
		orgService:          orgService,
		subscriptionService: subscriptionService,
	}
}

type PaymentIntentResponse struct {
	Customer             *stripe.Customer      `json:"customer"`
	CustomerEphemeralKey *stripe.EphemeralKey  `json:"customer_ephemeral_key"`
	PaymentIntent        *stripe.PaymentIntent `json:"payment_intent"`
}

type SetupIntentResponse struct {
	CustomerID                 string `json:"customer_id"`
	CustomerEphemeralKeySecret string `json:"customer_ephemeral_key_secret"`
	SetupIntentClientSecret    string `json:"setup_intent_client_secret"`
}

func (s *PaymentService) CreateSetupIntent(reqCtx *app.RequestContext, billing *model.Billing) (*SetupIntentResponse, error) {
	stripCustomer, err := s.createCustomerIfNotExists(reqCtx, billing)
	if err != nil {
		return nil, err
	}

	// Update billing in organization
	_, err = s.orgService.UpdateOrganizationBilling(reqCtx, billing)
	if err != nil {
		log.Printf("Error updating organization billing: %v", err)
		return nil, errors.New("failed to update organization billing")
	}

	params := &stripe.SetupIntentCreateParams{
		Customer:    stripe.String(stripCustomer.ID),
		Usage:       stripe.String(string(stripe.SetupIntentUsageOffSession)),
		Description: stripe.String("Setup Intent for Exto Platform"),
	}
	setupIntend, err := s.sc.V1SetupIntents.Create(context.Background(), params)
	if err != nil {
		log.Printf("Error creating Setup Intent: %v", err)
		return nil, errors.New("failed to create setup intent")
	}

	ephemeralKey, err := s.getCustomerEphemeralKey(stripCustomer.ID)
	if err != nil {
		log.Printf("Error getting customer ephemeral key: %v", err)
		return nil, errors.New("failed to get customer ephemeral key")
	}
	log.Printf("Successfully created setup intent: %s\n", setupIntend.ID)
	return &SetupIntentResponse{
		CustomerID:                 stripCustomer.ID,
		CustomerEphemeralKeySecret: ephemeralKey.Secret,
		SetupIntentClientSecret:    setupIntend.ClientSecret,
	}, nil
}

func (s *PaymentService) CreateSubscription(reqCtx *app.RequestContext) (*model.Subscription, error) {
	// Price ID for the subscription plan - price_1S86biJgAi1y3OSXfjFBwHO3
	org, err := s.orgService.GetOrganizationByID(reqCtx, reqCtx.Org.ID)
	if err != nil {
		return nil, err
	}
	if org.StripeCustomerId == "" {
		return nil, errors.New("stripe customer ID is missing for the organization")
	}

	billingPlanId := "price_1S86biJgAi1y3OSXfjFBwHO3"

	subscriptionParams := &stripe.SubscriptionCreateParams{
		Customer:          stripe.String(org.StripeCustomerId),
		OffSession:        stripe.Bool(true),
		TrialPeriodDays:   stripe.Int64(7),
		ProrationBehavior: stripe.String("none"),
		PaymentBehavior:   stripe.String("error_if_incomplete"),
		Items: []*stripe.SubscriptionCreateItemParams{
			{
				Price: stripe.String(billingPlanId),
			},
		},
		Discounts: []*stripe.SubscriptionCreateDiscountParams{
			{
				Coupon: stripe.String("UD3PLsTI"),
			},
		},
		Expand: []*string{
			stripe.String("latest_invoice.payment_intent"),
		},
	}
	sub, err := s.sc.V1Subscriptions.Create(context.Background(), subscriptionParams)
	if err != nil {
		log.Printf("Error creating subscription: %v", err)
		return nil, errors.New("failed to create subscription")
	}

	appSub, err := s.subscriptionService.CreateSubscription(reqCtx, &model.CreateSubscription{
		OrganizationID:  org.ID,
		StripeSubID:     sub.ID,
		StartDate:       time.Now(),
		TrialPeriodDays: 7,
		BillingCycle:    model.BillingCycleMonthly,
	})

	if err != nil {
		log.Printf("Error saving subscription to DB: %v", err)
		return nil, errors.New("failed to save subscription")
	}

	log.Printf("Successfully created subscription with ID: %s for Org: %s\n", sub.ID, org.Name)
	log.Printf("App Subscription record created with ID: %s\n", appSub.ID.Hex())
	return appSub, nil
}

func (s *PaymentService) CancelSubscription(reqCtx *app.RequestContext) error {
	//
	return nil
}

func (s *PaymentService) GetSubscriptionStatus(reqCtx *app.RequestContext) (string, error) {
	//
	return "", nil
}

func (s *PaymentService) ListInvoices(reqCtx *app.RequestContext) ([]*stripe.Invoice, error) {
	//
	return nil, nil
}

func (s *PaymentService) GetInvoicePDF(reqCtx *app.RequestContext, invoiceID string) (string, error) {
	//
	return "", nil
}

func (s *PaymentService) createCustomerIfNotExists(reqCtx *app.RequestContext, billing *model.Billing) (*stripe.Customer, error) {
	org, err := s.orgService.GetOrganizationByID(reqCtx, reqCtx.Org.ID)
	if err != nil {
		return nil, err
	}
	if org.StripeCustomerId != "" {
		// Customer already exists
		customer, err := s.sc.V1Customers.Retrieve(context.Background(), org.StripeCustomerId, nil)
		if err != nil {
			log.Printf("Error retrieving customer: %v", err)
			return nil, err
		}
		return customer, nil
	}

	// Create a new Customer
	customerParams := &stripe.CustomerCreateParams{
		Description: stripe.String("Exto Platform Customer for Org " + org.Name),
		Name:        stripe.String(org.Name),
		Email:       stripe.String(billing.Email),
		Address: &stripe.AddressParams{
			Line1: stripe.String(billing.StreetAddress),
			Line2: stripe.String(fmt.Sprintf("%s, %s", billing.City, billing.State)),
		},
		Phone: stripe.String(billing.Phone),
		Metadata: map[string]string{
			"organization_id": org.ID.Hex(),
		},
	}
	newCustomer, err := s.sc.V1Customers.Create(context.Background(), customerParams)
	if err != nil {
		log.Printf("Error creating customer: %v", err)
		return nil, err
	}

	_, err = s.orgService.UpdateStripeCustomerID(reqCtx, org.ID, newCustomer.ID)
	if err != nil {
		log.Printf("Error updating organization with Stripe Customer ID: %v", err)
		return nil, err
	}
	return newCustomer, nil
}

// to get customer ephemeral key
func (s *PaymentService) getCustomerEphemeralKey(stripeCustomerID string) (*stripe.EphemeralKey, error) {

	ephemeralKeyParams := &stripe.EphemeralKeyCreateParams{
		Customer: stripe.String(stripeCustomerID),
	}
	ephemeralKey, err := s.sc.V1EphemeralKeys.Create(context.Background(), ephemeralKeyParams)
	if err != nil {
		log.Fatalf("Error creating ephemeral key: %v", err)
	}
	return ephemeralKey, nil
}

func CreatePaymentIntent(appCtx *app.AppContext, amount int64, currency string) (*PaymentIntentResponse, error) {
	sc := stripe.NewClient(appCtx.Config.STRIPE_API_KEY)
	// ====================================================================
	// 1. Create a new Customer
	// This represents your user in the Stripe system.
	// ====================================================================

	// Load customer information
	customerResult := sc.V1Customers.List(context.Background(), &stripe.CustomerListParams{
		Email: stripe.String("devops@exto360.com"),
	})

	var c *stripe.Customer
	for customer, err := range customerResult {
		if err != nil {
			log.Printf("Error retrieving customer: %v", err)
			continue
		}
		c = customer
		log.Printf("Found existing customer: %v\n", c)
	}

	if c != nil {
		log.Printf("Using existing customer: %v\n", c)
	} else {
		log.Println("Creating a new Stripe customer...")
		customerParams := &stripe.CustomerCreateParams{
			Description: stripe.String("Exto Platform Test customer"),
			Name:        stripe.String("Exto Platform Test"),
			Email:       stripe.String("devops@exto360.com"),
		}
		newCustomer, err := sc.V1Customers.Create(context.Background(), customerParams)
		if err != nil {
			log.Fatalf("Error creating customer: %v", err)
		}
		log.Printf("Successfully created customer with ID: %s\n", newCustomer.ID)
		c = newCustomer
	}

	// 2. Create Customer ephemeral key
	log.Println("Creating a new ephemeral key for the customer...")
	ephemeralKeyParams := &stripe.EphemeralKeyCreateParams{
		Customer: stripe.String(c.ID),
	}
	ephemeralKey, err := sc.V1EphemeralKeys.Create(context.Background(), ephemeralKeyParams)
	if err != nil {
		log.Fatalf("Error creating ephemeral key: %v", err)
	}
	log.Printf("Successfully created ephemeral key with ID: %s\n", ephemeralKey.ID)

	// ====================================================================
	// 3. Create a Payment Intent
	// This represents your intent to collect a payment from a customer.
	// The client_secret is what you send to your front-end to confirm the payment.
	// ====================================================================
	log.Println("Creating a payment intent...")
	paymentIntentParams := &stripe.PaymentIntentCreateParams{
		Amount:   stripe.Int64(2000), // Amount in cents, e.g., $20.00
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		Customer: stripe.String(c.ID),
	}
	newPaymentIntent, err := sc.V1PaymentIntents.Create(context.Background(), paymentIntentParams)
	if err != nil {
		log.Fatalf("Error creating payment intent: %v", err)
		return nil, errors.New("failed to create payment intent")
	}
	// The client_secret should be sent to the mobile/web client.
	log.Printf("Successfully created payment intent with client secret: %s\n", newPaymentIntent.ClientSecret)

	log.Println("All done! The `client_secret` and `ephemeral_key.secret` should be sent to your client-side application.")
	return &PaymentIntentResponse{
		Customer:             c,
		CustomerEphemeralKey: ephemeralKey,
		PaymentIntent:        newPaymentIntent,
	}, nil
}

func (s *PaymentService) GetFreeTrialInfo(reqCtx *app.RequestContext) (*model.GetFreeTrialInfo, error) {
	orgInfo, err := s.orgService.GetOrganizationByID(reqCtx, reqCtx.Org.ID)
	log.Print("s", orgInfo.CreatedAt, orgInfo.LastActiveAt)

	trialInfo, err := s.subscriptionService.r.GetFreeTrialInfo(reqCtx)
	if err != nil {
		return nil, err
	}
	remainingScans := 250 - trialInfo.RecordCount
	trialEndDate := orgInfo.CreatedAt.AddDate(0, 0, 30)

	nowDate := time.Date(time.Now().UTC().Year(), time.Now().UTC().Month(), time.Now().UTC().Day(), 0, 0, 0, 0, time.UTC)
	trialDate := time.Date(trialEndDate.UTC().Year(), trialEndDate.UTC().Month(), trialEndDate.UTC().Day(), 0, 0, 0, 0, time.UTC)
	daysLeft := int(trialDate.Sub(nowDate).Hours() / 24)

	if daysLeft < 0 {
		daysLeft = 0
	}
	if remainingScans < 0 {
		remainingScans = 0
	}

	return &model.GetFreeTrialInfo{
		DaysLeft:       daysLeft,
		RemainingScans: remainingScans,
	}, nil
}
