package main

import (
	"context"
	"net/http"
)

type requestContextKey struct{}

func withHTTPRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, requestContextKey{}, r)
}

func httpRequestFromContext(ctx context.Context) (*http.Request, bool) {
	if ctx == nil {
		return nil, false
	}
	r, _ := ctx.Value(requestContextKey{}).(*http.Request)
	return r, r != nil
}
