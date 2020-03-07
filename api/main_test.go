package main

import (
	"context"
	"log"
	"math"
	"os"
	"strings"
	"testing"

	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"google.golang.org/grpc"
)

var client pb.LibraryCheckerServiceClient

type tokenKey struct{}

func TestProblemInfo(t *testing.T) {
	ctx := context.Background()
	problem, err := client.ProblemInfo(ctx, &pb.ProblemInfoRequest{
		Name: "aplusb",
	})
	if err != nil {
		t.Fatal(err)
	}
	if problem.Title != "A + B" {
		t.Fatal("Differ Title : ", problem.Title)
	}
	if math.Abs(problem.TimeLimit-2.0) > 0.01 {
		t.Fatal("Differ TimeLimit : ", problem.TimeLimit)
	}
}

func TestSubmissionSortOrderList(t *testing.T) {
	ctx := context.Background()
	for _, order := range []string{"", "-id", "+time"} {
		_, err := client.SubmissionList(ctx, &pb.SubmissionListRequest{
			Skip:  0,
			Limit: 100,
			Order: order,
		})
		if err != nil {
			t.Fatal("Failed SubmissionList Order: ", order)
		}
	}
	_, err := client.SubmissionList(ctx, &pb.SubmissionListRequest{
		Skip:  0,
		Limit: 100,
		Order: "dummy",
	})
	if err == nil {
		t.Fatal("Success SubmissionList Dummy Order")
	}
	t.Log(err)
}

func TestLangList(t *testing.T) {
	ctx := context.Background()
	list, err := client.LangList(ctx, &pb.LangListRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Langs) == 0 {
		t.Fatal(err)
	}
}

func TestSubmitBig(t *testing.T) {
	ctx := context.Background()
	bigSrc := strings.Repeat("a", 3*1000*1000) // 3 MB
	_, err := client.Submit(ctx, &pb.SubmitRequest{
		Problem: "aplusb",
		Source:  bigSrc,
		Lang:    "cpp",
	})
	if err == nil {
		t.Fatal("Success to submit big source")
	}
	t.Log(err)
}

func TestAnonymousRejudge(t *testing.T) {
	ctx := context.Background()
	src := strings.Repeat("a", 1000)
	resp, err := client.Submit(ctx, &pb.SubmitRequest{
		Problem: "aplusb",
		Source:  src,
		Lang:    "cpp",
	})
	if err != nil {
		t.Fatal("Unsuccess to submit source")
	}
	_, err = client.Rejudge(ctx, &pb.RejudgeRequest{
		Id: resp.Id,
	})
	if err == nil {
		t.Fatal("Success to rejudge")
	}
}

func TestAdmin(t *testing.T) {
	ctx := context.Background()
	loginResp, err := client.Login(ctx, &pb.LoginRequest{
		Name:     "admin",
		Password: "password",
	})
	if err != nil {
		t.Fatal("Failed to login")
	}
	ctx = context.WithValue(ctx, tokenKey{}, loginResp.Token)
	resp, err := client.UserInfo(ctx, &pb.UserInfoRequest{})
	if err != nil {
		t.Fatal("Failed isAdmin")
	}
	if !resp.IsAdmin {
		t.Fatal("isAdmin(admin) = False")
	}
}

func TestNotAdmin(t *testing.T) {
	ctx := context.Background()
	loginResp, err := client.Login(ctx, &pb.LoginRequest{
		Name:     "tester",
		Password: "password",
	})
	if err != nil {
		t.Fatal("Failed to login")
	}
	ctx = context.WithValue(ctx, tokenKey{}, loginResp.Token)
	resp, err := client.UserInfo(ctx, &pb.UserInfoRequest{})
	if err != nil {
		t.Fatal("Failed isAdmin")
	}
	if resp.IsAdmin {
		t.Fatal("isAdmin(tester) = True")
	}
}

type loginCreds struct{}

func (c *loginCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	dict := map[string]string{}
	if token, ok := ctx.Value(tokenKey{}).(string); ok && token != "" {
		dict["authorization"] = "bearer " + token
	}
	return dict, nil
}

func (c *loginCreds) RequireTransportSecurity() bool {
	return false
}

func TestMain(m *testing.M) {
	options := []grpc.DialOption{grpc.WithBlock(), grpc.WithPerRPCCredentials(&loginCreds{}), grpc.WithInsecure()}
	conn, err := grpc.Dial("localhost:50051", options...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client = pb.NewLibraryCheckerServiceClient(conn)

	os.Exit(m.Run())
}
