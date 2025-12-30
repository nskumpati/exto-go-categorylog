package service_test

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/mock/gomock"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/mocks"
	"github.com/gaeaglobal/exto/server/model"
	"github.com/gaeaglobal/exto/server/service"
)

func TestCreateAndSetupIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 1. Initialize AppContext with a test database connection.
	appCtx := app.NewMockAppContext()
	reqCtx := &app.RequestContext{}

	identityMockRepo := mocks.NewMockIdentityRepo(ctrl)
	newIdentityID := bson.NewObjectID()
	newOrgID := bson.NewObjectID()
	identityMockRepo.EXPECT().
		CreateIdentity(gomock.Eq(reqCtx), matchCreateIdentity(&model.CreateIdentity{
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
		})).
		Return(&model.Identity{
			Base: model.Base{
				ID:        newIdentityID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				CreatedBy: newIdentityID,
				UpdatedBy: newIdentityID,
			},
			Email:        "test@example.com",
			FirstName:    "Test",
			LastName:     "User",
			CurrentOrgID: newOrgID,
		}, nil)
	identityMockRepo.EXPECT().
		IsIdentityExists(gomock.Eq(reqCtx), gomock.Eq("test@example.com")).
		Return(false, nil)

	orgMockRepo := mocks.NewMockOrganizationRepo(ctrl)
	orgMockRepo.EXPECT().
		CreateOrganization(gomock.Eq(reqCtx), matchCreateOrganization(&model.CreateOrganization{
			Name: "test@example.com's Organization",
			Slug: "test@example.com-org",
		}), gomock.Eq(newIdentityID)).
		Return(&model.Organization{
			Base: model.Base{
				ID:        newOrgID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				CreatedBy: newIdentityID,
				UpdatedBy: newIdentityID,
			},
			Name:        "test@example.com's Organization",
			Slug:        "test@example.com-org",
			IsActive:    true,
			OwnerID:     newIdentityID,
			ScanCounter: 0,
		}, nil)
	orgMockRepo.EXPECT().
		IsOrganizationExists(gomock.Eq(reqCtx), gomock.Eq("test@example.com's Organization")).
		Return(false, nil)

	orgMockRepo.EXPECT().GetOrganizationCount(gomock.Eq(reqCtx)).
		Return(int64(0), nil)

	userMockRepo := mocks.NewMockUserRepo(ctrl)
	userMockRepo.EXPECT().
		CreateUser(gomock.Eq(reqCtx), gomock.Eq(newIdentityID), gomock.Eq(&model.CreateUser{
			Email:          "test@example.com",
			FirstName:      "Test",
			LastName:       "User",
			Role:           model.RoleOrganizationAdmin,
			IdentityID:     newIdentityID,
			OrganizationID: newOrgID,
		})).
		Return(&model.User{
			Base: model.Base{
				ID:        bson.NewObjectID(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				CreatedBy: newIdentityID,
				UpdatedBy: newIdentityID,
			},
			Email:          "test@example.com",
			FirstName:      "Test",
			LastName:       "User",
			Role:           model.RoleOrganizationAdmin,
			IdentityID:     newIdentityID,
			OrganizationID: newOrgID,
		}, nil)
	userMockRepo.EXPECT().IsUserExists(gomock.Eq(reqCtx), gomock.Eq(newOrgID), gomock.Eq("test@example.com")).
		Return(false, nil)

	mockSession := mocks.NewMockMongoSession(ctrl)
	mockSession.EXPECT().StartTransaction().Return(nil)
	mockSession.EXPECT().EndSession(gomock.Any())
	mockSession.EXPECT().AbortTransaction(gomock.Any()).Times(0)
	mockSession.EXPECT().CommitTransaction(gomock.Any()).Return(nil)

	mockSessionProvider := mocks.NewMockSessionProvider(ctrl)
	mockSessionProvider.EXPECT().StartSession().
		Return(mockSession, nil)

	identityService := service.NewIdentityService(identityMockRepo)
	orgService := service.NewOrganizationService(orgMockRepo)
	userService := service.NewUserService(mockSessionProvider, userMockRepo, identityService, orgService, service.NewGoogleSheetService())

	// 3. Call CreateAndSetupIdentity
	email := "test@example.com"
	firstName := "Test"
	lastName := "User"

	req, err := userService.CreateAndSetupIdentity(appCtx, reqCtx, email, firstName, lastName)
	user := req.User
	if err != nil {
		t.Fatalf("Failed to create and setup identity: %v", err)
	}

	// 4. Verify the user is created
	if user.IsZero() {
		t.Fatal("Expected user to be created")
	}

	if user.Email != email || user.FirstName != firstName || user.LastName != lastName {
		t.Fatal("User details do not match")
	}

	if user.Role != model.RoleOrganizationAdmin {
		t.Fatalf("Expected user role to be %s, got %s", model.RoleOrganizationAdmin, user.Role)
	}

	if user.IdentityID != newIdentityID {
		t.Fatalf("Expected user IdentityID to be %s, got %s", newIdentityID.Hex(), user.IdentityID.Hex())
	}

	if user.OrganizationID != newOrgID {
		t.Fatalf("Expected user OrganizationID to be %s, got %s", newOrgID.Hex(), user.OrganizationID.Hex())
	}

}

func matchCreateIdentity(expected *model.CreateIdentity) gomock.Matcher {
	return gomock.Cond(func(x interface{}) bool {
		arg, ok := x.(*model.CreateIdentity)
		if !ok {
			// The argument is not of the expected type.
			return false
		}

		// Here, you explicitly compare the fields you care about.
		return arg.Email == expected.Email &&
			arg.FirstName == expected.FirstName &&
			arg.LastName == expected.LastName &&
			!arg.CurrentOrgID.IsZero()
	})
}

func matchCreateOrganization(expected *model.CreateOrganization) gomock.Matcher {
	return gomock.Cond(func(x interface{}) bool {
		arg, ok := x.(*model.CreateOrganization)
		if !ok {
			// The argument is not of the expected type.
			return false
		}

		// Here, you explicitly compare the fields you care about.
		return arg.Name == expected.Name &&
			arg.Slug == expected.Slug &&
			!arg.ID.IsZero()
	})
}
