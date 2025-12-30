/*
 Copyright 2025 The Exto Project Solutions, Inc.
 All rights reserved.

 Author: Vimalraj Arumugam

 This software is the confidential and proprietary product of The Exto Project Solutions, Inc.
 and is protected by copyright and trade secret law.
 Use, reproduction, and distribution of this software is strictly forbidden.

 For more details, please refer to the LICENSE file in the root directory of this project.
*/

//go:generate mockgen -source=organization_repo.go -destination=../mocks/mock_organization_repo.go -package=mocks -copyright_file=../../copy_right.txt

package repo

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/db"
	"github.com/gaeaglobal/exto/server/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type OrganizationRepo interface {
	IBaseRepo
	CreateOrganization(reqCtx *app.RequestContext, org *model.CreateOrganization, createdByIdentity bson.ObjectID) (*model.Organization, error)
	GetOrganizationByID(reqCtx *app.RequestContext, id bson.ObjectID) (*model.Organization, error)
	GetOrganizations(reqCtx *app.RequestContext, filter bson.M) ([]*model.Organization, error)
	UpdateOrganization(reqCtx *app.RequestContext, id bson.ObjectID, org *model.UpdateOrganization) (*model.Organization, error)
	ListOrganizations(reqCtx *app.RequestContext, pageReq *app.PageRequest) (*app.PageResponse[*model.Organization], error)
	IsOrganizationExists(reqCtx *app.RequestContext, name string) (bool, error)
	GetOrganizationCount(reqCtx *app.RequestContext) (int64, error)
	GenerateNextScanCode(reqCtx *app.RequestContext) (string, error)
	DeleteOrganizationByOwner(reqCtx *app.RequestContext) ([]*model.Organization, error)
}

type MongoOrganizationRepo struct {
	BaseRepo
}

func NewOrganizationRepository(appDB *db.AppDB) *MongoOrganizationRepo {
	return &MongoOrganizationRepo{
		BaseRepo: BaseRepo{
			cname: "organizations",
			appDB: appDB,
		},
	}
}

func (r *MongoOrganizationRepo) CreateOrganization(reqCtx *app.RequestContext, org *model.CreateOrganization, createdByIdentity bson.ObjectID) (*model.Organization, error) {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()

	newOrg := model.Organization{
		Base: model.Base{
			ID:        org.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: createdByIdentity,
			UpdatedBy: createdByIdentity,
		},
		Name:        org.Name,
		Slug:        org.Slug,
		IsActive:    true,
		OwnerID:     createdByIdentity,
		ScanCounter: 0,
	}

	result, err := col.InsertOne(ctx, newOrg)
	if err != nil {
		log.Println("Error creating organization:", err)
		return nil, errors.New("error creating organization")
	}
	if !result.Acknowledged {
		return nil, errors.New("error creating organization")
	}
	return &newOrg, nil
}

func (r *MongoOrganizationRepo) GetOrganizationByID(reqCtx *app.RequestContext, id bson.ObjectID) (*model.Organization, error) {
	var col = r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()
	var org model.Organization
	err := col.FindOne(ctx, bson.M{"_id": id}).Decode(&org)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("organization not found")
		}
		log.Printf("failed to get organization by id: %v", err)
		return nil, errors.New("failed to get organization")
	}
	return &org, nil
}

func (r *MongoOrganizationRepo) UpdateOrganization(reqCtx *app.RequestContext, id bson.ObjectID, org *model.UpdateOrganization) (*model.Organization, error) {
	var col = r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()

	update := &model.MongoUpdateOrganization{
		BaseUpdate: model.BaseUpdate{
			UpdatedAt: time.Now(),
			UpdatedBy: reqCtx.User.IdentityID,
		},
		OwnerID:          org.OwnerID,
		IsActive:         org.IsActive,
		Billing:          org.Billing,
		StripeCustomerId: org.StripeCustomerId,
	}

	// Remove _id from the update document before updating
	// bsonBytes, err := bson.Marshal(update)
	// if err != nil {
	// 	log.Printf("failed to marshal organization update: %v", err)
	// 	return nil, errors.New("failed to update organization")
	// }
	// var updateMap bson.M
	// if err := bson.Unmarshal(bsonBytes, &updateMap); err != nil {
	// 	log.Printf("failed to unmarshal organization update: %v", err)
	// 	return nil, errors.New("failed to update organization")
	// }
	// delete(updateMap, "_id") // Remove _id field
	// log.Printf("Update Map: %+v", updateMap)

	_, err := col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	if err != nil {
		log.Printf("failed to update organization: %v", err)
		return nil, errors.New("failed to update organization")
	}
	return r.GetOrganizationByID(reqCtx, id)
}

