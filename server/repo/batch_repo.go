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

type BatchRepo interface {
	IBaseRepo
	GetBatchByID(reqCtx *app.RequestContext, BatchID bson.ObjectID) (*model.Batch, error)
	CreateBatch(reqCtx *app.RequestContext, batch *model.CreateBatchRequest) (*model.Batch, error)
	UpdateBatchStatus(reqCtx *app.RequestContext, batchID bson.ObjectID, status string) (*model.Batch, error)
}

type MongoBatchRepo struct {
	BaseRepo
}

// GetBatchByID implements BatchRepo.
func (r *MongoBatchRepo) GetBatchByID(reqCtx *app.RequestContext, BatchID bson.ObjectID) (*model.Batch, error) {
	panic("unimplemented")
}

// GetCollection implements BatchRepo.
// Subtle: this method shadows the method (BaseRepo).GetCollection of MongoBatchRepo.BaseRepo.
// func (r *MongoBatchRepo) GetCollection(orgName ...string) *mongo.Collection {
// 	panic("unimplemented")
// }

func NewBatchRepository(appDB *db.AppDB) *MongoBatchRepo {
	return &MongoBatchRepo{
		BaseRepo: BaseRepo{
			cname: "batch",
			appDB: appDB,
		},
	}
}

func (r *MongoBatchRepo) GetBatchID(reqCtx *app.RequestContext, batchID bson.ObjectID) (*model.Batch, error) {
	col := r.GetCollection(reqCtx.Org.Slug)
	ctx, cancel := db.GetDBContext()
	defer cancel()

	var batch model.Batch
	err := col.FindOne(ctx, bson.M{"_id": batchID}).Decode(&batch)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		log.Printf("failed to get batch details by ID: %v", err)
		return nil, errors.New("failed to get batch details")
	}

	return &batch, nil
}

func (r *MongoBatchRepo) CreateBatch(reqCtx *app.RequestContext, batch *model.CreateBatchRequest) (*model.Batch, error) {
	col := r.GetCollection(reqCtx.Org.Slug)
	ctx, cancel := db.GetDBContext()
	defer cancel()

	newBatch := &model.Batch{
		Base: model.Base{
			ID:        bson.NewObjectID(),
			CreatedAt: time.Now(),
			CreatedBy: reqCtx.User.IdentityID,
		},
		Name:   batch.Name,
		Status: "Open",
	}

	result, err := col.InsertOne(ctx, newBatch)
	if err != nil {
		log.Printf("failed to create batch: %v", err)
		return nil, errors.New("failed to create batch")
	}

	if !result.Acknowledged {
		return nil, errors.New("failed to create batch")
	}

	return newBatch, nil
}

// update batch status
func (r *MongoBatchRepo) UpdateBatchStatus(reqCtx *app.RequestContext, batchID bson.ObjectID, status string) (*model.Batch, error) {
	col := r.GetCollection(reqCtx.Org.Slug)
	ctx, cancel := db.GetDBContext()
	defer cancel()

	result, err := col.UpdateOne(ctx, bson.M{"_id": batchID}, bson.M{"$set": bson.M{"status": status, "updated_at": time.Now(), "updated_by": reqCtx.User.IdentityID}})
	if err != nil {
		log.Printf("failed to update batch status: %v", err)
		return nil, errors.New("failed to update batch status")
	}

	if result.MatchedCount == 0 {
		return nil, errors.New("batch not found")
	}

	updatedBatch, err := r.GetBatchID(reqCtx, batchID)
	if err != nil {
		return nil, err
	}

	return updatedBatch, nil
}
