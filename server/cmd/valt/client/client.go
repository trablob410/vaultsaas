package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for the Valt API.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// New creates a new Client.
func New(baseURL, token string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Token:      token,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ListSecrets returns the list of secrets.
func (c *Client) ListSecrets() ([]map[string]interface{}, error) {
	result, err := c.get("/secrets")
	if err != nil {
		return nil, err
	}
	if secrets, ok := result["secrets"].([]interface{}); ok {
		out := make([]map[string]interface{}, 0, len(secrets))
		for _, s := range secrets {
			if m, ok := s.(map[string]interface{}); ok {
				out = append(out, m)
			}
		}
		return out, nil
	}
	return []map[string]interface{}{}, nil
}

// CreateSecret creates a new secret.
func (c *Client) CreateSecret(name, credType, value string) (map[string]interface{}, error) {
	return c.post("/secrets", map[string]interface{}{
		"name":            name,
		"credential_type": credType,
		"plaintext_value": value,
	})
}

// ListAgents returns agents for a project.
func (c *Client) ListAgents(projectID string) ([]map[string]interface{}, error) {
	result, err := c.get(fmt.Sprintf("/projects/%s/agents", projectID))
	if err != nil {
		return nil, err
	}
	if agents, ok := result["agents"].([]interface{}); ok {
		out := make([]map[string]interface{}, 0, len(agents))
		for _, a := range agents {
			if m, ok := a.(map[string]interface{}); ok {
				out = append(out, m)
			}
		}
		return out, nil
	}
	return []map[string]interface{}{}, nil
}

// CreateAgent creates a new agent.
func (c *Client) CreateAgent(projectID, name string) (map[string]interface{}, error) {
	return c.post(fmt.Sprintf("/projects/%s/agents", projectID), map[string]interface{}{
		"name": name,
	})
}

func (c *Client) get(path string) (map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) post(path string, body interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

func (c *Client) do(req *http.Request) (map[string]interface{}, error) {
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(raw))
	}
	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return result, nil
}
