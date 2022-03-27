package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"google.golang.org/grpc"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	_ "github.com/lib/pq"
	"gorm.io/gorm"
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func toProtoSubmission(submission *Submission) (*pb.SubmissionOverview, error) {
	overview := &pb.SubmissionOverview{
		Id:           int32(submission.ID),
		ProblemName:  submission.Problem.Name,
		ProblemTitle: submission.Problem.Title,
		UserName:     submission.User.Name,
		Lang:         submission.Lang,
		IsLatest:     submission.Testhash == submission.Problem.Testhash,
		Status:       submission.Status,
		Hacked:       submission.Hacked,
		Time:         float64(submission.MaxTime) / 1000.0,
		Memory:       int64(submission.MaxMemory),
	}
	return overview, nil
}

type server struct {
	pb.UnimplementedLibraryCheckerServiceServer
	db *gorm.DB
}

var langs = []*pb.Lang{}

func init() {
	var tomlData struct {
		Langs []struct {
			ID      string `toml:"id"`
			Name    string `toml:"name"`
			Version string `toml:"version"`
		}
	}
	if _, err := toml.DecodeFile("./langs.toml", &tomlData); err != nil {
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
	db := dbConnect(
		getEnv("POSTGRE_HOST", "127.0.0.1"),
		getEnv("POSTGRE_PORT", "5432"),
		"librarychecker",
		getEnv("POSTGRE_USER", "postgres"),
		getEnv("POSTGRE_PASS", "passwd"),
		getEnv("API_DB_LOG", "") != "")

	// launch gRPC server
	port := getEnv("PORT", "50051")
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authnFunc)))
	pb.RegisterLibraryCheckerServiceServer(s, &server{
		db: db,
	})

	if getEnv("MODE", "") == "gRPCWeb" {
		log.Print("gRPC-Web Mode port=", port)
		wrappedGrpc := grpcweb.WrapServer(s)
		http.ListenAndServe(":"+port, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			resp.Header().Set("Access-Control-Allow-Origin", "*")
			resp.Header().Set("Access-Control-Allow-Headers", "Content-Type, x-user-agent, x-grpc-web, authorization")
			if wrappedGrpc.IsGrpcWebRequest(req) {
				wrappedGrpc.ServeHTTP(resp, req)
			}
		}))
	} else {
		log.Print("gRPC Mode port=", port)
		listen, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatal(err)
		}
		s.Serve(listen)
	}
}
