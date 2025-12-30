/*
 Copyright 2025 The Exto Project Solutions, Inc.
 All rights reserved.

 Author: Vimalraj Arumugam

 This software is the confidential and proprietary product of The Exto Project Solutions, Inc.
 and is protected by copyright and trade secret law.
 Use, reproduction, and distribution of this software is strictly forbidden.

 For more details, please refer to the LICENSE file in the root directory of this project.
*/

//go:generate mockgen -source=identity_repo.go -destination=../mocks/mock_identity_repo.go -package=mocks -copyright_file=../../copy_right.txt
package repo

import (
	"errors"
	"log"
	"time"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/db"
	"github.com/gaeaglobal/exto/server/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type IdentityRepo interface {
	IBaseRepo
	CreateIdentity(reqCtx *app.RequestContext, identity *model.CreateIdentity) (*model.Identity, error)
	GetIdentityByEmail(reqCtx *app.RequestContext, email string) (*model.Identity, error)
	UpdateIdentity(reqCtx *app.RequestContext, id bson.ObjectID, identity *model.UpdateIdentity) (*model.Identity, error)
	IsIdentityExists(reqCtx *app.RequestContext, email string) (bool, error)
	DeleteIdentity(reqCtx *app.RequestContext) error
}

type MongoIdentityRepo struct {
	BaseRepo
}

func NewIdentityRepository(appDB *db.AppDB) *MongoIdentityRepo {
	return &MongoIdentityRepo{
		BaseRepo: BaseRepo{
			cname: "identities",
			appDB: appDB,
		},
	}
}

func (r *MongoIdentityRepo) CreateIdentity(reqCtx *app.RequestContext, identity *model.CreateIdentity) (*model.Identity, error) {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()

	identityID := bson.NewObjectID()
	createdIdentityID := identityID
	if !reqCtx.User.IsZero() {
		createdIdentityID = reqCtx.User.IdentityID
	}

	newIdentity := model.Identity{
		Base: model.Base{
			ID:        identityID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: createdIdentityID,
			UpdatedBy: createdIdentityID,
		},
		Email:        identity.Email,
		FirstName:    identity.FirstName,
		LastName:     identity.LastName,
		CurrentOrgID: identity.CurrentOrgID,
		IsActive:     true,
	}

	result, err := col.InsertOne(ctx, newIdentity)
	if err != nil {
		log.Println("Error creating identity:", err)
		return nil, errors.New("error creating identity")
	}
	if !result.Acknowledged {
		return nil, errors.New("error creating identity")
	}

	return &newIdentity, nil
}

func (r *MongoIdentityRepo) GetIdentityByEmail(reqCtx *app.RequestContext, email string) (*model.Identity, error) {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()
	var identity model.Identity
	result := col.FindOne(ctx, bson.M{"email": email})
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, errors.New("identity not found")
		}
		log.Println("Error retrieving identity:", result.Err())
		return nil, errors.New("error retrieving identity")
	}
	if err := result.Decode(&identity); err != nil {
		log.Println("Error decoding identity:", err)
		return nil, errors.New("error decoding identity")
	}
	// check if identity empty
	if identity.IsZero() {
		return nil, errors.New("identity not found")
	}
	return &identity, nil
}

func (r *MongoIdentityRepo) UpdateIdentity(reqCtx *app.RequestContext, id bson.ObjectID, identity *model.UpdateIdentity) (*model.Identity, error) {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := col.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": identity}, opts)
	if result.Err() != nil {
		log.Println("Error updating identity:", result.Err())
		return nil, errors.New("error updating identity")
	}
	if !result.Acknowledged {
		return nil, errors.New("error updating identity")
	}
	var updatedIdentity model.Identity
	if err := result.Decode(&updatedIdentity); err != nil {
		log.Println("Error decoding updated identity:", err)
		return nil, errors.New("error decoding updated identity")
	}
	// check if identity empty
	if updatedIdentity.IsZero() {
		return nil, errors.New("identity not found after update")
	}
	return &updatedIdentity, nil
}

func (r *MongoIdentityRepo) IsIdentityExists(reqCtx *app.RequestContext, email string) (bool, error) {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()
	count, err := col.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		log.Println("Error checking identity existence:", err)
		return false, errors.New("error checking identity existence")
	}
	return count > 0, nil
}

func (r *MongoIdentityRepo) DeleteIdentity(reqCtx *app.RequestContext) error {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()

	result, err := col.DeleteOne(ctx, bson.M{"_id": reqCtx.User.IdentityID})
	if err != nil {
		log.Println("Error deleting identity:", err)
		return errors.New("error deleting identity")
	}
	if result.DeletedCount == 0 {
		return errors.New("no identity found to delete")
	}
	return nil
}
