package clientutil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

func ApiConnect(apiHost string, useTLS bool) *grpc.ClientConn {
	options := []grpc.DialOption{grpc.WithBlock(), grpc.WithPerRPCCredentials(&LoginCreds{}), grpc.WithTimeout(10 * time.Second)}
	if !useTLS {
		options = append(options, grpc.WithInsecure())
	} else {
		systemRoots, err := x509.SystemCertPool()
		if err != nil {
			log.Fatal(err)
		}
		creds := credentials.NewTLS(&tls.Config{
			RootCAs: systemRoots,
		})
		options = append(options, grpc.WithTransportCredentials(creds))
	}
	log.Printf("Connect to API host: %v, TLS: %v", apiHost, useTLS)
	conn, err := grpc.Dial(apiHost, options...)
	if err != nil {
		log.Fatal("Cannot connect to the API server:", err)
	}
	return conn
}
