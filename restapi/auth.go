package main

import (
	"context"
	"errors"
	"net/http"
	"strings"

	fbAuth "firebase.google.com/go/v4/auth"
)

// AuthClient provides minimal interface for verifying ID tokens.
type AuthClient interface {
	parseUID(ctx context.Context, token string) string
}

// FirebaseAuthClient implements AuthClient backed by Firebase Auth Admin SDK.
type FirebaseAuthClient struct {
	client *fbAuth.Client
}

func (c *FirebaseAuthClient) parseUID(ctx context.Context, token string) string {
	if c == nil || c.client == nil || token == "" {
		return ""
	}
	idToken, err := c.client.VerifyIDToken(ctx, token)
	if err != nil {
		return ""
	}
	return idToken.UID
}

func parseBearerToken(r *http.Request) string {
	authz := r.Header.Get("Authorization")
	if authz == "" {
		return ""
	}
	parts := strings.SplitN(authz, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

func (s *server) uidFromRequest(r *http.Request) (string, error) {
	token := parseBearerToken(r)
	if token == "" {
		return "", errors.New("no bearer token")
	}
	if s.authClient == nil {
		return "", errors.New("auth client not configured")
	}
	uid := s.authClient.parseUID(r.Context(), token)
	if uid == "" {
		return "", errors.New("invalid token")
	}
	return uid, nil
}

func (s *server) uidFromContext(ctx context.Context) (string, error) {
	r, ok := httpRequestFromContext(ctx)
	if !ok {
		return "", errors.New("request not found in context")
	}
	return s.uidFromRequest(r)
}
