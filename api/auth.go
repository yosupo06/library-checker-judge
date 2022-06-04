package main

import (
	"context"
	"fmt"
	"log"

	"github.com/golang-jwt/jwt"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	_ "github.com/lib/pq"
)

type AuthTokenManager struct {
	hmacKey []byte
}

func NewAuthTokenManager(hmacKey string) AuthTokenManager {
	if hmacKey == "" {
		log.Fatal("HMAC key is empty")
	}
	return AuthTokenManager{
		hmacKey: []byte(hmacKey),
	}
}

func (a *AuthTokenManager) IssueToken(user User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user.Name,
	})
	tokenString, err := token.SignedString(a.hmacKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (a *AuthTokenManager) authnFunc(ctx context.Context) (context.Context, error) {
	tokenStr, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		// don't login
		return ctx, nil
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.hmacKey, nil
	})

	if err != nil {
		return ctx, nil
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return ctx, nil
	}

	if val, ok := claims["user"]; ok {
		if name, ok := val.(string); ok {
			ctx = context.WithValue(ctx, UserNameKey{}, name)
		}
	}
	return ctx, nil
}

type UserNameKey struct{}

func getCurrentUserName(ctx context.Context) string {
	u := ctx.Value(UserNameKey{})
	if userName, ok := u.(string); ok {
		return userName
	}
	return ""
}
