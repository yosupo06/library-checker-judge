package main

import (
	"net"
	"time"
	"context"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
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

func (s *server) LangList(ctx context.Context, in *pb.LangListRequest) (*pb.LangListResponse, error) {
	return &pb.LangListResponse{Langs: langs}, nil
}

// LocalConnection return local connection of gRPC
func LocalConnection() *grpc.ClientConn {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	pb.RegisterLibraryCheckerServiceServer(s, &server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatal("Server exited with error: ", err)
		}
	}()
	bufDialer := func (string, time.Duration) (net.Conn, error) {
		return lis.Dial()	
	}
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithBlock(), grpc.WithInsecure(), grpc.WithDialer(bufDialer))
	if err != nil {
		log.Fatal("Grpc dial failed: ", err)
	}
	return conn
}

func main() {
	// launch gRPC server
	port := getEnv("PORT", "50051")
	listen, err := net.Listen("tcp", ":" + port)
	if err != nil {
		log.Fatal(err)
	}
	LoadLangsToml("../compiler/langs.toml")
	s := grpc.NewServer()
	pb.RegisterLibraryCheckerServiceServer(s, &server{})
	s.Serve(listen)
}
