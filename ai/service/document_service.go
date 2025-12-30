package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

var OPENAI_KEY = os.Getenv("OPENAI_API_KEY")

func Docservice(pdfPath string) (*PDFCategory, error) {

	// Set PDF path directly or use command line
	// pdfPath := "C:/projects/exto-go/ai/uploads/1765249316_Bikaner-part-11.pdf" // Change this to your PDF path

	apiKey := OPENAI_KEY
	// os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	fmt.Printf("Analyzing PDF: %s\n\n", pdfPath)

	// Extract text from PDF
	text, pageCount, err := extractTextFromPDF(pdfPath)
	if err != nil {
		log.Fatalf("Error extracting text: %v", err)
	}

	fmt.Printf("Extracted %d pages, %d characters\n", pageCount, len(text))

	// Categorize using OpenAI
	category, err := categorizePDF(text, apiKey)
	if err != nil {
		log.Fatalf("Error categorizing PDF: %v", err)
	}

	// Display results
	fmt.Println("\n=== PDF CATEGORIZATION RESULTS ===")
	fmt.Printf("Category: %s\n", category.Category)
	fmt.Printf("Sub-Category: %s\n", category.SubCategory)
	fmt.Printf("Confidence: %s\n", category.Confidence)
	fmt.Printf("Summary: %s\n", category.Summary)

	if len(category.Keywords) > 0 {
		fmt.Printf("Keywords: %s\n", strings.Join(category.Keywords, ", "))
	}
	fmt.Printf("CATEGORY : ")
	fmt.Printf("%s", category.Category)
	fmt.Println("ERROR : ")
	fmt.Println(err)

	// if len(category.Metadata) > 0 {
	// 	fmt.Println("\nMetadata:")
	// 	for key, value := range category.Metadata {
	// 		fmt.Printf("  %s: %v\n", key, value)
	// 	}
	// }
	return category, err
}

// PDFCategory holds the categorization result
type PDFCategory struct {
	Category    string                 `json:"category"`
	SubCategory string                 `json:"sub_category"`
	Confidence  string                 `json:"confidence"`
	Keywords    []string               `json:"keywords"`
	Summary     string                 `json:"summary"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// OpenAICategoryRequest for PDF categorization
type OpenAICategoryRequest struct {
	Model       string        `json:"model"`
	Messages    []CategoryMsg `json:"messages"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
}

type CategoryMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAICategoryResponse from API
type OpenAICategoryResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// extractTextFromPDF extracts all text content from a PDF file
func extractTextFromPDF(pdfPath string) (string, int, error) {
	f, r, err := pdf.Open(pdfPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	totalPages := r.NumPage()
	var textBuilder strings.Builder

	// Extract text from each page
	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			log.Printf("Warning: failed to extract text from page %d: %v", pageNum, err)
			continue
		}

		textBuilder.WriteString(text)
		textBuilder.WriteString("\n\n")
	}

	return textBuilder.String(), totalPages, nil
}

// categorizePDF sends the PDF text to OpenAI for categorization
func categorizePDF(text string, apiKey string) (*PDFCategory, error) {
	// Limit text length to avoid token limits (approximately 6000 words)
	maxChars := 24000
	if len(text) > maxChars {
		text = text[:maxChars] + "... [truncated]"
	}

	prompt := fmt.Sprintf(`Analyze the following PDF content and categorize it. Respond with ONLY a JSON object (no markdown, no extra text):
	{
		"category": "primary category (e.g., Invoice, Contract, Report, Resume, Legal Document, Medical Record, Receipt, etc.)",
		"sub_category": "more specific type",
		"confidence": "high/medium/low",
		"keywords": ["key", "terms", "found"],
		"summary": "brief 1-2 sentence summary",
		"metadata": {
				key: value
		}
	}
	PDF Content:
	%s`, text)

	reqBody := OpenAICategoryRequest{
		Model:       "gpt-4o-mini", // Fast and cost-effective for text analysis
		MaxTokens:   1000,
		Temperature: 0.0,
		Messages: []CategoryMsg{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request to OpenAI API
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var openaiResp OpenAICategoryResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if openaiResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", openaiResp.Error.Message)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	// Parse the JSON from the text response
	content := openaiResp.Choices[0].Message.Content
	content = cleanJSONResponse(content)
	fmt.Print("content### ")
	fmt.Print(content)

	var category PDFCategory
	if err := json.Unmarshal([]byte(content), &category); err != nil {
		return nil, fmt.Errorf("failed to parse category result: %w\nRaw content: %s", err, content)
	}

	return &category, nil
}

// cleanJSONResponse removes markdown code blocks if present
func cleanJSONResponse(content string) string {
	content = strings.TrimSpace(content)

	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
	}

	if strings.HasSuffix(content, "```") {
		content = strings.TrimSuffix(content, "```")
	}

	return strings.TrimSpace(content)
}
