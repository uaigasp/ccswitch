package usage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.anthropic.com"

type Window struct {
	UtilizationPercent float64 `json:"utilization"`
	ResetsAt           string  `json:"resets_at"`
}

type Result struct {
	FiveHour Window `json:"five_hour"`
	SevenDay Window `json:"seven_day"`
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient() *Client {
	return &Client{
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Fetch(accessToken string) (Result, error) {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/api/oauth/usage", nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")
	req.Header.Set("User-Agent", "ccswitch/1.0")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("usage endpoint devolvio status %d", resp.StatusCode)
	}

	var result Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Result{}, err
	}
	return result, nil
}
