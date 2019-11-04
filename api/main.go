package main

import (
	"database/sql"
	"context"
	"errors"
	"log"
	"net"
	"os"

	"github.com/BurntSushi/toml"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"golang.org/x/crypto/bcrypt"
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

type server struct {
	pb.UnimplementedLibraryCheckerServiceServer
}

var db *gorm.DB

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

func (s *server) ProblemInfo(ctx context.Context, in *pb.ProblemInfoRequest) (*pb.ProblemInfoResponse, error) {
	name := in.Name
	if name == "" {
		return nil, errors.New("Empty problem name")
	}
	var problem Problem
	if err := db.Select("name, title, statement, timelimit").Where("name = ?", name).First(&problem).Error; err != nil {
		return nil, errors.New("Failed to get problem")
	}
	return &pb.ProblemInfoResponse{
		Title:     problem.Title,
		Statement: problem.Statement,
		TimeLimit: problem.Timelimit / 1000.0,
	}, nil
}

func (s *server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	var user User
	if err := db.Where("name = ?", in.Name).First(&user).Error; err != nil {
		return nil, errors.New("Invalid username")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Passhash), []byte(in.Password)); err != nil {
		return nil, errors.New("Invalid password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": in.Name,
	})
	tokenString, err := token.SignedString(hmacSecret)
	if err != nil {
		return nil, err
	}
	return &pb.LoginResponse{
		Token: tokenString,
	}, nil
}

func (s *server) Submit(ctx context.Context, in *pb.SubmitRequest) (*pb.SubmitResponse, error) {
	if in.Source == "" {
		return nil, errors.New("Empty Source")
	}
	ok := false
	for _, lang := range langs {
		if lang.Id == in.Lang {
			ok = true
			break
		}
	}
	if !ok {
		return nil, errors.New("Unknown Lang")
	}
	if _, err := s.ProblemInfo(ctx, &pb.ProblemInfoRequest{
		Name: in.Problem,
	}); err != nil {
		return nil, errors.New("Unknown problem")
	}
	user := getUserName(ctx)
	submission := Submission{
		ProblemName: in.Problem,
		Lang:        in.Lang,
		Status:      "WJ",
		Source:      in.Source,
		MaxTime:     -1,
		MaxMemory:   -1,
		UserName:    sql.NullString{String: user, Valid: user != ""},
	}
	
	if err := db.Create(&submission).Error; err != nil {
		return nil, errors.New("Submit failed")
	}

	task := Task{}
	task.Submission = submission.ID
	if err := db.Create(&task).Error; err != nil {
		return nil, errors.New("Inserting to judge queue is failed")
	}

	log.Println("Submit ", submission.ID)

	return &pb.SubmitResponse{Id: int32(submission.ID)}, nil	
}

func main() {
	// connect db
	db = dbConnect()
	// launch gRPC server
	port := getEnv("PORT", "50051")
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	LoadLangsToml("../compiler/langs.toml")
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authnFunc)))
	pb.RegisterLibraryCheckerServiceServer(s, &server{})
	s.Serve(listen)
}
