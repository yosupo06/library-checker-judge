package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

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
		wrappedGrpc := grpcweb.WrapServer(s)
		http.ListenAndServe(":"+port, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			resp.Header().Set("Access-Control-Allow-Origin", "*")
			resp.Header().Set("Access-Control-Allow-Headers", "Content-Type, x-user-agent, x-grpc-web, authorization")
			if wrappedGrpc.IsGrpcWebRequest(req) {
				wrappedGrpc.ServeHTTP(resp, req)
			}
		}))
	} else {
		log.Print("launch gRPC server port=", port)
		listen, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatal(err)
		}
		s.Serve(listen)
	}
}
