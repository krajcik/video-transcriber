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

func TestClient_AnalyzeTerms_EmptyInput(t *testing.T) {
	client := newTestClient("", http.StatusOK)
	analysis, err := client.AnalyzeTerms("")
	require.Error(t, err)
	require.Nil(t, analysis)
	require.Contains(t, err.Error(), "input text is empty")
}

func TestClient_TranslateTextChunk_EmptyInput(t *testing.T) {
	client := newTestClient("", http.StatusOK)
	result, err := client.TranslateTextChunk("", nil)
	require.Error(t, err)
	require.Empty(t, result)
	require.Contains(t, err.Error(), "input chunk is empty")
}

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return false }

func TestClient_createCompletion_RetryOnTimeout(t *testing.T) {
	attempts := 0
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts < 3 {
			// simulate timeout error
			return nil, timeoutErr{}
		}
		// success on 3rd attempt
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
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(resp)),
			Header:     make(http.Header),
		}, nil
	})
	client := &Client{
		apiKey: "test-key",
		httpClient: &http.Client{
			Transport: rt,
		},
	}
	analysis, err := client.AnalyzeTerms("ABS is a safety system.")
	require.NoError(t, err)
	require.NotNil(t, analysis)
	require.Equal(t, 3, attempts)
}

type errBody struct {
	io.Reader
}

func (e *errBody) Close() error {
	return io.ErrUnexpectedEOF
}

func TestClient_createCompletion_ErrorOnBodyClose(t *testing.T) {
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
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
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       &errBody{Reader: bytes.NewBufferString(resp)},
			Header:     make(http.Header),
		}, nil
	})
	client := &Client{
		apiKey: "test-key",
		httpClient: &http.Client{
			Transport: rt,
		},
	}
	analysis, err := client.AnalyzeTerms("ABS is a safety system.")
	require.Error(t, err)
	require.Nil(t, analysis)
	require.Contains(t, err.Error(), "error closing response body")
}
