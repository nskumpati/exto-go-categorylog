/*
 Copyright 2025 The Exto Project Solutions, Inc.
 All rights reserved.

 Author: Vimalraj Arumugam

 This software is the confidential and proprietary product of The Exto Project Solutions, Inc.
 and is protected by copyright and trade secret law.
 Use, reproduction, and distribution of this software is strictly forbidden.

 For more details, please refer to the LICENSE file in the root directory of this project.
*/

//go:generate mockgen -source=format_repo.go -destination=../mocks/mock_format_repo.go -package=mocks -copyright_file=../../copy_right.txt
package repo

import (
	"errors"
	"log"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/db"
	"github.com/gaeaglobal/exto/server/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type FormatRepository interface {
	IBaseRepo
	GetFormatsByCategoryID(reqCtx *app.RequestContext, categoryID bson.ObjectID) ([]model.Format, error)
}

type MongoFormatRepo struct {
	BaseRepo
}

func NewFormatRepository(appDb *db.AppDB) *MongoFormatRepo {
	return &MongoFormatRepo{
		BaseRepo: BaseRepo{
			cname: "formats",
			appDB: appDb,
		},
	}
}

func (r *MongoFormatRepo) GetFormatsByCategoryID(reqCtx *app.RequestContext, categoryID bson.ObjectID) ([]model.Format, error) {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()

	var formats []model.Format
	cursor, err := col.Find(ctx, bson.M{"category_id": categoryID})
	if err != nil {
		log.Println("Error finding formats:", err)
		return nil, errors.New("error fetching formats")
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &formats); err != nil {
		return nil, errors.New("error fetching formats")
	}

	return formats, nil
}
