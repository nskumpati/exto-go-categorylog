package service

import (
	"context"
	"errors"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/model"
	"github.com/gaeaglobal/exto/server/repo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetUserOrgsByEmail(appCtx app.AppContext, email string) (*[]model.Organization, error) {
	identitiesCol := appCtx.DB.GetCoreDatabase().Collection("identities")
	var identity model.Identity

	// Set up projection
	identityOpts := options.FindOne().SetProjection(bson.M{"_id": 1, "email": 1, "is_active": 1})
	err := identitiesCol.FindOne(context.Background(), bson.M{"email": email}, identityOpts).Decode(&identity)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !identity.IsActive {
		return nil, errors.New("user is not active")
	}

	// load users from org
	userOrgCol := appCtx.DB.GetCoreDatabase().Collection("users")
	userOpts := options.Find().SetProjection(bson.M{"_id": 1, "org_id": 1, "is_active": 1})
	cursor, err := userOrgCol.Find(context.Background(), bson.M{"identity_id": identity.ID}, userOpts)
	if err != nil {
		return nil, errors.New("user don't belong to any organization")
	}
	var orgIds []bson.ObjectID
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		if !user.IsActive {
			continue // Skip inactive users
		}
		orgIds = append(orgIds, user.OrganizationID)
	}

	if len(orgIds) == 0 {
		return nil, errors.New("user don't belong to any organization")
	}

	orgCursor, err := appCtx.DB.GetCoreDatabase().Collection("organizations").Find(context.Background(), bson.M{"_id": bson.M{"$in": orgIds}})
	if err != nil {
		return nil, errors.New("failed to retrieve organizations")
	}
	defer orgCursor.Close(context.Background())
	var orgs []model.Organization
	for orgCursor.Next(context.Background()) {
		var org model.Organization
		if err := orgCursor.Decode(&org); err != nil {
			return nil, err
		}
		if org.IsActive {
			orgs = append(orgs, org)
		}
	}

	return &orgs, nil
}

type IdentityService struct {
	repo repo.IdentityRepo
}

func NewIdentityService(repo repo.IdentityRepo) *IdentityService {
	return &IdentityService{
		repo: repo,
	}
}

func (s *IdentityService) GetIdentityByEmail(reqCtx *app.RequestContext, email string) (*model.Identity, error) {
	return s.repo.GetIdentityByEmail(reqCtx, email)
}

func (s *IdentityService) CreateIdentity(reqCtx *app.RequestContext, identity *model.CreateIdentity) (*model.Identity, error) {
	exists, err := s.repo.IsIdentityExists(reqCtx, identity.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("identity already exists")
	}
	return s.repo.CreateIdentity(reqCtx, identity)
}

func (s *IdentityService) DeleteIdentity(reqCtx *app.RequestContext) error {
	return s.repo.DeleteIdentity(reqCtx)
}
