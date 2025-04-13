package openrouter

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestClient(responseBody string, statusCode int) *Client {
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
			Header:     make(http.Header),
		}, nil
	})
	return &Client{
		apiKey: "test-key",
		httpClient: &http.Client{
			Transport: rt,
		},
	}
}

func TestClient_AnalyzeTerms_Success(t *testing.T) {
	resp := `{
		"id": "test",
		"object": "chat.completion",
		"created": 123,
		"model": "meta-llama/llama-4-maverick",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "{ \"terms\": [ { \"term\": \"ABS\", \"description\": \"anti-lock braking system\" } ] }"
			},
			"finish_reason": "stop"
		}],
		"usage": {
			"prompt_tokens": 10,
			"completion_tokens": 10,
			"total_tokens": 20
		}
	}`

	client := newTestClient(resp, http.StatusOK)
	analysis, err := client.AnalyzeTerms("ABS is a safety system.")
	require.NoError(t, err)
	require.NotNil(t, analysis)
	require.Len(t, analysis.Terms, 1)
	require.Equal(t, "ABS", analysis.Terms[0].Term)
	require.Equal(t, "anti-lock braking system", analysis.Terms[0].Description)
}
