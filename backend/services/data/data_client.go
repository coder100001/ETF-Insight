package data

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
	baseURL := os.Getenv("DATA_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8092"
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Get(path string, params map[string]string) (interface{}, error) {
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

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	return result, nil
}

func (c *Client) Health() (interface{}, error) {
	return c.Get("/health", nil)
}

func (c *Client) FredSeries(seriesID string, params map[string]string) (interface{}, error) {
	return c.Get(fmt.Sprintf("/api/fred/series/%s", seriesID), params)
}

func (c *Client) WorldBankIndicators(country, indicator string, params map[string]string) (interface{}, error) {
	return c.Get(fmt.Sprintf("/api/worldbank/indicators/%s/%s", country, indicator), params)
}

func (c *Client) YFinanceQuote(symbol string) (interface{}, error) {
	return c.Get(fmt.Sprintf("/api/yfinance/quote/%s", symbol), nil)
}

func (c *Client) YFinanceHistorical(symbol string, params map[string]string) (interface{}, error) {
	return c.Get(fmt.Sprintf("/api/yfinance/historical/%s", symbol), params)
}

func (c *Client) AkShareStockSpot() (interface{}, error) {
	return c.Get("/api/akshare/stock/spot", nil)
}

func (c *Client) CoinGeckoPrice(ids, vsCurrencies string) (interface{}, error) {
	return c.Get("/api/coingecko/price", map[string]string{"ids": ids, "vs_currencies": vsCurrencies})
}
