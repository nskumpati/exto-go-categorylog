/*
 Copyright 2025 The Exto Project Solutions, Inc.
 All rights reserved.

 Author: Vimalraj Arumugam

 This software is the confidential and proprietary product of The Exto Project Solutions, Inc.
 and is protected by copyright and trade secret law.
 Use, reproduction, and distribution of this software is strictly forbidden.

 For more details, please refer to the LICENSE file in the root directory of this project.
*/

//go:generate mockgen -source=category_data_repo.go -destination=../mocks/mock_category_data_repo.go -package=mocks -copyright_file=../../copy_right.txt
package repo

import (
	"errors"
	"time"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/db"
	"github.com/gaeaglobal/exto/server/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type CategoryDataRepository interface {
	GetCollection(orgSlug string, categorySlug string) *mongo.Collection
	GetCategoryDataByID(reqCtx *app.RequestContext, categorySlug string, id bson.ObjectID) (*model.CategoryData, error)
	ListCategoryData(reqCtx *app.RequestContext, categorySlug string, pageReq *app.PageRequest) (*app.PageResponse[*model.CategoryData], error)
	CreateCategoryData(reqCtx *app.RequestContext, categorySlug string, categoryData *model.CreateCategoryDataRequest) (*model.CategoryData, error)
	UpdateCategoryData(reqCtx *app.RequestContext, categorySlug string, id bson.ObjectID, updateMetaData *model.UpdateCategoryDataRequest) (*model.CategoryData, error)
}

type MongoCategoryDataRepo struct {
	BaseRepo
}

func NewCategoryDataRepository(appDB *db.AppDB) *MongoCategoryDataRepo {
	return &MongoCategoryDataRepo{
		BaseRepo: BaseRepo{
			appDB: appDB,
		},
	}
}

func (r *MongoCategoryDataRepo) GetCollection(orgSlug string, categorySlug string) *mongo.Collection {
	return r.appDB.GetOrgDatabase(orgSlug).Collection(categorySlug + "_data")
}

func (r *MongoCategoryDataRepo) GetCategoryDataByID(reqCtx *app.RequestContext, categorySlug string, id bson.ObjectID) (*model.CategoryData, error) {
	col := r.GetCollection(reqCtx.Org.Slug, categorySlug)
	var categoryData model.CategoryData
	ctx, cancel := db.GetDBContext()
	defer cancel()
	result := col.FindOne(ctx, bson.M{"_id": id})
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("category data not found")
		}
		return nil, err
	}
	if err := result.Decode(&categoryData); err != nil {
		return nil, err
	}
	return &categoryData, nil
}

func (r *MongoCategoryDataRepo) ListCategoryData(reqCtx *app.RequestContext, categorySlug string, pageReq *app.PageRequest) (*app.PageResponse[*model.CategoryData], error) {
	col := r.GetCollection(reqCtx.Org.Slug, categorySlug)
	ctx, cancel := db.GetDBContext()
	defer cancel()

	filter := bson.M{}

	opts := pageReq.ToFindOptions()

	var (
		categoryData []*model.CategoryData
		count        int64
		findErr      error
		countErr     error
	)

	findDone := make(chan struct{})
	countDone := make(chan struct{})

	go func() {
		defer close(findDone)
		cursor, err := col.Find(ctx, filter, opts)
		if err != nil {
			findErr = err
			return
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var catData model.CategoryData
			if err := cursor.Decode(&catData); err != nil {
				findErr = err
				return
			}
			categoryData = append(categoryData, &catData)
		}
		if err := cursor.Err(); err != nil {
			findErr = err
			return
		}
	}()

	go func() {
		defer close(countDone)
		count, countErr = col.CountDocuments(ctx, filter)
	}()

	<-findDone
	<-countDone

	if findErr != nil {
		return nil, findErr
	}
	if countErr != nil {
		return nil, countErr
	}

	if categoryData == nil {
		categoryData = []*model.CategoryData{}
	}

	return app.NewPageResponse(count, pageReq.Page, pageReq.PageSize, categoryData), nil
}

func (r *MongoCategoryDataRepo) CreateCategoryData(reqCtx *app.RequestContext, categorySlug string, categoryData *model.CreateCategoryDataRequest) (*model.CategoryData, error) {
	col := r.GetCollection(reqCtx.Org.Slug, categorySlug)
	ctx, cancel := db.GetDBContext()
	defer cancel()

	data := &model.CategoryData{
		Base: model.Base{
			ID:        bson.NewObjectID(),
			CreatedAt: time.Now(),
			CreatedBy: reqCtx.User.IdentityID,
		},
		CategoryID:     categoryData.CategoryID,
		MetaData:       categoryData.MetaData,
		RawData:        categoryData.RawData,
		DocumentPaths:  categoryData.DocumentPaths,
		OrganizationID: reqCtx.Org.ID,
	}
	_, err := col.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *MongoCategoryDataRepo) UpdateCategoryData(reqCtx *app.RequestContext, categorySlug string, id bson.ObjectID, updateMetaData *model.UpdateCategoryDataRequest) (*model.CategoryData, error) {
	col := r.GetCollection(reqCtx.Org.Slug, categorySlug)
	ctx, cancel := db.GetDBContext()
	defer cancel()

	_, err := col.UpdateOne(ctx, bson.M{"_id": id}, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "metadata", Value: updateMetaData.MetaData},
			{Key: "updated_at", Value: time.Now()},
			{Key: "updated_by", Value: reqCtx.User.IdentityID},
		}},
	})
	if err != nil {
		return nil, err
	}
	// Fetch the updated category data
	return r.GetCategoryDataByID(reqCtx, categorySlug, id)
}
