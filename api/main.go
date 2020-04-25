package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/BurntSushi/toml"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"google.golang.org/grpc"

	"github.com/dgrijalva/jwt-go"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func canRejudge(ctx context.Context, submission *pb.SubmissionOverview) bool {
	userName := getUserName(ctx)
	if userName == "" {
		return false
	}
	if userName == submission.UserName {
		return true
	}
	if !submission.IsLatest && submission.Status == "AC" {
		return true
	}
	if isAdmin(ctx) {
		return true
	}
	return false
}

type server struct {
	pb.UnimplementedLibraryCheckerServiceServer
}

var db *gorm.DB

func issueToken(name string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": name,
	})
	tokenString, err := token.SignedString(hmacSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

var langs = []*pb.Lang{}

// LoadLangsToml load langs.toml
func LoadLangsToml(tomlPath string) {
	var tomlData struct {
		Langs []struct {
			ID      string `toml:"id"`
			Name    string `toml:"name"`
			Version string `toml:"version"`
		}
	}
	if _, err := toml.DecodeFile(tomlPath, &tomlData); err != nil {
		log.Fatal(err)
	}
	for _, lang := range tomlData.Langs {
		if lang.ID == "checker" {
			continue
		}
		langs = append(langs, &pb.Lang{
			Id:      lang.ID,
			Name:    lang.Name,
			Version: lang.Version,
		})
	}
}

func main() {
	// connect db
	db = dbConnect()
	defer db.Close()
	if getEnv("API_DB_LOG", "") != "" {
		db.LogMode(true)
	}

	// launch gRPC server
	port := getEnv("PORT", "50051")
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	LoadLangsToml("./langs.toml")
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authnFunc)))
	pb.RegisterLibraryCheckerServiceServer(s, &server{})
	s.Serve(listen)
}
