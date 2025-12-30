package service

import (
	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/db"
	"github.com/gaeaglobal/exto/server/model"
	"github.com/gaeaglobal/exto/server/repo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ScanHistoryService struct {
	dbSessionProvider db.SessionProvider
	repo              repo.ScanHistoryRepo
}

func NewScanHistoryService(dbSessionProvider db.SessionProvider, repo repo.ScanHistoryRepo) *ScanHistoryService {
	return &ScanHistoryService{
		dbSessionProvider: dbSessionProvider,
		repo:              repo,
	}
}

func (s *ScanHistoryService) ListScanHistories(reqCtx *app.RequestContext, pageReq *app.PageRequest) (*app.PageResponse[*model.ScanHistory], error) {
	return s.repo.ListScanHistories(reqCtx, pageReq)
}

func (s *ScanHistoryService) GetScanHistoryByID(reqCtx *app.RequestContext, scanHistoryID bson.ObjectID) (*model.ScanHistory, error) {
	return s.repo.GetScanHistoryByID(reqCtx, scanHistoryID)
}

func (s *ScanHistoryService) CreateScanHistory(reqCtx *app.RequestContext, scanHistory *model.CreateScanHistoryRequest) (*model.ScanHistory, error) {
	return s.repo.CreateScanHistory(reqCtx, scanHistory)
}
