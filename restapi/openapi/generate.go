package openapi

//go:generate go tool oapi-codegen -package api -generate types,chi-server,spec -o ../internal/api/api.gen.go openapi.yaml
