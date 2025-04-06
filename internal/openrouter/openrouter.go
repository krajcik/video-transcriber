package openrouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL        = "https://openrouter.ai/api/v1"
	defaultTimeout = 300 * time.Second
)

// Client represents the OpenRouter API client
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// Message represents a message in the OpenRouter API
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest represents a request to the completions endpoint
type CompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// CompletionResponse represents a response from the completions endpoint
type CompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// TermAnalysis represents the result of analyzing text for terms
type TermAnalysis struct {
	Terms []struct {
		Term        string   `json:"term"`
		Description string   `json:"description"`
		Context     []string `json:"context,omitempty"`
	} `json:"terms"`
}

// New creates a new OpenRouter client
func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// AnalyzeTerms analyzes text to identify terms that should not be translated
func (c *Client) AnalyzeTerms(text string) (*TermAnalysis, error) {
	prompt := fmt.Sprintf(analyzeTermsPrompt, text)

	// create the completion request
	req := CompletionRequest{
		Model: "meta-llama/llama-4-maverick",
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// send the request to OpenRouter
	resp, err := c.createCompletion(req)
	if err != nil {
		return nil, fmt.Errorf("error getting completion: %w", err)
	}

	// extract the JSON response
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := resp.Choices[0].Message.Content

	// extract JSON from response (handles both raw JSON and markdown with explanations)
	jsonStart := strings.Index(content, "```json")
	if jsonStart != -1 {
		// if markdown format, extract the JSON part
		content = content[jsonStart+7:] // skip "```json"
		jsonEnd := strings.LastIndex(content, "```")
		if jsonEnd != -1 {
			content = content[:jsonEnd]
		}
	} else {
		// if no markdown, look for the first {
		jsonStart = strings.Index(content, "{")
		if jsonStart != -1 {
			content = content[jsonStart:]
		}
	}

	content = strings.TrimSpace(content)

	// parse the JSON response
	var analysis TermAnalysis
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		return nil, fmt.Errorf("error parsing analysis: %w (content: %s)", err, content)
	}

	return &analysis, nil
}

