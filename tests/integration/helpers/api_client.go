package helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

type APIClient struct {
	BaseURL string
	T       *testing.T
}

func NewAPIClient(baseURL string, t *testing.T) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		T:       t,
	}
}

func (c *APIClient) SendRequest(method, path string, body interface{}) *http.Response {
	var reqBody io.Reader
	if body != nil {
		if m, ok := body.(map[string]interface{}); ok {
			for key, value := range m {
				if t, ok := value.(time.Time); ok {
					m[key] = t.Unix()
				}
			}
		}
		jsonData, err := json.Marshal(body)
		if err != nil {
			c.T.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		c.T.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.T.Fatalf("Failed to send request: %v", err)
	}
	return resp
}

func (c *APIClient) ParseResponse(resp *http.Response, result interface{}) {
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		c.T.Fatalf("Failed to parse response body: %v", err)
	}
}
