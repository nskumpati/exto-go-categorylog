package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gaeaglobal/exto/ai/db"
	"github.com/gaeaglobal/exto/ai/models"
	"github.com/gaeaglobal/exto/ai/repository"
	"github.com/gaeaglobal/exto/ai/service"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/extemporalgenome/npdfpages"
)

var (
	categoryRepo *repository.CategoryRepository
	documentRepo *repository.DocumentRepository
	formatRepo   *repository.FormatRepository
)

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func getCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	categories, err := categoryRepo.GetAll(ctx)
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func uploadDocumentHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (10MB max)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "File too large or invalid form data", http.StatusBadRequest)
		return
	}

	// Get the file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
		return
	}

	uploadDir := "./uploads"
	// Remove the directory if it already exists (and any contents).
	if err := os.RemoveAll(uploadDir); err != nil {
		http.Error(w, "Failed to remove existing upload directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the new, empty directory.
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "Failed to create upload directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	fileName := fmt.Sprintf("%d_%s", timestamp, header.Filename)
	filePath := filepath.Join(uploadDir, fileName)
	fmt.Printf("filepath: %s\n", filePath)

	// Save file to disk
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	fileSize, err := io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	pages := npdfpages.PagesAtPath(filePath)
	fmt.Printf("The PDF file '%s' has %d pages.\n", filePath, pages)

	// Extract and analyse PDF content
	category, err := service.Docservice(filePath)
	fmt.Printf("CATEGORY: %s\n", category)
	fmt.Printf("MetaData: %s\n", category.Metadata)

	if err != nil {
		log.Printf("Warning: Failed to extract PDF text: %v", err)
	}

	// Determine category name from PDF content
	categoryName := category.Category
	if categoryName == "" {
		categoryName = "General Document"
	}
	fmt.Printf("categoryName: %s\n", categoryName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if category exists (don't create it here)
	existingCategory, err := categoryRepo.FindByName(ctx, categoryName)
	if err != nil {
		log.Printf("Error finding category: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Extract metadata
	extractedData, err := service.DocMetaDataService(filePath)
	if err != nil {
		log.Printf("Error extracting metadata: %v", err)
		http.Error(w, "Failed to extract metadata", http.StatusInternalServerError)
		return
	}

	var extractedFields models.ExtractedFieldsWrapper

	// Marshal map to JSON bytes
	jsonBytes, err := json.Marshal(extractedData)
	if err != nil {
		log.Printf("Error marshaling map: %v", err)
	}

	// Unmarshal JSON bytes to struct
	err = json.Unmarshal(jsonBytes, &extractedFields)
	if err != nil {
		log.Printf("Error unmarshaling to struct: %v", err)
	}
	fmt.Println("Extracted fields:", extractedFields)

	// Prepare response without creating category or document yet
	response := models.DocumentUploadResponse{
		Success:         true,
		Message:         "Document processed successfully - awaiting schema approval",
		CategoryName:    categoryName,
		CategoryID:      primitive.NilObjectID,
		FileName:        fileName,
		FileSize:        fileSize,
		PageCount:       pages,
		ExtractedFields: extractedFields,
		Formats:         nil,
		Confidence:      "high",
		IsNewCategory:   existingCategory == nil,
	}

	if existingCategory != nil {
		response.CategoryID = existingCategory.ID
	}

	fmt.Printf("Document response: %+v\n", response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

	log.Printf("File processed: %s, Category: %s, IsNew: %v", fileName, categoryName, existingCategory == nil)
}

func ExtractFieldsFromWrapper(wrapper models.ExtractedFieldsWrapper) []models.Field {
	if len(wrapper.KeyValues) == 0 {
		return []models.Field{}
	}

	fields := make([]models.Field, 0, len(wrapper.KeyValues))

	for i, kv := range wrapper.KeyValues {
		field := models.Field{
			Name:     sanitizeFieldName(kv.Key),
			Label:    formatLabel(kv.Key),
			Type:     inferFieldType(kv.Value),
			Required: false,
			Unique:   false,
			Order:    i,
		}

		fields = append(fields, field)
	}

	return fields
}

// Helper function to extract fields from extractedFields map
func ExtractFieldsFromExtractedFields(extractedFieldsMap map[string]interface{}) []models.Field {
	if len(extractedFieldsMap) == 0 {
		return []models.Field{}
	}

	fields := make([]models.Field, 0, len(extractedFieldsMap))

	i := 0
	for key, value := range extractedFieldsMap {
		field := models.Field{
			Name:     sanitizeFieldName(key),
			Label:    formatLabel(key),
			Type:     inferFieldType(value),
			Required: false,
			Unique:   false,
			Order:    i,
		}
		fields = append(fields, field)
		i++
	}

	return fields
}

// sanitizeFieldName converts a key to a valid field name
func sanitizeFieldName(key string) string {
	// Remove special characters and convert to snake_case
	name := strings.ToLower(key)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	// Remove any non-alphanumeric characters except underscore
	var result strings.Builder
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// formatLabel converts a key to a human-readable label
func formatLabel(key string) string {
	// Convert snake_case or kebab-case to Title Case
	words := strings.FieldsFunc(key, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

// Helper functions for type inference
func isNumeric(s string) bool {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")

	if s == "" {
		return false
	}

	hasDecimal := false
	for i, char := range s {
		if char == '.' {
			if hasDecimal {
				return false
			}
			hasDecimal = true
			continue
		}
		if char == '-' && i == 0 {
			continue
		}
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func isDate(s string) bool {
	datePatterns := []string{
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"2006/01/02",
		"Jan 02, 2006",
		"02-Jan-2006",
	}

	for _, pattern := range datePatterns {
		if _, err := time.Parse(pattern, s); err == nil {
			return true
		}
	}
	return false
}

func isDateTime(s string) bool {
	datetimePatterns := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"01/02/2006 15:04:05",
		"02-Jan-2006 15:04:05",
	}

	for _, pattern := range datetimePatterns {
		if _, err := time.Parse(pattern, s); err == nil {
			return true
		}
	}
	return false
}

func isPhoneNumber(s string) bool {
	// Remove common phone number characters
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)

	// Check if it has 7-15 digits (typical phone number range)
	length := len(cleaned)
	return length >= 7 && length <= 15 &&
		(strings.Contains(s, "-") || strings.Contains(s, "(") ||
			strings.Contains(s, ")") || strings.Contains(s, " "))
}

// getFormatsByCategoryHandler handles GET /api/categories/{categoryId}/formats
func getFormatsByCategoryHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract category ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	categoryID := strings.TrimSuffix(path, "/formats")

	objectID, err := primitive.ObjectIDFromHex(categoryID)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	formats, err := formatRepo.FindByCategoryID(ctx, objectID)
	if err != nil {
		log.Printf("Error fetching formats: %v", err)
		http.Error(w, "Failed to fetch formats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(formats)
}

// getCategoryWithFormatsHandler handles GET /api/categories/{id}
func getCategoryWithFormatsHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	identifier := strings.TrimSuffix(path, "/formats")

	if identifier == "" {
		http.Error(w, "Category ID is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find category by ID or name
	var category *models.Category
	var err error

	if objectID, parseErr := primitive.ObjectIDFromHex(identifier); parseErr == nil {
		category, err = categoryRepo.FindByID(ctx, objectID)
	} else {
		category, err = categoryRepo.FindByName(ctx, identifier)
	}

	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	// Get all formats for this category
	formats, err := formatRepo.FindByCategoryID(ctx, category.ID)
	if err != nil {
		log.Printf("Error fetching formats: %v", err)
		http.Error(w, "Failed to fetch formats", http.StatusInternalServerError)
		return
	}

	response := struct {
		*models.Category
		Formats []*models.Format `json:"formats"`
	}{
		Category: category,
		Formats:  formats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// saveExtractedFieldsHandler handles POST /api/documents/{categoryIdOrName}/fields
// Compares extracted fields against existing formats and creates new format if needed
// NOW ALSO HANDLES CATEGORY CREATION
func saveExtractedFieldsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if repositories are initialized
	if categoryRepo == nil {
		log.Printf("ERROR: categoryRepo is nil")
		http.Error(w, "Internal server error: category repository not initialized", http.StatusInternalServerError)
		return
	}

	if formatRepo == nil {
		log.Printf("ERROR: formatRepo is nil")
		http.Error(w, "Internal server error: format repository not initialized", http.StatusInternalServerError)
		return
	}

	// Extract category ID or name from URL: /api/documents/{categoryIdOrName}/fields
	path := strings.TrimPrefix(r.URL.Path, "/api/documents/")
	categoryIdentifier := strings.TrimSuffix(path, "/fields")

	log.Printf("Extracted categoryIdentifier: %s", categoryIdentifier)

	if categoryIdentifier == "" || categoryIdentifier == path {
		log.Printf("Invalid categoryIdentifier extraction. Path: %s, Identifier: %s", path, categoryIdentifier)
		http.Error(w, "Category ID or name is required", http.StatusBadRequest)
		return
	}

	var req struct {
		ExtractedFields interface{} `json:"extractedFields"`
		CategorySummary string      `json:"categorySummary,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.ExtractedFields == nil {
		http.Error(w, "extractedFields is required", http.StatusBadRequest)
		return
	}

	// Convert extractedFields to map[string]interface{}
	extractedFieldsMap, err := convertExtractedFields(req.ExtractedFields)
	if err != nil {
		http.Error(w, "Invalid extractedFields format: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Extracted fields map: %v", extractedFieldsMap)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var category *models.Category
	var objectID primitive.ObjectID
	isNewCategory := false

	// Try to parse as ObjectID first
	if oid, err := primitive.ObjectIDFromHex(categoryIdentifier); err == nil {
		// It's an ObjectID
		category, err = categoryRepo.FindByID(ctx, oid)
		if err != nil {
			log.Printf("Category not found by ID: %v", err)
			http.Error(w, "Category not found: "+err.Error(), http.StatusNotFound)
			return
		}
		objectID = oid
	} else {
		// It's a category name - check if exists
		category, err = categoryRepo.FindByName(ctx, categoryIdentifier)
		if err != nil {
			log.Printf("Error finding category by name: %v", err)
		}

		if category == nil {
			// Category doesn't exist - create it
			log.Printf("Creating new category: %s", categoryIdentifier)

			fields := ExtractFieldsFromExtractedFields(extractedFieldsMap)

			newCategory := &models.Category{
				Name:        categoryIdentifier,
				FormatCount: 0,
				Summary:     req.CategorySummary,
				TotalDocs:   0,
				Schema:      fields,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			categoryID, err := categoryRepo.Create(ctx, newCategory)
			if err != nil {
				log.Printf("Error creating category: %v", err)
				http.Error(w, "Failed to create category", http.StatusInternalServerError)
				return
			}

			log.Printf("Created category with ID: %s", categoryID)
			objectID = categoryID
			isNewCategory = true

			// Fetch the newly created category
			category, err = categoryRepo.FindByID(ctx, categoryID)
			if err != nil {
				log.Printf("Error fetching newly created category: %v", err)
				http.Error(w, "Failed to fetch category", http.StatusInternalServerError)
				return
			}
		} else {
			objectID = category.ID
		}
	}

	if category == nil {
		log.Printf("Category is nil for identifier: %s", categoryIdentifier)
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	log.Printf("Found category: %s (ID: %s)", category.Name, category.ID.Hex())

	// Initialize category.Schema if nil
	if category.Schema == nil {
		category.Schema = []models.Field{}
	}

	// Build a map of existing field names in category.Schema
	existingFieldNames := make(map[string]bool)
	for _, field := range category.Schema {
		existingFieldNames[field.Name] = true
	}

	// Extract keys from extracted fields
	newFieldKeys := make([]string, 0, len(extractedFieldsMap))
	for key := range extractedFieldsMap {
		newFieldKeys = append(newFieldKeys, key)
	}
	log.Printf("Extracted field keys (%d): %v", len(newFieldKeys), newFieldKeys)

	// Check for new fields not in category.Schema
	newFieldNames := make([]string, 0)
	hasNewFields := false

	for key := range extractedFieldsMap {
		if !existingFieldNames[key] {
			newFieldNames = append(newFieldNames, key)
			hasNewFields = true
		}
	}

	if !hasNewFields && !isNewCategory {
		log.Printf("No new fields detected. All fields exist in category schema.")

		existingKeys := make([]string, 0, len(category.Schema))
		for _, field := range category.Schema {
			existingKeys = append(existingKeys, field.Name)
		}

		response := map[string]interface{}{
			"message":       "No new fields detected - all fields already exist in category schema",
			"categoryId":    category.ID.Hex(),
			"categoryName":  category.Name,
			"isNewFormat":   false,
			"isNewCategory": false,
			"hasNewFields":  false,
			"existingKeys":  existingKeys,
			"providedKeys":  newFieldKeys,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// New fields detected or new category - update category.Schema
	if hasNewFields {
		log.Printf("New fields detected (%d): %v", len(newFieldNames), newFieldNames)

		// Add new fields to category.Schema
		nextOrder := len(category.Schema)
		for _, fieldName := range newFieldNames {
			newField := models.Field{
				Name:     fieldName,
				Label:    formatLabel(fieldName),
				Type:     inferFieldType(extractedFieldsMap[fieldName]),
				Required: false,
				Unique:   false,
				Order:    nextOrder,
			}
			category.Schema = append(category.Schema, newField)
			nextOrder++
		}
		log.Printf("Updated schema with %d new fields", len(newFieldNames))
	}

	// Get all existing formats to determine next format number
	existingFormats, err := formatRepo.FindByCategoryID(ctx, objectID)
	if err != nil {
		log.Printf("Error fetching formats: %v", err)
		http.Error(w, "Failed to fetch formats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if existingFormats == nil {
		existingFormats = []*models.Format{}
	}

	newFormatNumber := len(existingFormats) + 1

	// Create new format with ALL extracted fields
	newFormat := &models.Format{
		CategoryID:      objectID,
		CategoryName:    category.Name,
		FormatNumber:    newFormatNumber,
		ExtractedFields: extractedFieldsMap,
		SampleCount:     1,
		CreatedAt:       time.Now(),
		// UpdatedAt:       time.Now(),
	}

	if err := formatRepo.Create(ctx, newFormat); err != nil {
		log.Printf("Error creating format: %v", err)
		http.Error(w, "Failed to create format: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Created Format %d with all extracted fields", newFormatNumber)

	// Update category: increment format count and total docs
	category.FormatCount = newFormatNumber
	category.TotalDocs++
	category.UpdatedAt = time.Now()

	if err := categoryRepo.Update(ctx, category); err != nil {
		log.Printf("Error updating category: %v", err)
		http.Error(w, "Failed to update category: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var message string
	if isNewCategory {
		message = fmt.Sprintf("New category created with Format 1")
	} else if hasNewFields {
		message = fmt.Sprintf("New fields detected - Format %d created and category schema updated", newFormatNumber)
	} else {
		message = fmt.Sprintf("Format %d created for existing category", newFormatNumber)
	}

	response := map[string]interface{}{
		"message":         message,
		"categoryId":      category.ID.Hex(),
		"categoryName":    category.Name,
		"formatId":        newFormat.ID.Hex(),
		"formatNumber":    newFormatNumber,
		"isNewFormat":     true,
		"isNewCategory":   isNewCategory,
		"hasNewFields":    hasNewFields,
		"newFields":       newFieldNames,
		"allFields":       newFieldKeys,
		"newFieldCount":   len(newFieldNames),
		"totalFieldCount": len(newFieldKeys),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Helper function to infer field type from value
func inferFieldType(value interface{}) models.FieldType {
	if value == nil {
		return models.FieldTypeText
	}

	switch v := value.(type) {
	case bool:
		return models.FieldTypeBoolean
	case float64, int, int64:
		return models.FieldTypeNumber
	case string:
		// Could add more sophisticated type inference here
		// (e.g., check for email, URL, date patterns)
		return models.FieldTypeText
	default:
		fmt.Print(v)
		return models.FieldTypeText
	}
}

// findFormatByKeys finds a format that has the exact same keys
func findFormatByKeys(formats []*models.Format, newKeys []string) *models.Format {
	for _, format := range formats {
		// Extract keys from format schema
		formatKeys := make([]string, 0, len(format.ExtractedFields))
		for key := range format.ExtractedFields {
			formatKeys = append(formatKeys, key)
		}

		// Compare key sets
		if keysMatch(formatKeys, newKeys) {
			return format
		}
	}
	return nil
}

// keysMatch checks if two key sets are identical (same count and same keys)
func keysMatch(keys1, keys2 []string) bool {
	// First check: must have same number of keys
	if len(keys1) != len(keys2) {
		return false
	}

	// Second check: all keys must match
	keyMap := make(map[string]bool)
	for _, key := range keys1 {
		keyMap[key] = true
	}

	for _, key := range keys2 {
		if !keyMap[key] {
			return false
		}
	}

	return true
}

// convertExtractedFields converts extractedFields from array or map format to map[string]interface{}
func convertExtractedFields(extractedFields interface{}) (map[string]interface{}, error) {
	// Check if it's already a map
	if fieldMap, ok := extractedFields.(map[string]interface{}); ok {
		return fieldMap, nil
	}

	// Check if it's an array
	if fieldArray, ok := extractedFields.([]interface{}); ok {
		resultMap := make(map[string]interface{})

		for _, item := range fieldArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				// Extract key and value from the array element
				key, keyExists := itemMap["key"].(string)
				value, valueExists := itemMap["value"]

				if keyExists && valueExists {
					// Store the entire object (key, value, description) or just value
					resultMap[key] = map[string]interface{}{
						"value":       value,
						"description": itemMap["description"],
					}
				}
			}
		}

		if len(resultMap) == 0 {
			return nil, fmt.Errorf("no valid key-value pairs found in array")
		}

		return resultMap, nil
	}

	return nil, fmt.Errorf("extractedFields must be either a map or an array")
}

func main() {
	// Initialize MongoDB
	config := db.Config{
		URI:          "mongodb://localhost:27017/",
		DatabaseName: "document_manager",
		Timeout:      10 * time.Second,
	}

	if err := db.Connect(config); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer db.Disconnect()

	// Initialize repositories
	categoryRepo = repository.NewCategoryRepository()
	documentRepo = repository.NewDocumentRepository()
	formatRepo = repository.NewFormatRepository()

	// Routes
	http.HandleFunc("/api/categories", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getCategoriesHandler(w, r)
		case "OPTIONS":
			enableCORS(w)
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Categories with ID route - matches /api/categories/{id} or /api/categories/{id}/formats
	http.HandleFunc("/api/categories/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/formats") {
			getFormatsByCategoryHandler(w, r)
		} else {
			getCategoryWithFormatsHandler(w, r)
		}
	})

	// Documents route - matches /api/documents/upload or /api/documents/{categoryId}/fields
	http.HandleFunc("/api/documents/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/upload") {
			uploadDocumentHandler(w, r)
		} else if strings.Contains(r.URL.Path, "/fields") {
			saveExtractedFieldsHandler(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	log.Println("Server starting on :8080")
	log.Println("MongoDB:", config.DatabaseName)
	log.Println("Endpoints available:")
	log.Println("  GET  /api/categories")
	log.Println("  GET  /api/categories/{id}")
	log.Println("  GET  /api/categories/{categoryId}/formats")
	log.Println("  POST /api/documents/upload")
	log.Println("  POST /api/documents/{categoryIdOrName}/fields")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
