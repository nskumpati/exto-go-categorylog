package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gaeaglobal/exto/server/app"
	openai "github.com/sashabaranov/go-openai"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ExtractedTemplateField struct {
	CategoryFieldName string   `json:"category_field_name"`
	PromptText        string   `json:"prompt_text"`
	SampleValues      []string `json:"sample_values"`
}

type KeyValue struct {
	Key             string      `json:"key"`
	Value           interface{} `json:"value"` // Accepts string, array, or object
	ConfidenceScore int         `json:"confidenceScore"`
}

type KeyValuesResponse struct {
	KeyValues []KeyValue `json:"keyValues"`
}

type OpenAIService struct {
	formatService *FormatService
}

func NewOpenAIService(formatService *FormatService) *OpenAIService {
	return &OpenAIService{formatService: formatService}
}

func (s *OpenAIService) ExtractDocumentData(reqCtx *app.RequestContext, categoryID bson.ObjectID, base64Image string) (any, error) {

	formats, err := s.formatService.GetFormatsByCategoryID(reqCtx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get formats: %v", err)
	}

	var extractedTemplateFields []ExtractedTemplateField

	for _, format := range formats {
		for _, field := range format.ExtractionFields {
			extracted := ExtractedTemplateField{
				CategoryFieldName: field.CategoryFieldName,
				PromptText:        field.Prompt.Text,
				SampleValues:      field.Prompt.SampleValues,
			}
			extractedTemplateFields = append(extractedTemplateFields, extracted)
		}
	}

	// Marshal extracted template fields to JSON
	templateFieldsJSON, err := json.MarshalIndent(extractedTemplateFields, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template fields: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(apiKey)

	systemPrompt := `
You are an intelligent document data extraction system.

You will be provided with:
1. An image of a document in base64 format.
2. A list of template fields (see below as JSON):
` + string(templateFieldsJSON) + `

Each template field contains:
- category_field_name: The key under which extracted value should be stored (in snake_case).
- prompt_text: Description of what to extract.
- sample_values: A few example values to guide extraction.

Your task is to extract data **only** for the fields defined in the template. Do not hallucinate or infer fields not present in the template.

Use the document image and the template fields to extract key-value pairs. Output a JSON with the following structure:

{
"keyValues": [
    {
    "key": "<category_field_name from template>",
    "value": "<extracted value from image>",
    "confidenceScore": <percentage from 1 to 100>
    },
    ...
]
}

### Confidence Score Calculation Guidelines:
Evaluate confidenceScore dynamically based on:
- **Text Clarity**: Is the value legible and artifact-free?
- **Structural Correctness**: Does the value follow expected format (e.g., dates, currency)?
- **Pattern Consistency**: Does the value align with sample values or domain patterns?
- **Model/OCR Confidence**: Use internal uncertainty or scoring metrics if available.

Return only JSON in the output. Do not include explanations or additional text.

---
Begin extraction based on the provided image and template fields.
`

	userMessage := openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleUser,
		MultiContent: []openai.ChatMessagePart{
			{
				Type: openai.ChatMessagePartTypeText,
				Text: "The document image is attached below.",
			},
			{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL:    base64Image,
					Detail: openai.ImageURLDetailAuto,
				},
			},
		},
	}

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: "gpt-4o",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			userMessage,
		},
		MaxTokens: 2000,
	})

	if err != nil {
		return nil, err
	}

	return resp.Choices[0].Message.Content, nil
}

func (s *OpenAIService) ConvertKeyValueToMap(jsonStr string) (map[string]any, error) {
	fmt.Println("Converting key values from JSON:", jsonStr)
	// Remove possible code block markers and "json\n" prefix
	cleaned := strings.TrimPrefix(jsonStr, "json\n")
	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var resp KeyValuesResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return nil, err
	}

	result := make(map[string]any)
	for _, kv := range resp.KeyValues {
		result[kv.Key] = kv.Value // Value can be string, []any, or map[string]any
	}
	return result, nil
}

func (s *OpenAIService) ConvertConfidenceScoresToMap(jsonStr string) (map[string]int, error) {
	fmt.Println("Converting confidence scores from JSON:", jsonStr)
	// Remove possible code block markers and "json\n" prefix
	cleaned := strings.TrimPrefix(jsonStr, "json\n")
	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var resp KeyValuesResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return nil, err
	}

	result := make(map[string]int)
	for _, kv := range resp.KeyValues {
		result[kv.Key] = kv.ConfidenceScore
	}
	return result, nil
}

func (s *OpenAIService) CalculateAverageConfidence(confidenceMap map[string]int) float64 {
	if len(confidenceMap) == 0 {
		return 0
	}
	var sum int
	for _, v := range confidenceMap {
		sum += v
	}
	return float64(sum) / float64(len(confidenceMap))
}
