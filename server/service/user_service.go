package service

import (
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/db"
	"github.com/gaeaglobal/exto/server/model"
	"github.com/gaeaglobal/exto/server/repo"
)

type UserService struct {
	dbSessionProvider  db.SessionProvider
	repo               repo.UserRepo
	identityService    *IdentityService
	orgService         *OrganizationService
	GoogleSheetService *GoogleSheetService
}

func NewUserService(dbSessionProvider db.SessionProvider, repo repo.UserRepo, identityService *IdentityService, orgService *OrganizationService, GoogleSheetService *GoogleSheetService) *UserService {
	return &UserService{
		dbSessionProvider:  dbSessionProvider,
		repo:               repo,
		identityService:    identityService,
		orgService:         orgService,
		GoogleSheetService: GoogleSheetService,
	}
}

func (s *UserService) CreateAndSetupIdentity(appCtx *app.AppContext, reqCtx *app.RequestContext, email string, firstName, lastName string) (*app.RequestContext, error) {

	session, err := s.dbSessionProvider.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(reqCtx.Context())

	err = session.StartTransaction()
	if err != nil {
		return nil, err
	}

	newOrgID := bson.NewObjectID()
	// Create Identity
	identity, err := s.identityService.CreateIdentity(reqCtx, &model.CreateIdentity{
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		CurrentOrgID: newOrgID,
	})
	if err != nil {
		_ = session.AbortTransaction(reqCtx.Context())
		return nil, err
	}

	log.Println("Created Identity:", identity.ID.Hex())

	// Create Organization
	orgCount, err := s.orgService.GetOrganizationCount(reqCtx)
	if err != nil {
		log.Println("Error getting organization count:", err)
		orgCount = 0 // Default to 0 if there's an error
	}
	organization, err := s.orgService.CreateOrganization(reqCtx, &model.CreateOrganization{
		ID:   newOrgID,
		Name: fmt.Sprintf("%s's Organization", email),
		Slug: fmt.Sprintf("org_%d", orgCount+1),
	}, identity.ID)
	if err != nil {
		_ = session.AbortTransaction(reqCtx.Context())
		return nil, err
	}

	// Create User
	if exists, err := s.repo.IsUserExists(reqCtx, organization.ID, email); err != nil {
		_ = session.AbortTransaction(reqCtx.Context())
		return nil, err
	} else if exists {
		_ = session.AbortTransaction(reqCtx.Context())
		return nil, fmt.Errorf("user with email %s already exists in organization %s", email, organization.Name)
	}
	user, err := s.repo.CreateUser(reqCtx, identity.ID, &model.CreateUser{
		Email:          email,
		FirstName:      firstName,
		LastName:       lastName,
		Role:           model.RoleOrganizationAdmin,
		IdentityID:     identity.ID,
		OrganizationID: organization.ID,
	})

	s.GoogleSheetService.SyncSheet("example-sheet-id", organization.Name, organization.Slug, user.FirstName, user.LastName, user.Email, identity.CreatedAt)

	if err != nil {
		_ = session.AbortTransaction(reqCtx.Context())
		return nil, err
	}

	if err := session.CommitTransaction(reqCtx.Context()); err != nil {
		return nil, err
	}

	_reqCtx := &app.RequestContext{
		User: app.RequestUser{
			ID:             user.ID,
			Email:          user.Email,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			Role:           user.Role,
			OrganizationID: user.OrganizationID,
			IdentityID:     user.IdentityID,
			IsActive:       user.IsActive,
		},
		Org: app.RequestOrg{
			ID:   organization.ID,
			Name: organization.Name,
			Slug: organization.Slug,
		},
	}

	return _reqCtx, nil
}

func (s *UserService) DeleteUser(reqCtx *app.RequestContext) error {
	// Delete the organization if owned by the user
	orgs, err := s.orgService.DeleteOrganizationByOwner(reqCtx)
	if err != nil {
		return err
	}

	// Delete the users in the organizations
	var orgIDs []bson.ObjectID
	for _, org := range orgs {
		orgIDs = append(orgIDs, org.ID)
	}
	if len(orgIDs) > 0 {
		if err := s.repo.DeleteUserByOrgIDs(reqCtx, orgIDs); err != nil {
			return err
		}
	}

	// Delete the identity
	if err := s.identityService.DeleteIdentity(reqCtx); err != nil {
		return err
	}

	return nil
}
