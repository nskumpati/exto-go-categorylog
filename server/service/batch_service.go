package service

import (
	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/db"
	"github.com/gaeaglobal/exto/server/model"
	"github.com/gaeaglobal/exto/server/repo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type BatchService struct {
	dbSessionProvider db.SessionProvider
	repo              repo.BatchRepo
}

func NewBatchService(dbSessionProvider db.SessionProvider, repo repo.BatchRepo) *BatchService {
	return &BatchService{
		dbSessionProvider: dbSessionProvider,
		repo:              repo,
	}
}

func (s *BatchService) GetBatchDetailsByID(reqCtx *app.RequestContext, batchID bson.ObjectID) (*model.Batch, error) {
	return s.repo.GetBatchByID(reqCtx, batchID)
}

func (s *BatchService) CreateBatch(reqCtx *app.RequestContext, batch *model.CreateBatchRequest) (*model.Batch, error) {
	return s.repo.CreateBatch(reqCtx, batch)
}

// update batch status
func (s *BatchService) UpdateBatchStatus(reqCtx *app.RequestContext, batchID bson.ObjectID, status string) (*model.Batch, error) {
	batch, err := s.repo.UpdateBatchStatus(reqCtx, batchID, status)
	return batch, err
}
