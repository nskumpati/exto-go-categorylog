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
	"time"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/db"
	"github.com/gaeaglobal/exto/server/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MeterEventRepository interface {
	IBaseRepo
	CreateMeterEvent(reqCtx *app.RequestContext, name string, value int, stripeCustomerID string) (*model.MeterEvent, error)
}

type MongoMeterEventRepo struct {
	BaseRepo
}

func NewMeterEventRepository(appDB *db.AppDB) *MongoMeterEventRepo {
	return &MongoMeterEventRepo{
		BaseRepo: BaseRepo{
			cname: "meter_events",
			appDB: appDB,
		},
	}
}

func (r *MongoMeterEventRepo) CreateMeterEvent(reqCtx *app.RequestContext, name string, value int, stripeCustomerID string) (*model.MeterEvent, error) {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()

	meterEvent := &model.MeterEvent{
		Base: model.Base{
			ID:        bson.NewObjectID(),
			CreatedAt: time.Now(),
			CreatedBy: reqCtx.User.IdentityID,
		},
		EventName:        name,
		EventValue:       value,
		StripeCustomerID: stripeCustomerID,
		OrganizationID:   reqCtx.Org.ID,
	}
	_, err := col.InsertOne(ctx, meterEvent)
	return meterEvent, err
}
