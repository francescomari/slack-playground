package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	URL         string
	AccessToken string
	HTTPClient  HTTPClient
}

type PostMessageRequest struct {
	Channel string `json:"channel,omitempty"`
	Text    string `json:"text,omitempty"`
}

type PostMessageResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func (c *Client) PostMessage(ctx context.Context, request *PostMessageRequest) (*PostMessageResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal arguments: %v", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/chat.postMessage", c.URL), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %v", err)
	}

	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))

	httpResponse, err := c.HTTPClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("perform request: %v", err)
	}

	var response PostMessageResponse

	if err := json.NewDecoder(httpResponse.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %v", err)
	}

	return &response, nil
}
