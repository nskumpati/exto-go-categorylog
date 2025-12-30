/*
 Copyright 2025 The Exto Project Solutions, Inc.
 All rights reserved.

 Author: Vimalraj Arumugam

 This software is the confidential and proprietary product of The Exto Project Solutions, Inc.
 and is protected by copyright and trade secret law.
 Use, reproduction, and distribution of this software is strictly forbidden.

 For more details, please refer to the LICENSE file in the root directory of this project.
*/

//go:generate mockgen -source=scan_history_repo.go -destination=../mocks/mock_scan_history_repo.go -package=mocks -copyright_file=../../copy_right.txt

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
)

type ScanHistoryRepo interface {
	IBaseRepo
	ListScanHistories(reqCtx *app.RequestContext, pageReq *app.PageRequest) (*app.PageResponse[*model.ScanHistory], error)
	GetScanHistoryByID(reqCtx *app.RequestContext, scanHistoryID bson.ObjectID) (*model.ScanHistory, error)
	CreateScanHistory(reqCtx *app.RequestContext, scanHistory *model.CreateScanHistoryRequest) (*model.ScanHistory, error)
}

type MongoScanHistoryRepo struct {
	BaseRepo
}

func NewScanHistoryRepository(appDB *db.AppDB) *MongoScanHistoryRepo {
	return &MongoScanHistoryRepo{
		BaseRepo: BaseRepo{
			cname: "scan_history",
			appDB: appDB,
		},
	}
}

func (r *MongoScanHistoryRepo) GetScanHistoryByID(reqCtx *app.RequestContext, scanHistoryID bson.ObjectID) (*model.ScanHistory, error) {
	col := r.GetCollection(reqCtx.Org.Slug)
	ctx, cancel := db.GetDBContext()
	defer cancel()

	var scanHistory model.ScanHistory
	err := col.FindOne(ctx, bson.M{"_id": scanHistoryID}).Decode(&scanHistory)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		log.Printf("failed to get scan history by ID: %v", err)
		return nil, errors.New("failed to get scan history")
	}

	return &scanHistory, nil
}

func (r *MongoScanHistoryRepo) ListScanHistories(reqCtx *app.RequestContext, pageReq *app.PageRequest) (*app.PageResponse[*model.ScanHistory], error) {
	col := r.GetCollection(reqCtx.Org.Slug)
	ctx, cancel := db.GetDBContext()
	defer cancel()

	opts := pageReq.ToFindOptions()

	var (
		scans    []*model.ScanHistory
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
			return
		}
		defer cursor.Close(ctx)

		if err := cursor.All(ctx, &scans); err != nil {
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
		log.Printf("failed to list scan histories: %v", findErr)
		return nil, errors.New("failed to list scan histories")
	}
	if countErr != nil {
		log.Printf("failed to count scan histories: %v", countErr)
		return nil, errors.New("failed to count scan histories")
	}

	if scans == nil {
		scans = []*model.ScanHistory{}
	}

	return app.NewPageResponse(count, pageReq.Page, pageReq.PageSize, scans), nil
}

func (r *MongoScanHistoryRepo) CreateScanHistory(reqCtx *app.RequestContext, scanHistory *model.CreateScanHistoryRequest) (*model.ScanHistory, error) {
	col := r.GetCollection(reqCtx.Org.Slug)
	ctx, cancel := db.GetDBContext()
	defer cancel()

	newScanHistory := &model.ScanHistory{
		Base: model.Base{
			ID:        bson.NewObjectID(),
			CreatedAt: time.Now(),
			CreatedBy: reqCtx.User.IdentityID,
			UpdatedAt: time.Now(),
			UpdatedBy: reqCtx.User.IdentityID,
		},
		CategoryID:      scanHistory.CategoryID,
		BatchID:         scanHistory.BatchID,
		CategoryDataID:  scanHistory.CategoryDataID,
		CategoryDataCol: scanHistory.CategoryDataCol,
		ScanCode:        scanHistory.ScanCode,
		Thumbnails:      scanHistory.Thumbnails,
	}

	result, err := col.InsertOne(ctx, newScanHistory)
	if err != nil {
		log.Printf("failed to create scan history: %v", err)
		return nil, errors.New("failed to create scan history")
	}

	if !result.Acknowledged {
		return nil, errors.New("failed to create scan history")
	}

	return newScanHistory, nil
}
