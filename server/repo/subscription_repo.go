package repo

import (
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/db"
	"github.com/gaeaglobal/exto/server/model"
)

type SubscriptionRepository interface {
	CreateSubscription(reqCtx *app.RequestContext, subscription *model.CreateSubscription) (*model.Subscription, error)
	GetSubscriptionByID(reqCtx *app.RequestContext, subscriptionID bson.ObjectID) (*model.Subscription, error)
	UpdateSubscription(reqCtx *app.RequestContext, subscriptionID bson.ObjectID, update *model.UpdateSubscription) (*model.Subscription, error)
	GetMySubscription(reqCtx *app.RequestContext) (*model.Subscription, error)
	GetFreeTrialInfo(reqCtx *app.RequestContext) (*model.GetFreeTrialInfo, error)
}

type MongoSubscriptionRepo struct {
	BaseRepo
}

func NewSubscriptionRepository(appDB *db.AppDB) *MongoSubscriptionRepo {
	return &MongoSubscriptionRepo{
		BaseRepo: BaseRepo{appDB: appDB,
			cname: "subscriptions",
		},
	}
}

func (r *MongoSubscriptionRepo) CreateSubscription(reqCtx *app.RequestContext, subscription *model.CreateSubscription) (*model.Subscription, error) {
	sub := &model.Subscription{
		Base: model.Base{
			ID:        bson.NewObjectID(),
			CreatedAt: time.Now(),
			CreatedBy: reqCtx.User.IdentityID,
		},
		OrganizationID:  reqCtx.Org.ID,
		StripeSubID:     subscription.StripeSubID,
		StartDate:       subscription.StartDate,
		EndDate:         subscription.EndDate,
		TrialPeriodDays: subscription.TrialPeriodDays,
		BillingCycle:    subscription.BillingCycle,
		Status:          model.SubscriptionStatusActive,
		IsCurrent:       true,
	}
	col := r.GetCollection()

	ctx, cancel := db.GetDBContext()
	defer cancel()

	insertResult, err := col.InsertOne(ctx, sub)
	if err != nil {
		log.Printf("error inserting subscription: %v", err)
		return nil, errors.New("failed to create subscription")
	}

	if !insertResult.Acknowledged {
		log.Printf("error inserting subscription: not acknowledged")
		return nil, errors.New("failed to create subscription")
	}

	return sub, nil
}

func (r *MongoSubscriptionRepo) GetSubscriptionByID(reqCtx *app.RequestContext, subscriptionID bson.ObjectID) (*model.Subscription, error) {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()

	var subscription model.Subscription
	findResult := col.FindOne(ctx, bson.M{"_id": subscriptionID})
	if err := findResult.Decode(&subscription); err != nil {
		log.Printf("error finding subscription: %v", err)
		return nil, errors.New("failed to get subscription")
	}
	return &subscription, nil
}

func (r *MongoSubscriptionRepo) UpdateSubscription(reqCtx *app.RequestContext, subscriptionID bson.ObjectID, update *model.UpdateSubscription) (*model.Subscription, error) {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()

	sub := &model.Subscription{
		Base: model.Base{
			UpdatedAt: time.Now(),
			UpdatedBy: reqCtx.User.IdentityID,
		},
		EndDate: update.EndDate,
		Status:  update.Status,
	}

	_, err := col.UpdateOne(ctx, bson.M{"_id": subscriptionID}, bson.M{"$set": sub})
	if err != nil {
		log.Printf("error updating subscription: %v", err)
		return nil, errors.New("failed to update subscription")
	}

	return sub, nil
}

func (r *MongoSubscriptionRepo) GetMySubscription(reqCtx *app.RequestContext) (*model.Subscription, error) {
	col := r.GetCollection()
	ctx, cancel := db.GetDBContext()
	defer cancel()

	var subscription model.Subscription
	findResult := col.FindOne(ctx, bson.M{"organization_id": reqCtx.User.OrganizationID, "is_current": true})
	if err := findResult.Decode(&subscription); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("no current subscription found")
		}
		log.Printf("error finding subscription: %v", err)
		return nil, errors.New("failed to get subscription")
	}

	return &subscription, nil
}

func (r *MongoSubscriptionRepo) GetFreeTrialInfo(reqCtx *app.RequestContext) (*model.GetFreeTrialInfo, error) {
	ctx, cancel := db.GetDBContext()
	defer cancel()

	dbOrg := r.appDB.GetOrgDatabase(reqCtx.Org.Slug)
	colScanHistory := dbOrg.Collection("scan_history")
	count, err := colScanHistory.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("error counting scan_history documents: %v", err)
		return nil, errors.New("failed to get scan count")
	}

	return &model.GetFreeTrialInfo{
		RecordCount: int(count),
	}, nil
}
