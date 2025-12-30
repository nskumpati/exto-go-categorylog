package service

import (
	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/model"
	"github.com/gaeaglobal/exto/server/repo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type FormatService struct {
	r repo.FormatRepository
}

func NewFormatService(r repo.FormatRepository) *FormatService {
	return &FormatService{r: r}
}

func (s *FormatService) GetFormatsByCategoryID(reqCtx *app.RequestContext, categoryID bson.ObjectID) ([]model.Format, error) {
	return s.r.GetFormatsByCategoryID(reqCtx, categoryID)
}
