package main_test

import (
	"context"
	"log"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"google.golang.org/grpc"
)

var client pb.LibraryCheckerServiceClient
var judgeCtx context.Context

func loginAsAdmin(t *testing.T) context.Context {
	ctx := context.Background()
	loginResp, err := client.Login(ctx, &pb.LoginRequest{
		Name:     "admin",
		Password: "password",
	})
	if err != nil {
		t.Fatal("Failed to Login")
	}
	ctx = context.WithValue(ctx, tokenKey{}, loginResp.Token)
	return ctx
}

func fetchSubmission(t *testing.T, id int32) *pb.SubmissionInfoResponse {
	ctx := context.Background()
	resp, err := client.SubmissionInfo(ctx, &pb.SubmissionInfoRequest{
		Id: id,
	})
	if err != nil {
		t.Fatalf("Failed to SubmissionInfo %d %d", id, err)
	}
	return resp
}

func assertEqualCases(t *testing.T, expectCases []*pb.SubmissionCaseResult, actualCases []*pb.SubmissionCaseResult) {
	if len(actualCases) != len(expectCases) {
		t.Fatal("Error CaseResults length differ", actualCases, expectCases)
	}
	for i := 0; i < len(actualCases); i++ {
		expect := expectCases[i]
		actual := actualCases[i]
		if expect.Case != actual.Case || expect.Memory != actual.Memory || expect.Status != actual.Status || expect.Time != actual.Time {
			t.Fatal("Case differ", expect, actual)
		}
	}
}

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

func TestChangeDummyUserInfo(t *testing.T) {
	// admin add bob
	ctx := context.Background()

	loginResp, err := client.Login(ctx, &pb.LoginRequest{
		Name:     "admin",
		Password: "password",
	})
	ctx = context.WithValue(ctx, tokenKey{}, loginResp.Token)
	if err != nil {
		t.Fatal("Failed to Login")
	}

	_, err = client.ChangeUserInfo(ctx, &pb.ChangeUserInfoRequest{
		User: &pb.User{
			Name:    "this_is_dummy_user_name",
			IsAdmin: true,
		},
	})
	if err == nil {
		t.Fatal("Success to change unknown user")
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

func clearTask(t *testing.T) {
	ctx := loginAsAdmin(t)
	for {
		task, err := client.PopJudgeTask(ctx, &pb.PopJudgeTaskRequest{
			JudgeName: "judge-dummy",
		})
		if err != nil {
			t.Fatal("Failed PopJudgeTask: ", err)
		}
		if task.SubmissionId == -1 {
			break
		}
		t.Log("Pop Task: ", task.SubmissionId)
	}
	t.Log("Clean Tasks")
}

func TestOtherJudge(t *testing.T) {
	clearTask(t)

	ctx := context.Background()
	judgeCtx := loginAsAdmin(t)

	src := "this is a test source"
	submitResp, err := client.Submit(ctx, &pb.SubmitRequest{
		Problem: "aplusb",
		Source:  src,
		Lang:    "cpp",
	})
	id := submitResp.Id
	if err != nil {
		t.Fatal("Success to submit big source: ", err)
	}
	t.Log("Submit: ", id)

	resp, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
		JudgeName: "judge-test",
	})
	if err != nil {
		t.Fatal("Failed to PopJudgeTask:", err)
	}

	if id != resp.SubmissionId {
		t.Fatalf("ID is differ, %v vs %v", submitResp.Id, resp.SubmissionId)
	}

	_, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
		JudgeName:    "judge-other",
		Status:       "Judging",
		SubmissionId: id,
		IsFinished:   false,
	})
	if err == nil {
		t.Fatal("Success to SyncJudgeTaskStatus")
	}
	t.Log(err)
}

func TestJudgeSyncAfterFinished(t *testing.T) {
	clearTask(t)

	ctx := context.Background()
	judgeCtx := loginAsAdmin(t)

	src := "this is a test source"
	submitResp, err := client.Submit(ctx, &pb.SubmitRequest{
		Problem: "aplusb",
		Source:  src,
		Lang:    "cpp",
	})
	id := submitResp.Id
	if err != nil {
		t.Fatal("Success to submit big source: ", err)
	}
	t.Log("Submit: ", id)

	resp, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
		JudgeName: "judge-test",
	})
	if err != nil {
		t.Fatal("Failed to PopJudgeTask:", err)
	}

	if id != resp.SubmissionId {
		t.Fatalf("ID is differ, %v vs %v", submitResp.Id, resp.SubmissionId)
	}

	_, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
		JudgeName:    "judge-test",
		Status:       "AC",
		SubmissionId: id,
		IsFinished:   true,
	})

	if err != nil {
		t.Fatal("JudgeSync Failed:", err)
	}

	_, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
		JudgeName:    "judge-test",
		Status:       "AC",
		SubmissionId: id,
		IsFinished:   true,
	})
	if err == nil {
		t.Fatal("Success to SyncJudgeTaskStatus")
	}
	t.Log(err)
}

func TestSimulateJudge(t *testing.T) {
	clearTask(t)

	ctx := context.Background()
	judgeCtx := loginAsAdmin(t)

	src := "this is a test source"
	submitResp, err := client.Submit(ctx, &pb.SubmitRequest{
		Problem: "aplusb",
		Source:  src,
		Lang:    "cpp",
	})
	id := submitResp.Id
	if err != nil {
		t.Fatal("Success to submit big source: ", err)
	}
	t.Log("Submit: ", id)

	resp, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
		JudgeName: "judge-test",
	})
	if err != nil {
		t.Fatal("Failed to PopJudgeTask:", err)
	}

	if id != resp.SubmissionId {
		t.Fatalf("ID is differ, %v vs %v", submitResp.Id, resp.SubmissionId)
	}

	cases := []*pb.SubmissionCaseResult{
		{
			Case:   "test00",
			Status: "AC",
			Time:   1.0,
			Memory: 1,
		},
		{
			Case:   "test01",
			Status: "WA",
			Time:   1.0,
			Memory: 1,
		},
		{
			Case:   "test02",
			Status: "TLE",
			Time:   1.0,
			Memory: 1,
		},
	}
	if _, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
		JudgeName:    "judge-test",
		Status:       "Judging",
		CaseResults:  cases[0:2],
		SubmissionId: id,
		IsFinished:   false,
	}); err != nil {
		t.Fatal("Failed to SyncJudgeTaskStatus:", err)
	}

	sub := fetchSubmission(t, id)
	if sub.Overview.Status != "Judging" {
		t.Fatal("Status is not changed: ", sub.Overview.Status)
	}
	assertEqualCases(t, cases[0:2], sub.CaseResults)

	if err != nil {
		t.Fatal("Failed to PopJudgeTask:", err)
	}

	if _, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
		JudgeName:    "judge-test",
		Status:       "TLE",
		CaseResults:  cases[2:3],
		SubmissionId: id,
		IsFinished:   true,
	}); err != nil {
		t.Fatal("Failed to SyncJudgeTaskStatus:", err)
	}

	sub = fetchSubmission(t, id)
	if sub.Overview.Status != "TLE" {
		t.Fatal("Status is not changed")
	}
	assertEqualCases(t, cases[0:3], sub.CaseResults)
}

type tokenKey struct{}
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
	options := []grpc.DialOption{grpc.WithBlock(), grpc.WithPerRPCCredentials(&loginCreds{}), grpc.WithInsecure(), grpc.WithTimeout(3 * time.Second)}
	conn, err := grpc.Dial("localhost:50051", options...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client = pb.NewLibraryCheckerServiceClient(conn)

	os.Exit(m.Run())
}
