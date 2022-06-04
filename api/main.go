package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

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

type healthHandler struct {
}

func (h *healthHandler) Check(context.Context, *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	return &health.HealthCheckResponse{
		Status: health.HealthCheckResponse_SERVING,
	}, nil
}

func (h *healthHandler) Watch(*health.HealthCheckRequest, health.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "watch is not implemented.")
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
	db    *gorm.DB
	langs []*pb.Lang
}

func NewGRPCServer(db *gorm.DB, langsTomlPath string) *grpc.Server {
	// launch gRPC server
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authnFunc)))
	pb.RegisterLibraryCheckerServiceServer(s, &server{
		db:    db,
		langs: ReadLangs(langsTomlPath),
	})
	return s
}

func main() {
	langsTomlPath := flag.String("langs", "../langs/langs.toml", "toml path of langs.toml")
	isGRPCWeb := flag.Bool("grpcweb", false, "launch gRPCWeb server")
	port := getEnv("PORT", "50051")
	portArg := flag.Int("port", -1, "port number")

	if *portArg != -1 {
		port = strconv.Itoa(*portArg)
	}

	flag.Parse()

	// connect db
	db := dbConnect(
		getEnv("POSTGRE_HOST", "127.0.0.1"),
		getEnv("POSTGRE_PORT", "5432"),
		"librarychecker",
		getEnv("POSTGRE_USER", "postgres"),
		getEnv("POSTGRE_PASS", "passwd"),
		getEnv("API_DB_LOG", "") != "")

	s := NewGRPCServer(db, *langsTomlPath)

	if *isGRPCWeb {
		log.Print("launch gRPCWeb server port=", port)
		wrappedGrpc := grpcweb.WrapServer(s, grpcweb.WithOriginFunc(func(origin string) bool { return true }))
		http.HandleFunc("/health", func(resp http.ResponseWriter, req *http.Request) {
			io.WriteString(resp, "SERVING")
		})
		http.ListenAndServe(":"+port, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			if wrappedGrpc.IsAcceptableGrpcCorsRequest(req) || wrappedGrpc.IsGrpcWebRequest(req) {
				// TODO: remove these headers
				resp.Header().Set("Access-Control-Allow-Origin", "*")
				resp.Header().Set("Access-Control-Allow-Headers", "Content-Type, x-user-agent, x-grpc-web, authorization")
				wrappedGrpc.ServeHTTP(resp, req)
				return
			}
			http.DefaultServeMux.ServeHTTP(resp, req)
		}))
	} else {
		log.Print("launch gRPC server port=", port)
		health.RegisterHealthServer(s, &healthHandler{})
		listen, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatal(err)
		}
		s.Serve(listen)
	}
}
