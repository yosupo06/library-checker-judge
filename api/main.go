package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

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

func createFirebaseApp(ctx context.Context) (*firebase.App, error) {
	return firebase.NewApp(ctx, &firebase.Config{
		ProjectID: "library-checker-project",
	})
}

func main() {
	ctx := context.Background()

	langsTomlPath := flag.String("langs", "../langs/langs.toml", "toml path of langs.toml")

	pgHost := flag.String("pghost", "127.0.0.1", "postgre host")
	pgUser := flag.String("pguser", "postgres", "postgre user")
	pgPass := flag.String("pgpass", "passwd", "postgre password")
	pgTable := flag.String("pgtable", "librarychecker", "postgre table name")

	portArg := flag.Int("port", -1, "port number")
	flag.Parse()

	port := getEnv("PORT", "12380")
	if *portArg != -1 {
		port = strconv.Itoa(*portArg)
	}

	// connect db
	db := database.Connect(
		*pgHost,
		getEnv("POSTGRE_PORT", "5432"),
		*pgTable,
		*pgUser,
		*pgPass,
		getEnv("API_DB_LOG", "") != "")

	// connect firebase
	firebaseAuth, err := connectFirebaseAuth(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	s := NewGRPCServer(db, firebaseAuth, *langsTomlPath)

	log.Println("launch gRPCWeb server port:", port)
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
