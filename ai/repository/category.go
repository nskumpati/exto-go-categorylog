package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gaeaglobal/exto/ai/db"
	"github.com/gaeaglobal/exto/ai/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CategoryRepository struct {
	collection *mongo.Collection
}

func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{
		collection: db.Collections.Categories,
	}
}

// Update updates an existing category
func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) error {
	if category == nil {
		return errors.New("category cannot be nil")
	}

	if category.ID.IsZero() {
		return errors.New("category ID is required")
	}

	// Update timestamp
	category.UpdatedAt = time.Now()

	filter := bson.M{"_id": category.ID}
	update := bson.M{
		"$set": bson.M{
			"name":        category.Name,
			"schema":      category.Schema,
			"formatCount": category.FormatCount,
			"totalDocs":   category.TotalDocs,
			"updatedAt":   category.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("category not found")
	}

	return nil
}

// Delete deletes a category by ID
func (r *CategoryRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("category not found")
	}

	return nil
}

// FindByID finds a category by its ID
func (r *CategoryRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Category, error) {
	var category models.Category

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	return &category, nil
}

// IncrementFormatCount increments the format count for a category
func (r *CategoryRepository) IncrementFormatCount(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$inc": bson.M{"formatCount": 1},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("category not found")
	}

	return nil
}

// IncrementTotalDocs increments the total documents count for a category
func (r *CategoryRepository) IncrementTotalDocs(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$inc": bson.M{"totalDocs": 1},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("category not found")
	}

	return nil
}

// // CreateIndexes creates necessary indexes for the categories collection
// func (r *CategoryRepository) CreateIndexes(ctx context.Context) error {
// 	indexes := []mongo.IndexModel{
// 		// Unique index on category name
// 		{
// 			Keys:    bson.D{{Key: "name", Value: 1}},
// 			Options: options.Index().SetUnique(true),
// 		},
// 		// Index on createdAt for time-based queries
// 		{
// 			Keys: bson.D{{Key: "createdAt", Value: -1}},
// 		},
// 	}

// 	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// GetCategoryStats returns statistics about a category
func (r *CategoryRepository) GetCategoryStats(ctx context.Context) (map[string]interface{}, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{
			"_id":              nil,
			"totalCategories":  bson.M{"$sum": 1},
			"totalFormats":     bson.M{"$sum": "$formatCount"},
			"totalDocuments":   bson.M{"$sum": "$totalDocs"},
			"avgFormatsPerCat": bson.M{"$avg": "$formatCount"},
			"avgDocsPerCat":    bson.M{"$avg": "$totalDocs"},
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
			"totalCategories":  0,
			"totalFormats":     0,
			"totalDocuments":   0,
			"avgFormatsPerCat": 0,
			"avgDocsPerCat":    0,
		}, nil
	}

	return results[0], nil
}

// GetAll retrieves all categories
func (r *CategoryRepository) GetAll(ctx context.Context) ([]models.Category, error) {
	cursor, err := r.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		fmt.Print("ERROR")
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []models.Category

	if err = cursor.All(ctx, &categories); err != nil {
		return nil, err
	}

	if categories == nil {
		categories = []models.Category{}
	}
	return categories, nil
}

// Create inserts a new category and returns its ID
func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) (primitive.ObjectID, error) {
	fmt.Println("created one")
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("Testing direct insert with bson.M...")
	startTime := time.Now()

	result, err := r.collection.InsertOne(ctx, category)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("❌ Direct insert FAILED after %v: %v", duration, err)
		return primitive.NilObjectID, err
	}

	log.Printf("✓ Direct insert SUCCESS in %v, ID: %v", duration, result.InsertedID)

	id := result.InsertedID.(primitive.ObjectID)
	fmt.Println("id: ")
	fmt.Println(id)

	return id, nil
}

// FindByName finds a category by name (case-insensitive)
func (r *CategoryRepository) FindByName(ctx context.Context, name string) (*models.Category, error) {
	var category models.Category
	filter := bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: "^" + name + "$", Options: "i"}}}

	err := r.collection.FindOne(ctx, filter).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &category, nil
}

// // IncrementFormatCount increments the format count for a category
// func (r *CategoryRepository) IncrementFormatCount(ctx context.Context, categoryID primitive.ObjectID) error {
// 	update := bson.M{
// 		"$inc": bson.M{"formatCount": 1},
// 		"$set": bson.M{"updatedAt": time.Now()},
// 	}

// 	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": categoryID}, update)
// 	return err
// }

// GetByID retrieves a category by ID
func (r *CategoryRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Category, error) {
	var category models.Category
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}
