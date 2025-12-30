package service

import (
	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/model"
	"github.com/gaeaglobal/exto/server/repo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type SubscriptionService struct {
	r repo.SubscriptionRepository
}

func NewSubscriptionService(r repo.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		r: r,
	}
}

func (s *SubscriptionService) CreateSubscription(reqCtx *app.RequestContext, subscription *model.CreateSubscription) (*model.Subscription, error) {
	return s.r.CreateSubscription(reqCtx, subscription)
}

func (s *SubscriptionService) UpdateSubscription(reqCtx *app.RequestContext, id bson.ObjectID, subscription *model.UpdateSubscription) (*model.Subscription, error) {
	return s.r.UpdateSubscription(reqCtx, id, subscription)
}

func (s *SubscriptionService) GetMySubscription(reqCtx *app.RequestContext) (*model.Subscription, error) {
	return s.r.GetMySubscription(reqCtx)
}
