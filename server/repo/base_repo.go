package repo

import (
	"github.com/gaeaglobal/exto/server/db"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type IBaseRepo interface {
	GetCollection(orgName ...string) *mongo.Collection
}

type BaseRepo struct {
	cname string
	appDB *db.AppDB
}

func (r *BaseRepo) GetCollection(names ...string) *mongo.Collection {
	if r.cname == "" {
		panic("collection name cannot be empty")
	}
	if len(names) == 1 && names[0] != "" {
		return r.appDB.GetOrgDatabase(names[0]).Collection(r.cname)
	}
	return r.appDB.GetCoreDatabase().Collection(r.cname)
}
