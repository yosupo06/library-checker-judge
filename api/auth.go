package main

import (
	"context"

	"firebase.google.com/go/v4/auth"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	_ "github.com/lib/pq"
	"github.com/yosupo06/library-checker-judge/database"
)

type AuthClient interface {
	parseUID(ctx context.Context, token string) string
}

type FirebaseAuthClient struct {
	client *auth.Client
}

func (c *FirebaseAuthClient) parseUID(ctx context.Context, token string) string {
	if idToken, err := c.client.VerifyIDToken(ctx, token); err != nil {
		return ""
	} else {
		return idToken.UID
	}
}

func connectFirebaseAuth(ctx context.Context) (AuthClient, error) {
	firebaseApp, err := createFirebaseApp(ctx)
	if err != nil {
		return nil, err
	}
	firebaseAuth, err := firebaseApp.Auth(ctx)
	if err != nil {
		return nil, err
	}

	return &FirebaseAuthClient{
		client: firebaseAuth,
	}, nil
}

func (s *server) currentUserUID(ctx context.Context) string {
	idTokenStr, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return ""
	}
	return s.authClient.parseUID(ctx, idTokenStr)
}

func (s *server) currentUser(ctx context.Context) *database.User {
	uid := s.currentUserUID(ctx)

	if uid == "" {
		return nil
	}

	if user, err := database.FetchUserFromUID(s.db, uid); err != nil {
		return nil
	} else {
		return user
	}
}

func (s *server) currentUserName(ctx context.Context) string {
	if user := s.currentUser(ctx); user != nil {
		return user.Name
	} else {
		return ""
	}
}
