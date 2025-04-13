package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
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
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("input text is empty")
	}
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

	// extract the response content
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := resp.Choices[0].Message.Content
	fmt.Printf("Raw API response content:\n%s\n", content)

	// try to find JSON in various formats
	var jsonStr string
	if strings.Contains(content, "```json") {
		// markdown code block format
		parts := strings.Split(content, "```json")
		if len(parts) > 1 {
			jsonStr = strings.Split(parts[1], "```")[0]
		}
	} else if strings.Contains(content, "```") {
		// generic code block format
		parts := strings.Split(content, "```")
		if len(parts) > 1 {
			jsonStr = parts[1]
		}
	} else {
		// try to find raw JSON
		start := strings.Index(content, "{")
		end := strings.LastIndex(content, "}")
		if start != -1 && end != -1 && end > start {
			jsonStr = content[start : end+1]
		}
	}

	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	// parse the JSON response
	var analysis TermAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("error parsing analysis: %w\nJSON content:\n%s", err, jsonStr)
	}

	// validate we got at least some terms
	if len(analysis.Terms) == 0 {
		return nil, fmt.Errorf("no terms found in analysis")
	}

	return &analysis, nil
}

// TranslateTextChunk translates a single chunk of text from English to Russian, preserving specified terms
func (c *Client) TranslateTextChunk(chunk string, terms []string) (string, error) {
	if strings.TrimSpace(chunk) == "" {
		return "", fmt.Errorf("input chunk is empty")
	}
	// join terms for the prompt
	termsList := ""
	for _, term := range terms {
		termsList += "- " + term + "\n"
	}

	prompt := fmt.Sprintf(translateTextPrompt, "English", "Russian", termsList, chunk)

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

	const maxAttempts = 3
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// send the request
		httpResp, err := c.httpClient.Do(httpReq)
		if err != nil {
			// check for temporary network errors
			var netErr net.Error
			if errors.Is(err, context.DeadlineExceeded) ||
				(errors.As(err, &netErr) && netErr.Timeout()) {
				lastErr = fmt.Errorf("temporary network error (attempt %d/%d): %w", attempt, maxAttempts, err)
				time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
				continue
			}
			return nil, fmt.Errorf("error sending request: %w", err)
		}

		// read the response body
		respBody, err := io.ReadAll(httpResp.Body)
		errClose := httpResp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("error reading response: %w", err)
		}
		if errClose != nil {
			return nil, fmt.Errorf("error closing response body: %w", errClose)
		}

		// log request and response for debugging
		fmt.Printf("Request URL: %s\n", httpReq.URL)
		fmt.Printf("Request Headers: %v\n", httpReq.Header)
		fmt.Printf("Response Status: %s\n", httpResp.Status)
		fmt.Printf("Response Body: %s\n", string(respBody))

		// check for errors
		if httpResp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API error: %s, body: %s", httpResp.Status, string(respBody))
			// retry on 5xx errors
			if httpResp.StatusCode >= 500 && httpResp.StatusCode < 600 && attempt < maxAttempts {
				time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
				continue
			}
			return nil, lastErr
		}

		// unmarshal the response
		var resp CompletionResponse
		if err := json.Unmarshal(respBody, &resp); err != nil {
			return nil, fmt.Errorf("error unmarshaling response: %w", err)
		}

		return &resp, nil
	}

	return nil, lastErr
}
