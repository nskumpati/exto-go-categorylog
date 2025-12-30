/*
 Copyright 2025 The Exto Project Solutions, Inc.
 All rights reserved.

 Author: Vimalraj Arumugam

 This software is the confidential and proprietary product of The Exto Project Solutions, Inc.
 and is protected by copyright and trade secret law.
 Use, reproduction, and distribution of this software is strictly forbidden.

 For more details, please refer to the LICENSE file in the root directory of this project.
*/

//go:generate mockgen -source=category_repo.go -destination=../mocks/mock_category_repo.go -package=mocks -copyright_file=../../copy_right.txt
package repo

import (
	"errors"
	"log"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/db"
	"github.com/gaeaglobal/exto/server/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type CategoryRepository interface {
	IBaseRepo
	ListCategories(reqCtx *app.RequestContext, pageReq *app.PageRequest) (*app.PageResponse[*model.Category], error)
	GetCategoryByID(reqCtx *app.RequestContext, categoryID bson.ObjectID) (*model.Category, error)
}

type MongoCategoryRepo struct {
	BaseRepo
}

func NewCategoryRepository(appDB *db.AppDB) *MongoCategoryRepo {
	return &MongoCategoryRepo{
		BaseRepo: BaseRepo{
			cname: "categories",
			appDB: appDB,
		},
	}
}

func (r *MongoCategoryRepo) GetCategoryByID(reqCtx *app.RequestContext, categoryID bson.ObjectID) (*model.Category, error) {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()

	var category model.Category
	err := col.FindOne(ctx, bson.M{"_id": categoryID}).Decode(&category)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		log.Printf("failed to get category by ID: %v", err)
		return nil, errors.New("failed to get category")
	}
	return &category, nil
}

func (r *MongoCategoryRepo) ListCategories(reqCtx *app.RequestContext, pageReq *app.PageRequest) (*app.PageResponse[*model.Category], error) {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()

	opts := pageReq.ToFindOptions()

	var (
		categories []*model.Category
		count      int64
		findErr    error
		countErr   error
	)

	findDone := make(chan struct{})
	countDone := make(chan struct{})

	go func() {
		defer close(findDone)
		cursor, err := col.Find(ctx, bson.M{}, opts)
		if err != nil {
			findErr = err
			return
		}
		defer cursor.Close(ctx)

		if err := cursor.All(ctx, &categories); err != nil {
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
		log.Printf("failed to list categories: %v", findErr)
		return nil, errors.New("failed to list categories")
	}
	if countErr != nil {
		log.Printf("failed to count categories: %v", countErr)
		return nil, errors.New("failed to count categories")
	}

	if categories == nil {
		categories = []*model.Category{}
	}

	return app.NewPageResponse(count, pageReq.Page, pageReq.PageSize, categories), nil
}
