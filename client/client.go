package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	BaseURL = "https://app.asana.com/api/1.0"
)

type Client struct {
	httpClient  *http.Client
	accessToken string
	baseURL     string
}

func NewClient(accessToken string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		accessToken: accessToken,
		baseURL:     BaseURL,
	}
}

func (c *Client) doRequest(method, endpoint string, params url.Values, body interface{}) ([]byte, error) {
	fullURL := c.baseURL + endpoint
	if params != nil && len(params) > 0 {
		fullURL = fullURL + "?" + params.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errorResp struct {
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(respBody, &errorResp); err == nil && len(errorResp.Errors) > 0 {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errorResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return respBody, nil
}

func (c *Client) Get(endpoint string, params url.Values) ([]byte, error) {
	return c.doRequest("GET", endpoint, params, nil)
}

func (c *Client) Post(endpoint string, body interface{}) ([]byte, error) {
	return c.doRequest("POST", endpoint, nil, body)
}

func (c *Client) PostForm(endpoint string, formData url.Values) ([]byte, error) {
	fullURL := c.baseURL + endpoint

	req, err := http.NewRequest("POST", fullURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errorResp struct {
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(respBody, &errorResp); err == nil && len(errorResp.Errors) > 0 {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errorResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return respBody, nil
}

func (c *Client) Put(endpoint string, body interface{}) ([]byte, error) {
	return c.doRequest("PUT", endpoint, nil, body)
}

func (c *Client) Delete(endpoint string) ([]byte, error) {
	return c.doRequest("DELETE", endpoint, nil, nil)
}

func (c *Client) GetHTTPClient() *http.Client {
	return c.httpClient
}

func (c *Client) GetAccessToken() string {
	return c.accessToken
}

func (c *Client) GetBaseURL() string {
	return c.baseURL
}

func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

func (c *Client) SetAccessToken(token string) {
	c.accessToken = token
}

func (c *Client) SetHTTPClient(httpClient *http.Client) {
	c.httpClient = httpClient
}
