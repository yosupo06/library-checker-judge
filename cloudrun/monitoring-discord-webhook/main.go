package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const maxRequestBodyBytes = 1 << 20

type server struct {
	discordWebhook string
	authToken      string
	client         *http.Client
}

type discordWebhookMessage struct {
	Content string         `json:"content,omitempty"`
	Embeds  []discordEmbed `json:"embeds,omitempty"`
}

type discordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	URL         string              `json:"url,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Fields      []discordEmbedField `json:"fields,omitempty"`
}

type discordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

func main() {
	s := &server{
		discordWebhook: os.Getenv("DISCORD_WEBHOOK"),
		authToken:      os.Getenv("WEBHOOK_AUTH_TOKEN"),
		client:         &http.Client{Timeout: 10 * time.Second},
	}
	if s.discordWebhook == "" {
		log.Fatal("DISCORD_WEBHOOK is required")
	}
	if s.authToken == "" {
		log.Fatal("WEBHOOK_AUTH_TOKEN is required")
	}

	http.HandleFunc("/", s.handle)

	port := envOrDefault("PORT", "8080")
	log.Printf("monitoring-discord-webhook listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

func (s *server) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes))
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "decode json", http.StatusBadRequest)
		return
	}

	msg := buildDiscordMessage(payload)
	if err := s.postDiscord(r.Context(), msg); err != nil {
		log.Printf("post discord: %v", err)
		http.Error(w, "post discord", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *server) authorized(r *http.Request) bool {
	if s.authToken == "" {
		return false
	}
	if subtleCompare(r.URL.Query().Get("auth_token"), s.authToken) {
		return true
	}
	if subtleCompare(r.URL.Query().Get("token"), s.authToken) {
		return true
	}
	const bearerPrefix = "Bearer "
	auth := r.Header.Get("Authorization")
	return strings.HasPrefix(auth, bearerPrefix) && subtleCompare(strings.TrimPrefix(auth, bearerPrefix), s.authToken)
}

func (s *server) postDiscord(ctx context.Context, msg discordWebhookMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal discord payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.discordWebhook, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("discord response close: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		preview, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("discord status %d: %s", resp.StatusCode, string(preview))
	}
	return nil
}

func buildDiscordMessage(payload map[string]any) discordWebhookMessage {
	incident := objectValue(payload, "incident")
	state := firstNonEmptyString(
		stringValue(incident, "state"),
		stringValue(payload, "state"),
	)
	policy := firstNonEmptyString(
		stringValue(incident, "policy_name"),
		stringValue(payload, "policy_name"),
		"Cloud Monitoring alert",
	)
	summary := firstNonEmptyString(
		stringValue(incident, "summary"),
		stringValue(payload, "summary"),
		stringValue(incident, "documentation"),
		stringValue(payload, "documentation"),
	)
	url := firstNonEmptyString(
		stringValue(incident, "url"),
		stringValue(payload, "url"),
	)

	color := 0xff9900
	normalizedState := strings.ToUpper(state)
	switch normalizedState {
	case "OPEN":
		color = 0xff3333
	case "CLOSED":
		color = 0x2ecc71
	}

	fields := []discordEmbedField{}
	addField := func(name, value string, inline bool) {
		if value == "" {
			return
		}
		fields = append(fields, discordEmbedField{Name: name, Value: truncate(value, 1024), Inline: inline})
	}
	addField("State", state, true)
	addField("Condition", stringValue(incident, "condition_name"), true)
	addField("Resource", firstNonEmptyString(
		stringValue(incident, "resource_display_name"),
		stringValue(incident, "resource_name"),
		stringValue(incident, "resource_id"),
	), false)

	metric := objectValue(incident, "metric")
	addField("Metric", firstNonEmptyString(
		stringValue(metric, "displayName"),
		stringValue(metric, "type"),
	), false)

	return discordWebhookMessage{
		Embeds: []discordEmbed{
			{
				Title:       truncate(fmt.Sprintf("%s: %s", firstNonEmptyString(state, "ALERT"), policy), 256),
				Description: truncate(summary, 4096),
				URL:         url,
				Color:       color,
				Fields:      fields,
			},
		},
	}
}

func objectValue(m map[string]any, key string) map[string]any {
	v, ok := m[key]
	if !ok {
		return map[string]any{}
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return obj
}

func stringValue(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64, bool:
		return fmt.Sprint(t)
	default:
		return ""
	}
}

func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func subtleCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := range a {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
