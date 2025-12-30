package service

import (
	"errors"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/model"
	"github.com/gaeaglobal/exto/server/repo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrganizationService struct {
	repo repo.OrganizationRepo
}

func NewOrganizationService(repo repo.OrganizationRepo) *OrganizationService {
	return &OrganizationService{
		repo: repo,
	}
}

func (s *OrganizationService) CreateOrganization(reqCtx *app.RequestContext, org *model.CreateOrganization, createdByIdentityId bson.ObjectID) (*model.Organization, error) {
	if exists, err := s.repo.IsOrganizationExists(reqCtx, org.Name); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("organization already exists")
	}
	return s.repo.CreateOrganization(reqCtx, org, createdByIdentityId)
}

func (s *OrganizationService) GetOrganizationByID(reqCtx *app.RequestContext, orgID bson.ObjectID) (*model.Organization, error) {
	return s.repo.GetOrganizationByID(reqCtx, orgID)
}

func (s *OrganizationService) UpdateStripeCustomerID(reqCtx *app.RequestContext, orgID bson.ObjectID, stripeCustomerID string) (*model.Organization, error) {
	return s.repo.UpdateOrganization(reqCtx, orgID, &model.UpdateOrganization{
		StripeCustomerId: stripeCustomerID,
	})
}

func (s *OrganizationService) GetOrganizationCount(reqCtx *app.RequestContext) (int64, error) {
	return s.repo.GetOrganizationCount(reqCtx)
}

func (s *OrganizationService) GenerateNextScanCode(reqCtx *app.RequestContext) (string, error) {
	return s.repo.GenerateNextScanCode(reqCtx)
}

func (s *OrganizationService) UpdateOrganizationBilling(reqCtx *app.RequestContext, billing *model.Billing) (*model.Organization, error) {
	return s.repo.UpdateOrganization(reqCtx, reqCtx.Org.ID, &model.UpdateOrganization{
		Billing: *billing,
	})
}

func (s *OrganizationService) DeleteOrganizationByOwner(reqCtx *app.RequestContext) ([]*model.Organization, error) {
	return s.repo.DeleteOrganizationByOwner(reqCtx)
}
