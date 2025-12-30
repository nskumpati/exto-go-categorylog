package repository

import (
	"context"
	"time"

	"github.com/gaeaglobal/exto/ai/db"
	"github.com/gaeaglobal/exto/ai/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DocumentRepository struct {
	collection *mongo.Collection
}

func NewDocumentRepository() *DocumentRepository {
	return &DocumentRepository{
		collection: db.Collections.Documents,
	}
}

// Create inserts a new document
func (r *DocumentRepository) Create(ctx context.Context, document *models.Document) error {
	document.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, document)
	if err != nil {
		return err
	}

	document.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByCategoryID retrieves all documents for a category
func (r *DocumentRepository) GetByCategoryID(ctx context.Context, categoryID primitive.ObjectID) ([]models.Document, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"categoryId": categoryID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var documents []models.Document
	if err = cursor.All(ctx, &documents); err != nil {
		return nil, err
	}

	if documents == nil {
		documents = []models.Document{}
	}

	return documents, nil
}

// GetByID retrieves a document by ID
func (r *DocumentRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Document, error) {
	var document models.Document
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&document)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &document, nil
}

// Delete removes a document
func (r *DocumentRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
