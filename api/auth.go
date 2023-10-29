package main

import (
	"context"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	_ "github.com/lib/pq"
	"github.com/yosupo06/library-checker-judge/database"

	pb "github.com/yosupo06/library-checker-judge/api/proto"
)

func (s *server) currentUserUID(ctx context.Context) string {
	idTokenStr, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return ""
	}
	if idToken, err := s.firebaseAuth.VerifyIDToken(ctx, idTokenStr); err != nil {
		return ""
	} else {
		return idToken.UID
	}
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

func toProtoUser(user *database.User) *pb.User {
	if user == nil {
		return nil
	}
	return &pb.User{
		Name:        user.Name,
		LibraryUrl:  user.LibraryURL,
		IsDeveloper: user.IsDeveloper,
	}
}
