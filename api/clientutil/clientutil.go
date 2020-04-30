package clientutil

import (
	"context"
)

type tokenKey struct{}

// LoginCreds used as grpc.WithPerRPCCredentials(&loginCreds{})
type LoginCreds struct{}

func (c *LoginCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	dict := map[string]string{}
	if token, ok := ctx.Value(tokenKey{}).(string); ok && token != "" {
		dict["authorization"] = "bearer " + token
	}
	return dict, nil
}

func (c *LoginCreds) RequireTransportSecurity() bool {
	return false
}

// ContextWithToken return context with token(return value of register, login)
func ContextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey{}, token)
}
