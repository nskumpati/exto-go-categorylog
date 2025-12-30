package service

import (
	"log"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/model"
	"github.com/gaeaglobal/exto/server/repo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CategoryService struct {
	repo  repo.CategoryRepository
	cache *ristretto.Cache[string, *model.Category]
}

func NewCategoryService(repo repo.CategoryRepository) *CategoryService {
	cache, err := ristretto.NewCache(&ristretto.Config[string, *model.Category]{
		NumCounters: 100, // number of keys to track frequency of (10x MaxCost).
		MaxCost:     100, // maximum cost of cache (in arbitrary units).
		BufferItems: 2,   // number of keys per Get buffer.
	})
	if err != nil {
		log.Fatalf("failed to initialize cache: %v", err)
	}
	return &CategoryService{
		repo:  repo,
		cache: cache,
	}
}

func (s *CategoryService) GetCategoryByID(reqCtx *app.RequestContext, categoryID bson.ObjectID) (*model.Category, error) {
	if category, found := s.cache.Get(categoryID.Hex()); found {
		return category, nil
	}
	category, err := s.repo.GetCategoryByID(reqCtx, categoryID)
	if err != nil {
		return nil, err
	}
	s.cache.Set(categoryID.Hex(), category, 1)
	s.cache.Wait()
	return category, nil
}

func (s *CategoryService) ListCategories(reqCtx *app.RequestContext, pageReq *app.PageRequest) (*app.PageResponse[*model.Category], error) {
	return s.repo.ListCategories(reqCtx, pageReq)
}

func (s *CategoryService) Close() {
	s.cache.Close()
}
