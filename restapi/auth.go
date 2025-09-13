package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"

	firebase "firebase.google.com/go/v4"
	fbAuth "firebase.google.com/go/v4/auth"
)

var (
	fbOnce sync.Once
	fbCli  *fbAuth.Client
	fbErr  error
)

func getFirebaseAuth() (*fbAuth.Client, error) {
	fbOnce.Do(func() {
		// Resolve project ID with fallbacks suitable for emulator
		project := os.Getenv("FIREBASE_PROJECT")
		if project == "" {
			project = os.Getenv("GOOGLE_CLOUD_PROJECT")
		}
		if project == "" {
			project = os.Getenv("GCLOUD_PROJECT")
		}
		if project == "" && os.Getenv("FIREBASE_AUTH_EMULATOR_HOST") != "" {
			// Emulator in use; default a project ID if none provided
			project = "dev-library-checker-project"
		}

		var cfg *firebase.Config
		if project != "" {
			cfg = &firebase.Config{ProjectID: project}
		} else {
			// As a last resort, allow nil config (admin SDK may infer)
			cfg = nil
		}

		app, err := firebase.NewApp(context.Background(), cfg)
		if err != nil {
			fbErr = err
			return
		}
		fbCli, fbErr = app.Auth(context.Background())
	})
	return fbCli, fbErr
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

func parseUIDFromRequest(r *http.Request) (string, error) {
	token := parseBearerToken(r)
	if token == "" {
		return "", errors.New("no bearer token")
	}
	cli, err := getFirebaseAuth()
	if err != nil {
		return "", err
	}
	idToken, err := cli.VerifyIDToken(r.Context(), token)
	if err != nil {
		return "", err
	}
	return idToken.UID, nil
}
