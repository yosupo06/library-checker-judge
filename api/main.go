package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

type server struct {
	pb.UnimplementedLibraryCheckerServiceServer
}

var langs = []*pb.Lang{}

func init() {
	// langs init
	var tomlData struct {
		Langs []struct {
			ID      string `toml:"id"`
			Name    string `toml:"name"`
			Version string `toml:"version"`
		}
	}
	if _, err := toml.DecodeFile("compiler/langs.toml", &tomlData); err != nil {
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

func (s *server) LangList(ctx context.Context, in *pb.LangListRequest) (*pb.LangListResponse, error) {
	return &pb.LangListResponse{Langs: langs}, nil
}

func main() {
	port := getEnv("PORT", "50051")
	s := grpc.NewServer()
	pb.RegisterLibraryCheckerServiceServer(s, &server{})
	reflection.Register(s)
	wrapped := grpcweb.WrapServer(s)
	http.ListenAndServe(":" + port, wrapped)
}
