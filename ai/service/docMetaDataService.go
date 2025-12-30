package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func DocMetaDataService(pdfPath string) (map[string]interface{}, error) {

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: OPENAI_API_KEY environment variable not set")
	}

	fmt.Printf("filepath: %s\n", pdfPath)
	fmt.Print(pdfPath)

	// Check if PDF exists
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		log.Fatal("Error: PDF file '%s' not found\n", pdfPath)
	}

	// Extract ALL key-value,description pairs from the PDF
	fmt.Println("=== Extracting ALL key-value pairs from PDF ===")
	extractedData, err := ExtractDataFromPDF(apiKey, pdfPath)
	if err != nil {
		log.Fatal("Error extracting data:", err)
	}

	fmt.Println("\nExtracted Data:")
	for key, value := range extractedData {
		fmt.Printf("  %s: %v\n", key, value)
	}

	return extractedData, nil
}

// OpenAIRequest represents the request structure for OpenAI Responses API
type OpenAIRequest struct {
	Model string      `json:"model"`
	Input []InputItem `json:"input"`
}

// InputItem represents an input item in the conversation
type InputItem struct {
	Role    string        `json:"role"`
	Content []ContentItem `json:"content"`
}

// ContentItem represents a content item (text or file)
type ContentItem struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Filename string `json:"filename,omitempty"`
	FileData string `json:"file_data,omitempty"`
}

// OpenAIResponse represents the response from OpenAI Responses API
type OpenAIResponse struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Model      string `json:"model"`
	OutputText string `json:"output_text"`
	Output     []struct {
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

// ExtractDataFromPDF uses OpenAI Responses API to extract all key-value pairs from a PDF
func ExtractDataFromPDF(apiKey, pdfPath string) (map[string]interface{}, error) {
	// Read PDF and convert to base64
	fmt.Println("Reading PDF file...")
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	base64PDF := base64.StdEncoding.EncodeToString(pdfBytes)

	// 	// Create the prompt for extraction
	// 	prompt := `Analyze this document and extract ALL key-value pairs you can find.

	// **Instructions:**
	// - Return ONLY a valid JSON object (no markdown, no code blocks, no explanation)
	// - Extract every piece of structured information (labels and their values)
	// - The document could contain key-value
	// - keyValues - key-value pair data. Include keys with blank values also. Each key-value pair should include
	// 	- key - The extracted field name
	//  	- value - The extracted field value.
	//     -description - A brief description of the field, explaining its significance and context within the document.
	// - Use clear, descriptive keys in snake_case format
	// - Each value should be a string
	// - Add description to key value
	// - Extract dates in YYYY-MM-DD format when possible
	// - Include amounts, numbers, names, dates, addresses, and any other relevant information
	// - Group related information logically if appropriate
	// - Be comprehensive - extract everything that appears to be a labeled field or data point

	// Return the JSON object now:`
	// Create the prompt for extraction
	prompt := `Analyze this document and extract ALL key-value pairs you can find.

		**Instructions:**
		- Return ONLY a valid JSON object (no markdown, no code blocks, no explanation)
		- The JSON must have this exact structure:
		{
		"key_values": [
			{
			"key": "field_name",
			"value": "field value",
			"description": "brief description of what this field represents"
			}
		]
		}
		- Extract every piece of structured information (labels and their values) into the key_values array
		- Include keys with blank/empty values also
		- Each object in the key_values array must have exactly three fields:
		* key - The extracted field name in snake_case format (e.g., contract_no, tender_date)
		* value - The extracted field value as a string (even if empty, use empty string "")
		* description - A brief description explaining the significance and context of this field
		- Extract dates in YYYY-MM-DD format when possible
		- Extract all amounts, numbers, names, dates, addresses, and any other relevant information
		- Be comprehensive - extract everything that appears to be a labeled field or data point
		- If the document has sections or categories, still flatten all fields into the single key_values array
		- Use clear, descriptive keys that indicate what the field represents

		Return the JSON object now:`

	// Construct the OpenAI request using the Responses API format
	requestBody := OpenAIRequest{
		Model: "gpt-4o",
		Input: []InputItem{
			{
				Role: "user",
				Content: []ContentItem{
					{
						Type: "input_text",
						Text: prompt,
					},
					{
						Type:     "input_file",
						Filename: "document.pdf",
						FileData: fmt.Sprintf("data:application/pdf;base64,%s", base64PDF),
					},
				},
			},
		},
	}

	// Marshal request to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request to the Responses API endpoint
	fmt.Println("Sending request to OpenAI Responses API...")
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/responses", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse OpenAI response
	var openaiResp OpenAIResponse
	err = json.Unmarshal(body, &openaiResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w\nResponse body: %s", err, string(body))
	}

	// Get content from the response - try multiple fields
	content := openaiResp.OutputText
	if content == "" && len(openaiResp.Output) > 0 && len(openaiResp.Output[0].Content) > 0 {
		content = openaiResp.Output[0].Content[0].Text
	}

	if content == "" {
		return nil, fmt.Errorf("empty response from OpenAI. Full response: %s", string(body))
	}

	// Remove markdown code blocks if present
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Try to find JSON object in the content
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		content = content[startIdx : endIdx+1]
	}

	// Parse the extracted data
	var extractedData map[string]interface{}
	err = json.Unmarshal([]byte(content), &extractedData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extracted data: %w\nCleaned Content: %s", err, content)
	}
	return extractedData, nil
}

// Flatten takes a nested map and returns a new one where nested maps are replaced
// by dot-delimited keys.
func Flatten(m map[string]interface{}) map[string]interface{} {
	o := make(map[string]interface{})
	for k, v := range m {
		switch child := v.(type) {
		case map[string]interface{}:
			// If the value is a nested map, recursively flatten it
			nm := Flatten(child)
			for nk, nv := range nm {
				// Combine parent key and child key with a dot delimiter
				o[k+"."+nk] = nv
			}
		case []interface{}:
			// Handle slices/arrays by iterating and flattening each element
			for i, iv := range child {
				// Create an indexed key (e.g., "key.0", "key.1")
				o[fmt.Sprintf("%s.%d", k, i)] = iv
			}
		default:
			// For basic types, just add the key-value pair to the result
			o[k] = v
		}
	}
	return o
}
