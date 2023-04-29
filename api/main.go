package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"github.com/yosupo06/library-checker-judge/database"
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

type server struct {
	pb.UnimplementedLibraryCheckerServiceServer
	db               *gorm.DB
	langs            []*pb.Lang
	authTokenManager AuthTokenManager
}

func NewGRPCServer(db *gorm.DB, authTokenManager AuthTokenManager, langsTomlPath string) *grpc.Server {
	// launch gRPC server
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authTokenManager.authnFunc)))

	pb.RegisterLibraryCheckerServiceServer(s, &server{
		db:               db,
		langs:            ReadLangs(langsTomlPath),
		authTokenManager: authTokenManager,
	})

	return s
}

func main() {
	langsTomlPath := flag.String("langs", "../langs/langs.toml", "toml path of langs.toml")

	pgHost := flag.String("pghost", "127.0.0.1", "postgre host")
	pgUser := flag.String("pguser", "postgres", "postgre user")
	pgPass := flag.String("pgpass", "passwd", "postgre password")
	pgTable := flag.String("pgtable", "librarychecker", "postgre table name")

	hmacKey := flag.String("hmackey", "", "hmac key")

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
	authTokenManager := NewAuthTokenManager(*hmacKey)
	s := NewGRPCServer(db, authTokenManager, *langsTomlPath)

	log.Print("launch gRPCWeb server port=", port)
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
