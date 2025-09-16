package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	"github.com/go-chi/chi/v5"
	"github.com/yosupo06/library-checker-judge/database"
	restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
	"gorm.io/gorm"
)

// server is REST server implementation for OpenAPI handlers.
type server struct {
	db           *gorm.DB
	authClient   AuthClient
	updateUserFn func(*gorm.DB, database.User) error
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func createFirebaseApp(ctx context.Context, projectID string) (*firebase.App, error) {
	return firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID})
}

func (s *server) updateUser(db *gorm.DB, user database.User) error {
	if s != nil && s.updateUserFn != nil {
		return s.updateUserFn(db, user)
	}
	return database.UpdateUser(db, user)
}

func main() {
	db := database.Connect(database.GetDSNFromEnv(), getEnv("API_DB_LOG", "") != "")

	// connect firebase auth (required)
	ctx := context.Background()
	firebaseProject := os.Getenv("FIREBASE_PROJECT")
	if firebaseProject == "" {
		slog.Error("FIREBASE_PROJECT must be set")
		os.Exit(1)
	}
	app, err := createFirebaseApp(ctx, firebaseProject)
	if err != nil {
		slog.Error("create firebase app failed", "error", err)
		os.Exit(1)
	}
	authCli, err := app.Auth(ctx)
	if err != nil {
		slog.Error("connect firebase auth failed", "error", err)
		os.Exit(1)
	}
	ac := &FirebaseAuthClient{client: authCli}

	r := chi.NewRouter()
	// CORS (dev)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if req.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, req)
		})
	})

	// Register OpenAPI handlers on chi router
	_ = restapi.HandlerFromMux(&server{db: db, authClient: ac}, r)
	r.Get("/openapi.yaml", func(w http.ResponseWriter, req *http.Request) { http.ServeFile(w, req, "openapi/openapi.yaml") })
	r.Get("/health", func(w http.ResponseWriter, req *http.Request) { _, _ = w.Write([]byte("SERVING")) })

	port := getEnv("PORT", "12381")
	slog.Info("Launch REST server", "port", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		slog.Error("server exit", "error", err)
		os.Exit(1)
	}
}
