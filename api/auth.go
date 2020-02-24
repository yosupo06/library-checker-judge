package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dgrijalva/jwt-go"
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

// UserKey is context key of username
type UserKey struct{}

func getUserName(ctx context.Context) string {
	u := ctx.Value(UserKey{})
	if name, ok := u.(string); ok {
		return name
	}
	return ""
}

func isAdmin(ctx context.Context) bool {
	u := getUserName(ctx)
	if u == "" {
		// don't login
		return false
	}
	var user User
	if err := db.Where("name = ?", u).First(&user).Error; err != nil {
		// invalid user name
		return false
	}
	return user.Admin
}

func authnFunc(ctx context.Context) (context.Context, error) {
	tokenStr, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		// don't login
		return ctx, nil
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
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
		if user, ok := val.(string); ok {
			ctx = context.WithValue(ctx, UserKey{}, user)
		}
	}
	return ctx, nil
}
