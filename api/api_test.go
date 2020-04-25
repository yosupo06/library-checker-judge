package main

import (
	"context"
	"log"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
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
		t.Fatal("Failed UserInfo")
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
		t.Fatal("Failed UserInfo")
	}
	if resp.IsAdmin {
		t.Fatal("isAdmin(tester) = True")
	}
}

func TestUserList(t *testing.T) {
	ctx := context.Background()
	loginResp, err := client.Login(ctx, &pb.LoginRequest{
		Name:     "admin",
		Password: "password",
	})
	if err != nil {
		t.Fatal("Failed to login")
	}
	ctx = context.WithValue(ctx, tokenKey{}, loginResp.Token)
	_, err = client.UserList(ctx, &pb.UserListRequest{})
	if err != nil {
		t.Fatal("Failed UserList")
	}
}

func TestNotAdminUserList(t *testing.T) {
	ctx := context.Background()
	loginResp, err := client.Login(ctx, &pb.LoginRequest{
		Name:     "tester",
		Password: "password",
	})
	if err != nil {
		t.Fatal("Failed to login")
	}
	ctx = context.WithValue(ctx, tokenKey{}, loginResp.Token)
	_, err = client.UserList(ctx, &pb.UserListRequest{})
	if err == nil {
		t.Fatal("Success UserList with tester")
	}
	t.Log(err)
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	resp, err := client.Register(ctx, &pb.RegisterRequest{
		Name:     uuid.New().String(),
		Password: "password",
	})
	if err != nil {
		t.Fatal("Failed to Register")
	}
	ctx = context.WithValue(ctx, tokenKey{}, resp.Token)
}

func TestChangeUserInfo(t *testing.T) {
	// admin add bob
	ctx := context.Background()

	loginResp, err := client.Login(ctx, &pb.LoginRequest{
		Name:     "admin",
		Password: "password",
	})
	aliceCtx := context.WithValue(ctx, tokenKey{}, loginResp.Token)
	if err != nil {
		t.Fatal("Failed to Login")
	}

	bobName := uuid.New().String()
	regResp, err := client.Register(ctx, &pb.RegisterRequest{
		Name:     bobName,
		Password: "password",
	})
	bobCtx := context.WithValue(ctx, tokenKey{}, regResp.Token)
	if err != nil {
		t.Fatal("Failed to Register")
	}

	_, err = client.ChangeUserInfo(aliceCtx, &pb.ChangeUserInfoRequest{
		User: &pb.User{
			Name:    bobName,
			IsAdmin: true,
		},
	})
	if err != nil {
		t.Fatal("Failed to add Admin:", err)
	}
	resp, err := client.UserInfo(bobCtx, &pb.UserInfoRequest{})
	if err != nil {
		t.Fatal("Failed UserInfo")
	}
	if !resp.IsAdmin {
		t.Fatal("Not promote to admin")
	}

	_, err = client.ChangeUserInfo(aliceCtx, &pb.ChangeUserInfoRequest{
		User: &pb.User{
			Name:    bobName,
			IsAdmin: false,
		},
	})
	if err != nil {
		t.Fatal("Failed to add Admin:", err)
	}
	resp, err = client.UserInfo(bobCtx, &pb.UserInfoRequest{})
	if err != nil {
		t.Fatal("Failed UserInfo")
	}
	if resp.IsAdmin {
		t.Fatal("Cannot remove admin")
	}

	_, err = client.ChangeUserInfo(aliceCtx, &pb.ChangeUserInfoRequest{
		User: &pb.User{
			Name:    "admin",
			IsAdmin: false,
		},
	})
	if err == nil {
		t.Fatal("Success to remove myself")
	}
	t.Log(err)
}

func TestAddAdminByNotAdmin(t *testing.T) {
	// admin add bob
	ctx := context.Background()

	loginResp, err := client.Login(ctx, &pb.LoginRequest{
		Name:     "tester",
		Password: "password",
	})
	aliceCtx := context.WithValue(ctx, tokenKey{}, loginResp.Token)
	if err != nil {
		t.Fatal("Failed to Login")
	}

	bobName := uuid.New().String()
	regResp, err := client.Register(ctx, &pb.RegisterRequest{
		Name:     bobName,
		Password: "password",
	})
	bobCtx := context.WithValue(ctx, tokenKey{}, regResp.Token)
	if err != nil {
		t.Fatal("Failed to Register")
	}

	_, err = client.ChangeUserInfo(aliceCtx, &pb.ChangeUserInfoRequest{
		User: &pb.User{
			Name:    bobName,
			IsAdmin: true,
		},
	})
	if err == nil {
		t.Fatal("Success to add Admin")
	}
	t.Log(err)
	resp, err := client.UserInfo(bobCtx, &pb.UserInfoRequest{})
	if err != nil {
		t.Fatal("Failed UserInfo")
	}
	if resp.IsAdmin {
		t.Fatal("Promote to admin")
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
