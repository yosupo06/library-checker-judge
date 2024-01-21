package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"github.com/yosupo06/library-checker-judge/database"
	"google.golang.org/grpc"

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

type server struct {
	pb.UnimplementedLibraryCheckerServiceServer
	db         *gorm.DB
	authClient AuthClient
	langs      []*pb.Lang
}

func NewGRPCServer(db *gorm.DB, authClient AuthClient, langsTomlPath string) *grpc.Server {
	// launch gRPC server
	s := grpc.NewServer()

	pb.RegisterLibraryCheckerServiceServer(s, &server{
		db:         db,
		authClient: authClient,
		langs:      ReadLangs(langsTomlPath),
	})

	return s
}

func createFirebaseApp(ctx context.Context, projectID string) (*firebase.App, error) {
	return firebase.NewApp(ctx, &firebase.Config{
		ProjectID: projectID,
	})
}

func main() {
	ctx := context.Background()

	langsTomlPath := flag.String("langs", "../langs/langs.toml", "toml path of langs.toml")
	flag.Parse()

	db := database.Connect(database.GetDSNFromEnv(), getEnv("API_DB_LOG", "") != "")

	// connect firebase
	firebaseProject := os.Getenv("FIREBASE_PROJECT")
	if firebaseProject == "" {
		log.Fatalln("Must be specify FIREBASE_PROJECT")
	}
	firebaseApp, err := createFirebaseApp(ctx, firebaseProject)
	if err != nil {
		log.Fatalln(err)
	}
	firebaseAuth, err := connectFirebaseAuth(ctx, firebaseApp)
	if err != nil {
		log.Fatalln(err)
	}

	// launch api service
	port := getEnv("PORT", "12380")
	log.Println("launch gRPCWeb server port:", port)
	s := NewGRPCServer(db, firebaseAuth, *langsTomlPath)
	wrappedGrpc := grpcweb.WrapServer(s, grpcweb.WithOriginFunc(func(origin string) bool { return true }))
	http.HandleFunc("/health", func(resp http.ResponseWriter, req *http.Request) {
		io.WriteString(resp, "SERVING")
	})
	http.ListenAndServe(":"+port, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if wrappedGrpc.IsAcceptableGrpcCorsRequest(req) || wrappedGrpc.IsGrpcWebRequest(req) {
			wrappedGrpc.ServeHTTP(resp, req)
			return
		}
		http.DefaultServeMux.ServeHTTP(resp, req)
	}))
}
