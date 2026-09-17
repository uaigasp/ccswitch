package usage

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchParsesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("header authorization incorrecto: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("anthropic-beta") != "oauth-2025-04-20" {
			t.Fatalf("header anthropic-beta incorrecto")
		}

		body := map[string]any{
			"five_hour": map[string]any{"utilization": 12.5, "resets_at": "2026-09-17T05:10:00Z"},
			"seven_day": map[string]any{"utilization": 8.0, "resets_at": "2026-09-22T04:00:00Z"},
		}
		json.NewEncoder(w).Encode(body)
	}))
	defer server.Close()

	c := NewClient()
	c.BaseURL = server.URL

	result, err := c.Fetch("test-token")
	if err != nil {
		t.Fatalf("fetch fallo: %v", err)
	}

	if result.FiveHour.UtilizationPercent != 12.5 {
		t.Fatalf("got five hour %v", result.FiveHour.UtilizationPercent)
	}
	if result.SevenDay.UtilizationPercent != 8.0 {
		t.Fatalf("got seven day %v", result.SevenDay.UtilizationPercent)
	}
}

func TestFetchReturnsErrorOnBadStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	c := NewClient()
	c.BaseURL = server.URL

	_, err := c.Fetch("test-token")
	if err == nil {
		t.Fatal("esperaba error por status 429")
	}
}
