package analytics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	baseURL := os.Getenv("ANALYTICS_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8093"
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Get(path string, params map[string]string) (any, error) {
	u, _ := url.Parse(c.baseURL + path)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	return result, nil
}

func (c *Client) Health() (any, error) {
	return c.Get("/health", nil)
}

func (c *Client) OptimizePortfolio(symbols []string, strategy string) (any, error) {
	params := map[string]string{
		"strategy": strategy,
	}
	for i, s := range symbols {
		params[fmt.Sprintf("symbols[%d]", i)] = s
	}
	return c.Get("/api/optimization/optimize", params)
}

func (c *Client) CalculateVaR(returns []float64, confidence float64) (any, error) {
	params := map[string]string{
		"confidence": fmt.Sprintf("%f", confidence),
	}
	return c.Get("/api/risk/var", params)
}

func (c *Client) CalculateCAPM(riskFreeRate, marketReturn, beta float64) (any, error) {
	params := map[string]string{
		"risk_free_rate": fmt.Sprintf("%f", riskFreeRate),
		"market_return":  fmt.Sprintf("%f", marketReturn),
		"beta":           fmt.Sprintf("%f", beta),
	}
	return c.Get("/api/analytics/capm", params)
}
