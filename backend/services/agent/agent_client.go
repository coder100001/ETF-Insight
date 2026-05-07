package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type AgentInfo struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Category            string `json:"category"`
	Description         string `json:"description"`
	SystemPromptPreview string `json:"system_prompt_preview"`
}

type AgentRunRequest struct {
	AgentID     string         `json:"agent_id"`
	Query       string         `json:"query"`
	Context     map[string]any `json:"context,omitempty"`
	LLMProvider string         `json:"llm_provider"`
	Model       string         `json:"model"`
	Temperature float64        `json:"temperature,omitempty"`
	MaxTokens   int            `json:"max_tokens,omitempty"`
}

type AgentRunResponse struct {
	AgentID    string `json:"agent_id"`
	AgentName  string `json:"agent_name"`
	Response   string `json:"response"`
	Model      string `json:"model"`
	TokensUsed int    `json:"tokens_used"`
	DurationMs int    `json:"duration_ms"`
}

type AgentTeamRequest struct {
	AgentIDs    []string `json:"agent_ids"`
	Query       string   `json:"query"`
	Rounds      int      `json:"rounds"`
	LLMProvider string   `json:"llm_provider"`
	Model       string   `json:"model"`
}

type AgentTeamResponse struct {
	Query     string           `json:"query"`
	Rounds    []map[string]any `json:"rounds"`
	Synthesis string           `json:"synthesis"`
}

type apiEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	baseURL := os.Getenv("AGENT_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8091"
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, endpoint string, body any, result any) error {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBytes))
	}

	var envelope apiEnvelope
	if err := json.Unmarshal(respBytes, &envelope); err == nil && envelope.Data != nil {
		if !envelope.Success {
			return fmt.Errorf("API error: %s", envelope.Error)
		}
		return json.Unmarshal(envelope.Data, result)
	}

	return json.Unmarshal(respBytes, result)
}

func (c *Client) Health() (map[string]any, error) {
	var result map[string]any
	if err := c.doRequest(http.MethodGet, "/health", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) Discover() ([]AgentInfo, error) {
	var result []AgentInfo
	if err := c.doRequest(http.MethodGet, "/agents/discover", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) Run(req AgentRunRequest) (*AgentRunResponse, error) {
	var result AgentRunResponse
	if err := c.doRequest(http.MethodPost, "/agents/run", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) RunTeam(req AgentTeamRequest) (*AgentTeamResponse, error) {
	var result AgentTeamResponse
	if err := c.doRequest(http.MethodPost, "/agents/team", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