// TranslateTextChunk translates a single chunk of text from English to Russian, preserving specified terms
func (c *Client) TranslateTextChunk(chunk string, terms []string) (string, error) {
	// join terms for the prompt
	termsList := ""
	for _, term := range terms {
		termsList += "- " + term + "\n"
	}

	prompt := fmt.Sprintf(translateTextPrompt, termsList, chunk)

	// create the completion request
	req := CompletionRequest{
		Model: "meta-llama/llama-4-maverick",
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// send the request to OpenRouter
	resp, err := c.createCompletion(req)
	if err != nil {
		return "", fmt.Errorf("error getting translation: %w", err)
	}

	// extract the text from the response
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return resp.Choices[0].Message.Content, nil
}

// TranslateText translates text from English to Russian in chunks, preserving specified terms
func (c *Client) TranslateText(text string, terms []string) (string, error) {
	// split text into paragraphs
	paragraphs := splitIntoParagraphs(text)

	// maximum paragraphs per chunk (adjust as needed based on token limits)
	maxParagraphsPerChunk := 5

	var translatedText string
	var chunks []string

	// group paragraphs into chunks
	for i := 0; i < len(paragraphs); i += maxParagraphsPerChunk {
		end := i + maxParagraphsPerChunk
		if end > len(paragraphs) {
			end = len(paragraphs)
		}

		chunk := strings.Join(paragraphs[i:end], "\n\n")
		chunks = append(chunks, chunk)
	}

	fmt.Printf("Translating text in %d chunks...\n", len(chunks))

	// translate each chunk
	for i, chunk := range chunks {
		fmt.Printf("Translating chunk %d of %d...\n", i+1, len(chunks))

		translatedChunk, err := c.TranslateTextChunk(chunk, terms)
		if err != nil {
			return "", fmt.Errorf("error translating chunk %d: %w", i+1, err)
		}

		if i > 0 {
			translatedText += "\n\n"
		}
		translatedText += translatedChunk
	}

	return translatedText, nil
}

// splitIntoParagraphs splits text into paragraphs
func splitIntoParagraphs(text string) []string {
	// split by double newlines
	paragraphs := strings.Split(text, "\n\n")

	// filter out empty paragraphs
	var filteredParagraphs []string
	for _, p := range paragraphs {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			filteredParagraphs = append(filteredParagraphs, trimmed)
		}
	}

	// if there are no clear paragraphs, try to split by single newlines
	if len(filteredParagraphs) <= 1 {
		lines := strings.Split(text, "\n")

		// group consecutive non-empty lines into paragraphs
		var currentParagraph string
		filteredParagraphs = []string{}

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				if currentParagraph != "" {
					filteredParagraphs = append(filteredParagraphs, currentParagraph)
					currentParagraph = ""
				}
			} else {
				if currentParagraph != "" {
					currentParagraph += " "
				}
				currentParagraph += trimmed
			}
		}

		if currentParagraph != "" {
			filteredParagraphs = append(filteredParagraphs, currentParagraph)
		}
	}

	// if still no paragraphs, split by sentences as a last resort
	if len(filteredParagraphs) <= 1 && len(strings.TrimSpace(text)) > 1000 {
		// simple regex-free sentence splitting (not perfect but should work for most cases)
		sentences := splitIntoSentences(text)

		// group sentences into reasonable-sized paragraphs
		const maxSentencesPerParagraph = 5
		filteredParagraphs = []string{}
		var currentParagraph string
		sentenceCount := 0

		for _, sentence := range sentences {
			trimmed := strings.TrimSpace(sentence)
			if trimmed == "" {
				continue
			}

			if sentenceCount >= maxSentencesPerParagraph {
				if currentParagraph != "" {
					filteredParagraphs = append(filteredParagraphs, currentParagraph)
				}
				currentParagraph = trimmed
				sentenceCount = 1
			} else {
				if currentParagraph != "" {
					currentParagraph += " "
				}
				currentParagraph += trimmed
				sentenceCount++
			}
		}

		if currentParagraph != "" {
			filteredParagraphs = append(filteredParagraphs, currentParagraph)
		}
	}

	return filteredParagraphs
}

// splitIntoSentences splits text into sentences
func splitIntoSentences(text string) []string {
	// this is a simplified version that handles common ending punctuation
	text = strings.ReplaceAll(text, "! ", "!|")
	text = strings.ReplaceAll(text, "? ", "?|")
	text = strings.ReplaceAll(text, ". ", ".|")

	// handle special cases for common abbreviations
	text = strings.ReplaceAll(text, "Mr.|", "Mr. ")
	text = strings.ReplaceAll(text, "Mrs.|", "Mrs. ")
	text = strings.ReplaceAll(text, "Dr.|", "Dr. ")
	text = strings.ReplaceAll(text, "Prof.|", "Prof. ")
	text = strings.ReplaceAll(text, "i.e.|", "i.e. ")
	text = strings.ReplaceAll(text, "e.g.|", "e.g. ")
	text = strings.ReplaceAll(text, "vs.|", "vs. ")

	return strings.Split(text, "|")
}

// createCompletion sends a completion request to the OpenRouter API
func (c *Client) createCompletion(req CompletionRequest) (*CompletionResponse, error) {
	// marshal the request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// create the HTTP request
	httpReq, err := http.NewRequest("POST", baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// set the headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/AssemblyAI/assemblyai-go-sdk")

	// send the request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer httpResp.Body.Close()

	// read the response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// log request and response for debugging
	fmt.Printf("Request URL: %s\n", httpReq.URL)
	fmt.Printf("Request Headers: %v\n", httpReq.Header)
	fmt.Printf("Response Status: %s\n", httpResp.Status)
	fmt.Printf("Response Body: %s\n", string(respBody))

	// check for errors
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s, body: %s", httpResp.Status, string(respBody))
	}

	// unmarshal the response
	var resp CompletionResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &resp, nil
}
