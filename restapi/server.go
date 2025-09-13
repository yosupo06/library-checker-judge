package main

import (
    "log/slog"
    "net/http"
    "os"

    "github.com/go-chi/chi/v5"
    restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
    "github.com/yosupo06/library-checker-judge/database"
    "gorm.io/gorm"
)

// server is REST server implementation for OpenAPI handlers.
type server struct{ db *gorm.DB }

func getEnv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func main() {
    db := database.Connect(database.GetDSNFromEnv(), getEnv("API_DB_LOG", "") != "")

    r := chi.NewRouter()
    // CORS (dev)
    r.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            if req.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }
            next.ServeHTTP(w, req)
        })
    })

    // Register OpenAPI handlers on chi router
    _ = restapi.HandlerFromMux(&server{db: db}, r)
    r.Get("/openapi.yaml", func(w http.ResponseWriter, req *http.Request) { http.ServeFile(w, req, "openapi/openapi.yaml") })
    r.Get("/health", func(w http.ResponseWriter, req *http.Request) { _, _ = w.Write([]byte("SERVING")) })

    port := getEnv("PORT", "12381")
    slog.Info("Launch REST server", "port", port)
    if err := http.ListenAndServe(":"+port, r); err != nil {
        slog.Error("server exit", "error", err)
        os.Exit(1)
    }
}
