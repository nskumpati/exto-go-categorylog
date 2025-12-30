/*
 Copyright 2025 The Exto Project Solutions, Inc.
 All rights reserved.

 Author: Vimalraj Arumugam

 This software is the confidential and proprietary product of The Exto Project Solutions, Inc.
 and is protected by copyright and trade secret law.
 Use, reproduction, and distribution of this software is strictly forbidden.

 For more details, please refer to the LICENSE file in the root directory of this project.

*/

//go:generate mockgen -source=user_repo.go -destination=../mocks/mock_user_repo.go -package=mocks -copyright_file=../../copy_right.txt

package repo

import (
	"errors"
	"log"
	"time"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/db"
	"github.com/gaeaglobal/exto/server/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserRepo interface {
	IBaseRepo
	CreateUser(reqCtx *app.RequestContext, createdByIdentityId bson.ObjectID, user *model.CreateUser) (*model.User, error)
	GetUserByID(reqCtx *app.RequestContext, id bson.ObjectID) (*model.User, error)
	UpdateUser(reqCtx *app.RequestContext, id bson.ObjectID, user *model.UpdateUser) (*model.User, error)
	GetUsersByOrgID(reqCtx *app.RequestContext, orgID bson.ObjectID, pageReq *app.PageRequest) (*app.PageResponse[*model.User], error)
	IsUserExists(reqCtx *app.RequestContext, orgID bson.ObjectID, email string) (bool, error)
	DeleteUserByOrgIDs(reqCtx *app.RequestContext, orgIDs []bson.ObjectID) error
}

type MongoUserRepo struct {
	BaseRepo
}

func NewMongoUserRepository(appDB *db.AppDB) *MongoUserRepo {
	return &MongoUserRepo{
		BaseRepo: BaseRepo{
			cname: "users",
			appDB: appDB,
		},
	}
}

func (r *MongoUserRepo) CreateUser(reqCtx *app.RequestContext, createdByIdentityId bson.ObjectID, user *model.CreateUser) (*model.User, error) {
	var col = r.GetCollection()

	usr := &model.User{
		Base: model.Base{
			ID:        bson.NewObjectID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: createdByIdentityId,
			UpdatedBy: createdByIdentityId,
		},
		IdentityID:     user.IdentityID,
		OrganizationID: user.OrganizationID,
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Role:           user.Role,
		IsActive:       true,
	}
	ctx, cancel := db.GetDBContext()
	defer cancel()
	result, err := col.InsertOne(ctx, usr)
	if err != nil {
		log.Printf("failed to create user: %v", err)
		return nil, errors.New("failed to create user")
	}
	usr.ID = result.InsertedID.(bson.ObjectID)
	return usr, nil
}

func (r *MongoUserRepo) GetUserByID(reqCtx *app.RequestContext, id bson.ObjectID) (*model.User, error) {
	var col = r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()
	var user model.User
	err := col.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		log.Printf("failed to get user by id: %v", err)
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (r *MongoUserRepo) UpdateUser(reqCtx *app.RequestContext, id bson.ObjectID, user *model.UpdateUser) (*model.User, error) {
	var col = r.GetCollection()

	usr := &model.User{
		Base: model.Base{
			UpdatedAt: time.Now(),
			UpdatedBy: reqCtx.User.IdentityID,
		},
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		IsActive:  user.IsActive,
	}

	ctx, cancel := db.GetDBContext()
	defer cancel()
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	updatedUser := &model.User{}
	result := col.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": usr}, opts)
	if result.Err() != nil {
		log.Printf("failed to update user: %v", result.Err())
		return nil, errors.New("failed to update user")
	}
	if err := result.Decode(updatedUser); err != nil {
		log.Printf("failed to decode updated user: %v", err)
		return nil, errors.New("failed to update user")
	}
	return updatedUser, nil
}

func (r *MongoUserRepo) GetUsersByOrgID(reqCtx *app.RequestContext, orgID bson.ObjectID, pageReq *app.PageRequest) (*app.PageResponse[*model.User], error) {
	var col = r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()
	opts := pageReq.ToFindOptions()

	filter := bson.M{"org_id": orgID}

	var (
		users    []*model.User
		count    int64
		findErr  error
		countErr error
	)

	findDone := make(chan struct{})
	countDone := make(chan struct{})

	// Find documents in parallel
	go func() {
		defer close(findDone)
		cursor, err := col.Find(ctx, filter, opts)
		if err != nil {
			findErr = err
			return
		}
		defer cursor.Close(ctx)
		if err := cursor.All(ctx, &users); err != nil {
			findErr = err
			return
		}
	}()

	// Count documents in parallel
	go func() {
		defer close(countDone)
		count, countErr = col.CountDocuments(ctx, filter)
	}()

	<-countDone
	<-findDone

	if findErr != nil {
		log.Printf("failed to count users by org id: %v", findErr)
		return nil, errors.New("failed to get users count")
	}
	if countErr != nil {
		log.Printf("failed to get users by org id: %v", countErr)
		return nil, errors.New("failed to get users")
	}

	if users == nil {
		users = []*model.User{}
	}

	return app.NewPageResponse(count, pageReq.Page, pageReq.PageSize, users), nil
}

func (r *MongoUserRepo) IsUserExists(reqCtx *app.RequestContext, orgID bson.ObjectID, email string) (bool, error) {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()
	count, err := col.CountDocuments(ctx, bson.M{"org_id": orgID, "email": email})
	if err != nil {
		log.Println("Error checking user existence:", err)
		return false, errors.New("error checking user existence")
	}
	return count > 0, nil
}

func (r *MongoUserRepo) DeleteUserByOrgIDs(reqCtx *app.RequestContext, orgIDs []bson.ObjectID) error {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()

	_, err := col.DeleteMany(ctx, bson.M{"org_id": bson.M{"$in": orgIDs}})
	if err != nil {
		log.Printf("error deleting users by org ids: %v", err)
		return errors.New("failed to delete users")
	}

	db.DeleteCacheUser(reqCtx.User.Email)
	return nil
}
