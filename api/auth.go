package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/golang-jwt/jwt"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	_ "github.com/lib/pq"
)

var hmacSecret []byte

func init() {
	s := os.Getenv("HMAC_SECRET")
	if s == "" {
		log.Print("Should set HMAC_SECRET")
		s = "dummy_secret"
	}
	hmacSecret = []byte(s)
}

// UserNameKey is context key of UserName
type UserNameKey struct{}

func issueToken(user User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user.Name,
	})
	tokenString, err := token.SignedString(hmacSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func getCurrentUserName(ctx context.Context) string {
	u := ctx.Value(UserNameKey{})
	if userName, ok := u.(string); ok {
		return userName
	}
	return ""
}

func authnFunc(ctx context.Context) (context.Context, error) {
	tokenStr, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		// don't login
		return ctx, nil
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return hmacSecret, nil
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
