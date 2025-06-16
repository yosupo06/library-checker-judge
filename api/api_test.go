package main

import (
	"context"
	"log"
	"math"
	"net"
	"testing"

	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"github.com/yosupo06/library-checker-judge/database"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
)

func TestRegister(t *testing.T) {
	const token = "token"
	const name = "name"

	client := createTestAPIClientWithSetup(t, func(db *gorm.DB, authClient *DummyAuthClient) {
		authClient.registerUID(token, "uid")
	})

	if _, err := client.Register(contextWithToken(context.Background(), token), &pb.RegisterRequest{
		Name: name,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestChangeCurrentUserInfo(t *testing.T) {
	const token = "token"
	const name = "name"
	const libraryURL = "https://library.yosupo.jp"

	client := createTestAPIClientWithSetup(t, func(db *gorm.DB, authClient *DummyAuthClient) {
		authClient.registerUID(token, "uid")
	})

	if _, err := client.Register(contextWithToken(context.Background(), token), &pb.RegisterRequest{
		Name: name,
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := client.ChangeCurrentUserInfo(contextWithToken(context.Background(), token), &pb.ChangeCurrentUserInfoRequest{
		User: &pb.User{
			Name:       "name",
			LibraryUrl: libraryURL,
		},
	}); err != nil {
		t.Fatal(err)
	}

	resp, err := client.CurrentUserInfo(contextWithToken(context.Background(), token), &pb.CurrentUserInfoRequest{})
	if err != nil {
		t.Fatal(err)
	}

	if resp.User.LibraryUrl != libraryURL {
		t.Fatal("library URL is differ:", resp.User.LibraryUrl)
	}
}

func TestProblemInfo(t *testing.T) {
	client := createTestAPIClientWithSetup(t, func(db *gorm.DB, authClient *DummyAuthClient) {
		if err := database.SaveProblem(db, DUMMY_PROBLEM); err != nil {
			t.Fatal("Failed to save problem:", err)
		}
	})

	ctx := context.Background()

	problem, err := client.ProblemInfo(ctx, &pb.ProblemInfoRequest{
		Name: DUMMY_PROBLEM.Name,
	})
	if err != nil {
		t.Fatal(err)
	}
	if problem.Title != DUMMY_PROBLEM.Title {
		t.Fatal("Differ Title:", problem.Title)
	}
	if problem.SourceUrl != DUMMY_PROBLEM.SourceUrl {
		t.Fatal("Differ SourceURL:", problem.SourceUrl)
	}
	if math.Abs(problem.TimeLimit-2.0) > 0.01 {
		t.Fatal("Differ TimeLimit:", problem.TimeLimit)
	}
	if problem.TestcasesVersion != DUMMY_PROBLEM.TestCasesVersion {
		t.Fatal("Differ TestcasesVersion:", problem.TestcasesVersion)
	}
	if problem.Version != DUMMY_PROBLEM.Version {
		t.Fatal("Differ Version:", problem.Version)
	}
}

func TestNoExistProblemInfo(t *testing.T) {
	client := createTestAPIClient(t)

	ctx := context.Background()

	_, err := client.ProblemInfo(ctx, &pb.ProblemInfoRequest{
		Name: "This-problem-is-not-found",
	})
	if err == nil {
		t.Fatal("error is nil")
	}
	t.Log(err)
}

func TestLangList(t *testing.T) {
	client := createTestAPIClient(t)

	ctx := context.Background()
	list, err := client.LangList(ctx, &pb.LangListRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Langs) == 0 {
		t.Fatal(err)
	}
}

func TestChangeNoExistUserInfo(t *testing.T) {
	client := createTestAPIClient(t)

	ctx := context.Background()
	_, err := client.ChangeUserInfo(ctx, &pb.ChangeUserInfoRequest{
		User: &pb.User{
			Name: "this_is_dummy_user_name",
		},
	})
	if err == nil {
		t.Fatal("Success to change unknown user")
	}
	t.Log(err)
}

func createTestAPIClientWithSetup(t *testing.T, setUp func(db *gorm.DB, authClient *DummyAuthClient)) pb.LibraryCheckerServiceClient {
	// launch gRPC server
	listen, err := net.Listen("tcp", ":50053")
	if err != nil {
		t.Fatal(err)
	}

	// connect database
	db := database.CreateTestDB(t)

	// connect authClient
	authClient := &DummyAuthClient{
		tokenToUID: map[string]string{},
	}

	s := NewGRPCServer(db, authClient)
	go func() {
		if err := s.Serve(listen); err != nil {
			log.Fatal("Server exited: ", err)
		}
	}()

	options := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.NewClient(
		"localhost:50053",
		options...,
	)
	if err != nil {
		t.Fatal(err)
	}

	setUp(db, authClient)

	t.Cleanup(func() {
		_ = conn.Close()
		s.Stop()
	})

	return pb.NewLibraryCheckerServiceClient(conn)
}

func createTestAPIClient(t *testing.T) pb.LibraryCheckerServiceClient {
	return createTestAPIClientWithSetup(t, func(db *gorm.DB, authClient *DummyAuthClient) {})
}

type DummyAuthClient struct {
	tokenToUID map[string]string
}

func (c *DummyAuthClient) parseUID(ctx context.Context, token string) string {
	return c.tokenToUID[token]
}

func (c *DummyAuthClient) registerUID(token string, uid string) {
	c.tokenToUID[token] = uid
}

func contextWithToken(ctx context.Context, token string) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "bearer "+token))
}

var DUMMY_PROBLEM = database.Problem{
	Name:             "aplusb",
	Title:            "A + B",
	Timelimit:        2000,
	TestCasesVersion: "dummy-testcase-version",
	Version:          "dummy-version",
	SourceUrl:        "https://github.com/yosupo06/library-checker-problems/tree/master/sample/aplusb",
}
