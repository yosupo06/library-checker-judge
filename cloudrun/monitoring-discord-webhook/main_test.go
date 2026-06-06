package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleForwardsMonitoringPayloadToDiscord(t *testing.T) {
	var got discordWebhookMessage
	discord := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode discord payload: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer discord.Close()

	s := &server{
		discordWebhook: discord.URL,
		authToken:      "secret-token",
		client:         discord.Client(),
	}

	req := httptest.NewRequest(http.MethodPost, "/?auth_token=secret-token", strings.NewReader(`{
		"incident": {
			"state": "OPEN",
			"policy_name": "Judge disk usage high",
			"condition_name": "Judge disk usage above 85%",
			"resource_display_name": "judge-abc",
			"summary": "Disk usage crossed the alert threshold.",
			"url": "https://console.cloud.google.com/monitoring/alerting/incidents/123"
		}
	}`))
	rr := httptest.NewRecorder()

	s.handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if len(got.Embeds) != 1 {
		t.Fatalf("embeds = %d, want 1", len(got.Embeds))
	}
	embed := got.Embeds[0]
	if !strings.Contains(embed.Title, "Judge disk usage high") {
		t.Fatalf("title = %q, want policy name", embed.Title)
	}
	if embed.Color != 0xff3333 {
		t.Fatalf("color = %#x, want open alert color", embed.Color)
	}
	if embed.URL == "" {
		t.Fatal("url is empty")
	}
}

func TestHandleRejectsInvalidToken(t *testing.T) {
	s := &server{
		discordWebhook: "http://127.0.0.1",
		authToken:      "secret-token",
		client:         http.DefaultClient,
	}

	req := httptest.NewRequest(http.MethodPost, "/?auth_token=wrong", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()

	s.handle(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}
