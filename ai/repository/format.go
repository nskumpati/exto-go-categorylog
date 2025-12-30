package repository

import (
	"context"
	"errors"
	"time"

	"github.com/gaeaglobal/exto/ai/db"
	"github.com/gaeaglobal/exto/ai/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FormatRepository struct {
	collection *mongo.Collection
}

// NewFormatRepository creates a new format repository
func NewFormatRepository() *FormatRepository {
	return &FormatRepository{
		collection: db.Collections.Formats,
	}
}

// Create inserts a new format
func (r *FormatRepository) Create(ctx context.Context, format *models.Format) error {
	if format == nil {
		return errors.New("format cannot be nil")
	}

	// Set timestamps
	now := time.Now()
	format.CreatedAt = now

	// Generate new ObjectID if not set
	if format.ID.IsZero() {
		format.ID = primitive.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, format)
	if err != nil {
		return err
	}

	return nil
}

// Update updates an existing format
func (r *FormatRepository) Update(ctx context.Context, format *models.Format) error {
	if format == nil {
		return errors.New("format cannot be nil")
	}

	if format.ID.IsZero() {
		return errors.New("format ID is required")
	}

	filter := bson.M{"_id": format.ID}
	update := bson.M{
		"$set": bson.M{
			"categoryId":      format.CategoryID,
			"categoryName":    format.CategoryName,
			"formatNumber":    format.FormatNumber,
			"ExtractedFields": format.ExtractedFields,
			"sampleCount":     format.SampleCount,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("format not found")
	}

	return nil
}

// FindByID finds a format by its ID
func (r *FormatRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Format, error) {
	var format models.Format

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&format)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("format not found")
		}
		return nil, err
	}

	return &format, nil
}

// FindByCategoryID finds all formats for a specific category
func (r *FormatRepository) FindByCategoryID(ctx context.Context, categoryID primitive.ObjectID) ([]*models.Format, error) {
	filter := bson.M{"categoryId": categoryID}
	opts := options.Find().SetSort(bson.D{{Key: "formatNumber", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var formats []*models.Format
	if err := cursor.All(ctx, &formats); err != nil {
		return nil, err
	}

	// Return empty slice instead of nil if no formats found
	if formats == nil {
		formats = []*models.Format{}
	}

	return formats, nil
}

// FindByCategoryIDAndNumber finds a specific format by category ID and format number
func (r *FormatRepository) FindByCategoryIDAndNumber(ctx context.Context, categoryID primitive.ObjectID, formatNumber int) (*models.Format, error) {
	var format models.Format

	filter := bson.M{
		"categoryId":   categoryID,
		"formatNumber": formatNumber,
	}

	err := r.collection.FindOne(ctx, filter).Decode(&format)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("format not found")
		}
		return nil, err
	}

	return &format, nil
}

// GetAll retrieves all formats
func (r *FormatRepository) GetAll(ctx context.Context) ([]*models.Format, error) {
	opts := options.Find().SetSort(bson.D{
		{Key: "categoryName", Value: 1},
		{Key: "formatNumber", Value: 1},
	})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var formats []*models.Format
	if err := cursor.All(ctx, &formats); err != nil {
		return nil, err
	}

	return formats, nil
}

// Delete deletes a format by ID
func (r *FormatRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("format not found")
	}

	return nil
}

// DeleteByCategoryID deletes all formats for a category
func (r *FormatRepository) DeleteByCategoryID(ctx context.Context, categoryID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"categoryId": categoryID})
	return err
}

// IncrementSampleCount increments the sample count for a format
func (r *FormatRepository) IncrementSampleCount(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$inc": bson.M{"sampleCount": 1}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("format not found")
	}

	return nil
}

// CountByCategory counts the number of formats for a category
func (r *FormatRepository) CountByCategory(ctx context.Context, categoryID primitive.ObjectID) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"categoryId": categoryID})
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CreateIndexes creates necessary indexes for the formats collection
func (r *FormatRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		// Compound unique index on categoryId + formatNumber
		{
			Keys: bson.D{
				{Key: "categoryId", Value: 1},
				{Key: "formatNumber", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		// Index on categoryId for faster lookups
		{
			Keys: bson.D{{Key: "categoryId", Value: 1}},
		},
		// Index on categoryName for easier querying
		{
			Keys: bson.D{{Key: "categoryName", Value: 1}},
		},
		// Index on createdAt for time-based queries
		{
			Keys: bson.D{{Key: "createdAt", Value: -1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}

	return nil
}

// GetFormatStats returns statistics about formats
func (r *FormatRepository) GetFormatStats(ctx context.Context, categoryID primitive.ObjectID) (map[string]interface{}, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"categoryId": categoryID}}},
		{{Key: "$group", Value: bson.M{
			"_id":                 "$categoryId",
			"totalFormats":        bson.M{"$sum": 1},
			"totalSamples":        bson.M{"$sum": "$sampleCount"},
			"avgSamplesPerFormat": bson.M{"$avg": "$sampleCount"},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return map[string]interface{}{
			"totalFormats":        0,
			"totalSamples":        0,
			"avgSamplesPerFormat": 0,
		}, nil
	}

	return results[0], nil
}
