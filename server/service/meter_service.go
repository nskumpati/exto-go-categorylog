package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/model"
	"github.com/gaeaglobal/exto/server/repo"
	"github.com/stripe/stripe-go/v82"
)

type MeterService struct {
	sc         *stripe.Client
	repo       repo.MeterEventRepository
	orgService *OrganizationService
}

func NewMeterService(sc *stripe.Client, repo repo.MeterEventRepository, orgService *OrganizationService) *MeterService {
	return &MeterService{
		sc:         sc,
		repo:       repo,
		orgService: orgService,
	}
}

func (s *MeterService) IncrementMeterEvent(reqCtx *app.RequestContext, name string, value int) (*model.MeterEvent, error) {
	org, err := s.orgService.GetOrganizationByID(reqCtx, reqCtx.Org.ID)
	if err != nil {
		return nil, err
	}
	if org.StripeCustomerId == "" {
		return nil, nil
	}
	params := &stripe.BillingMeterEventCreateParams{
		EventName: stripe.String(name),
		Payload: map[string]string{
			"value":              fmt.Sprintf("%d", value),
			"stripe_customer_id": org.StripeCustomerId,
		},
	}
	result, err := s.sc.V1BillingMeterEvents.Create(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if result.Identifier == "" {
		return nil, errors.New("failed to create meter event in stripe")
	}
	return s.repo.CreateMeterEvent(reqCtx, name, value, org.StripeCustomerId)
}
