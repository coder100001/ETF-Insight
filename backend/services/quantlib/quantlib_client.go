package quantlib

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"etf-insight/models"
)

type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

type cache struct {
	mu    sync.RWMutex
	items map[string]*cacheEntry
}

func newCache() *cache {
	c := &cache{
		items: make(map[string]*cacheEntry),
	}
	go c.cleanup()
	return c
}

func (c *cache) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.items {
			if now.After(entry.expiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

func (c *cache) get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, exists := c.items[key]
	if !exists {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		delete(c.items, key)
		return nil, false
	}
	return entry.data, true
}

func (c *cache) set(key string, data []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = &cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	cache      *cache
}

func NewClient() *Client {
	baseURL := os.Getenv("QUANTLIB_API_URL")
	if baseURL == "" {
		baseURL = "https://api.fincept.in/quantlib"
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: os.Getenv("QUANTLIB_API_KEY"),
		cache:  newCache(),
	}
}

func (c *Client) doRequest(method, endpoint string, body interface{}, result interface{}) error {
	var reqBody io.Reader

	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "ETF-Insight/1.0")

	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiResp models.QuantLibAPIResponse
		if json.Unmarshal(respBytes, &apiResp) == nil && apiResp.Message != "" {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, apiResp.Message)
		}
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBytes))
	}

	var apiResp models.QuantLibAPIResponse
	if err := json.Unmarshal(respBytes, &apiResp); err == nil && (apiResp.Success || apiResp.Message != "" || apiResp.Data != nil) {
		if !apiResp.Success {
			return fmt.Errorf("API error: %s", apiResp.Message)
		}
		if apiResp.Data != nil {
			dataBytes, err := json.Marshal(apiResp.Data)
			if err != nil {
				return fmt.Errorf("failed to re-marshal data: %w", err)
			}
			if err := json.Unmarshal(dataBytes, result); err != nil {
				return fmt.Errorf("failed to unmarshal data: %w", err)
			}
			return nil
		}
	}

	if err := json.Unmarshal(respBytes, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

func (c *Client) doRequestWithCache(method, endpoint string, body interface{}, result interface{}, ttl time.Duration) error {
	cacheKey := c.generateCacheKey(method, endpoint, body)
	if cachedData, found := c.cache.get(cacheKey); found {
		if err := json.Unmarshal(cachedData, result); err != nil {
			return fmt.Errorf("failed to unmarshal cached data: %w", err)
		}
		return nil
	}

	if err := c.doRequest(method, endpoint, body, result); err != nil {
		return err
	}

	// Cache the result for future requests
	if dataBytes, err := json.Marshal(result); err == nil {
		c.cache.set(cacheKey, dataBytes, ttl)
	}

	return nil
}

func (c *Client) generateCacheKey(method, endpoint string, body interface{}) string {
	h := sha256.New()
	h.Write([]byte(method))
	h.Write([]byte(endpoint))
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		h.Write(bodyBytes)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) PriceEuropeanOption(req models.EuropeanOptionRequest) (*models.OptionResult, error) {
	var result models.OptionResult
	if err := c.doRequest(http.MethodPost, "/options/european", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PriceAmericanOption(req models.AmericanOptionRequest) (*models.OptionResult, error) {
	var result models.OptionResult
	if err := c.doRequest(http.MethodPost, "/options/american", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CalculateGreeks(req models.GreeksRequest) (*models.OptionResult, error) {
	var result models.OptionResult
	if err := c.doRequest(http.MethodPost, "/options/greeks", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) BuildYieldCurve(req models.YieldCurveRequest) (*models.YieldCurveResult, error) {
	var result models.YieldCurveResult
	if err := c.doRequest(http.MethodPost, "/yield-curve/build", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) PriceBond(req models.BondRequest) (*models.BondResult, error) {
	var result models.BondResult
	if err := c.doRequest(http.MethodPost, "/bonds/fixed", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CalculateVaR(req models.VaRRequest) (*models.VaRResult, error) {
	var result models.VaRResult
	if err := c.doRequest(http.MethodPost, "/risk/var", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

const referenceDataCacheTTL = 1 * time.Hour

func (c *Client) GetSupportedCurrencies() (interface{}, error) {
	var result interface{}
	if err := c.doRequestWithCache(http.MethodGet, "/core/types/currencies", nil, &result, referenceDataCacheTTL); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetFrequencies() (interface{}, error) {
	var result interface{}
	if err := c.doRequestWithCache(http.MethodGet, "/core/types/frequencies", nil, &result, referenceDataCacheTTL); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetCalendars() (interface{}, error) {
	var result interface{}
	if err := c.doRequestWithCache(http.MethodGet, "/scheduling/calendar/list", nil, &result, referenceDataCacheTTL); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetDayCountConventions() (interface{}, error) {
	var result interface{}
	if err := c.doRequestWithCache(http.MethodGet, "/scheduling/daycount/conventions", nil, &result, referenceDataCacheTTL); err != nil {
		return nil, err
	}
	return result, nil
}