func (r *MongoOrganizationRepo) ListOrganizations(reqCtx *app.RequestContext, pageReq *app.PageRequest) (*app.PageResponse[*model.Organization], error) {
	var col = r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()

	// get total count and find with pagination
	opts := pageReq.ToFindOptions()

	var (
		orgs     []*model.Organization
		count    int64
		findErr  error
		countErr error
	)

	findDone := make(chan struct{})
	countDone := make(chan struct{})

	go func() {
		defer close(findDone)
		cursor, err := col.Find(ctx, bson.M{}, opts)
		if err != nil {
			findErr = err
			close(findDone)
			return
		}
		defer cursor.Close(ctx)
		if err := cursor.All(ctx, &orgs); err != nil {
			findErr = err
			return
		}
	}()

	go func() {
		defer close(countDone)
		count, countErr = col.CountDocuments(ctx, bson.M{})
	}()

	<-findDone
	<-countDone

	if findErr != nil {
		log.Printf("failed to list organizations: %v", findErr)
		return nil, errors.New("failed to list organizations")
	}
	if countErr != nil {
		log.Printf("failed to count organizations: %v", countErr)
		return nil, errors.New("failed to count organizations")
	}

	if orgs == nil {
		orgs = []*model.Organization{}
	}

	return app.NewPageResponse(count, pageReq.Page, pageReq.PageSize, orgs), nil
}

func (r *MongoOrganizationRepo) IsOrganizationExists(reqCtx *app.RequestContext, name string) (bool, error) {
	var col = r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()
	count, err := col.CountDocuments(ctx, bson.M{"name": name})
	if err != nil {
		log.Printf("failed to check organization by name: %v", err)
		return false, errors.New("failed to check organization by name")
	}
	return count > 0, nil
}

func (r *MongoOrganizationRepo) GetOrganizationCount(reqCtx *app.RequestContext) (int64, error) {
	var col = r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()
	count, err := col.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("failed to count organizations: %v", err)
		return 0, errors.New("failed to count organizations")
	}
	return count, nil
}

func (r *MongoOrganizationRepo) GenerateNextScanCode(reqCtx *app.RequestContext) (string, error) {
	col := r.GetCollection()
	filter := bson.M{"_id": reqCtx.Org.ID}
	update := bson.M{"$inc": bson.M{"scan_counter": 1}}

	// Set options for the update.
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	ctx, cancel := db.GetDBContext()
	defer cancel()
	var result bson.M
	err := col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("SCAN-%d", result["scan_counter"]), nil
}

func (r *MongoOrganizationRepo) GetOrganizations(reqCtx *app.RequestContext, filter bson.M) ([]*model.Organization, error) {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()

	cursor, err := col.Find(ctx, filter)
	if err != nil {
		log.Printf("error finding organizations: %v", err)
		return nil, errors.New("failed to get organizations")
	}
	defer cursor.Close(ctx)

	var orgs []*model.Organization
	if err := cursor.All(ctx, &orgs); err != nil {
		log.Printf("error decoding organizations: %v", err)
		return nil, errors.New("failed to decode organizations")
	}

	return orgs, nil
}

func (r *MongoOrganizationRepo) DeleteOrganizationByOwner(reqCtx *app.RequestContext) ([]*model.Organization, error) {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()
	filter := bson.M{"owner_id": reqCtx.User.IdentityID}
	orgs, err := r.GetOrganizations(reqCtx, filter)
	if err != nil {
		log.Printf("error getting organizations for deletion: %v", err)
		return nil, errors.New("failed to get organizations for deletion")
	}

	for _, org := range orgs {
		// Drop database
		if err := r.appDB.GetOrgDatabase(org.Slug).Drop(ctx); err != nil {
			log.Printf("error dropping database for organization %s: %v", org.Slug, err)
		}
	}

	_, err = col.DeleteMany(ctx, filter)
	if err != nil {
		log.Printf("error deleting organizations: %v", err)
		return nil, errors.New("failed to delete organizations")
	}

	return orgs, nil
}
