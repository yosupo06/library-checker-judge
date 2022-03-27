package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	clientutil "github.com/yosupo06/library-checker-judge/api/clientutil"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
	"gorm.io/gorm"
)

func createTestDB(t *testing.T) *gorm.DB {
	dbName := uuid.New().String()
	t.Log("create DB: ", dbName)

	dumpCmd := exec.Command("pg_dump",
		"-h", "localhost",
		"-U", "postgres",
		"-d", "librarychecker",
		"-p", "5432",
		"-s")
	dumpCmd.Env = append(os.Environ(), "PGPASSWORD=passwd")

	sql, err := dumpCmd.Output()
	if err != nil {
		t.Fatal(err)
	}

	createCmd := exec.Command("createdb",
		"-h", "localhost",
		"-U", "postgres",
		"-p", "5432",
		dbName)
	createCmd.Env = append(os.Environ(), "PGPASSWORD=passwd")
	if err := createCmd.Run(); err != nil {
		t.Fatal("exec failed: ", err.Error())
	}

	tableCmd := exec.Command("psql",
		"-h", "localhost",
		"-U", "postgres",
		"-d", dbName,
		"-p", "5432")
	tableCmd.Env = append(os.Environ(), "PGPASSWORD=passwd")
	tableCmd.Stdin = bytes.NewReader(sql)
	if err := tableCmd.Run(); err != nil {
		t.Fatal("exec failed: ", err.Error())
	}

	db := dbConnect("localhost", "5432", dbName, "postgres", "passwd", getEnv("API_DB_LOG", "") != "")

	registerUser(db, "admin", "password", true)
	registerUser(db, "tester", "password", false)

	client, close := createAPIClient(t, db)
	defer close()

	ctx := loginAsAdmin(t, client)
	if _, err := client.ChangeProblemInfo(ctx, &pb.ChangeProblemInfoRequest{
		Name:        "aplusb",
		Title:       "A + B",
		Statement:   "Please calculate A + B",
		TimeLimit:   2.0,
		CaseVersion: "dummy-initial-version",
		SourceUrl:   "https://github.com/yosupo06/library-checker-problems/tree/master/sample/aplusb",
	}); err != nil {
		t.Fatal(err)
	}

	return db
}

func createAPIClient(t *testing.T, db *gorm.DB) (pb.LibraryCheckerServiceClient, func()) {
	// launch gRPC server
	listen, err := net.Listen("tcp", ":50053")
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authnFunc)))
	pb.RegisterLibraryCheckerServiceServer(s, &server{
		db: db,
	})
	go func() {
		if err := s.Serve(listen); err != nil {
			log.Fatal("Server exited: ", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	options := []grpc.DialOption{grpc.WithBlock(), grpc.WithPerRPCCredentials(&clientutil.LoginCreds{}), grpc.WithInsecure()}
	conn, err := grpc.DialContext(
		ctx,
		"localhost:50053",
		options...,
	)
	if err != nil {
		t.Fatal(err)
	}

	return pb.NewLibraryCheckerServiceClient(conn), func() {
		cancel()
		conn.Close()
		s.Stop()
	}
}

func loginContext(t *testing.T, name string, client pb.LibraryCheckerServiceClient) context.Context {
	ctx := context.Background()
	loginResp, err := client.Login(ctx, &pb.LoginRequest{
		Name:     name,
		Password: "password",
	})
	if err != nil {
		t.Fatal("Failed to Login")
	}
	return clientutil.ContextWithToken(ctx, loginResp.Token)
}

func loginAsAdmin(t *testing.T, client pb.LibraryCheckerServiceClient) context.Context {
	return loginContext(t, "admin", client)
}

func loginAsTester(t *testing.T, client pb.LibraryCheckerServiceClient) context.Context {
	return loginContext(t, "tester", client)
}

func submitSomething(t *testing.T, client pb.LibraryCheckerServiceClient) int32 {
	ctx := context.Background()
	src := "this is a test source"
	submitResp, err := client.Submit(ctx, &pb.SubmitRequest{
		Problem: "aplusb",
		Source:  src,
		Lang:    "cpp",
	})
	id := submitResp.Id
	if err != nil {
		t.Fatal("Failed to submit:", err)
	}
	t.Log("Submit: ", id)
	return id
}

func testFetchSubmission(t *testing.T, id int32, client pb.LibraryCheckerServiceClient) *pb.SubmissionInfoResponse {
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
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

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
	if problem.SourceUrl != "https://github.com/yosupo06/library-checker-problems/tree/master/sample/aplusb" {
		t.Fatal("Differ SourceURL : ", problem.SourceUrl)
	}
	if math.Abs(problem.TimeLimit-2.0) > 0.01 {
		t.Fatal("Differ TimeLimit : ", problem.TimeLimit)
	}
	if problem.CaseVersion == "" {
		t.Fatal("Case Version is empty")
	}
}

func TestSubmissionSortOrderList(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

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
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

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
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

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
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

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

func TestRejudgeTwice(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	judgeCtx := loginAsAdmin(t, client)
	src := strings.Repeat("a", 1000)
	resp, err := client.Submit(judgeCtx, &pb.SubmitRequest{
		Problem: "aplusb",
		Source:  src,
		Lang:    "cpp",
	})
	if err != nil {
		t.Fatal("Unsuccess to submit source")
	}

	task, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
		JudgeName:    "judge-test",
		ExpectedTime: durationpb.New(2 * time.Second),
	})
	if err != nil {
		t.Fatal(err)
	}

	id := task.SubmissionId
	if id == -1 {
		t.Fatal("Cannot fetch task")
	}

	if id != resp.Id {
		t.Fatalf("Differ ID %v vs %v", id, resp.Id)
	}

	if _, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
		JudgeName: "judge-test",
		Status:    "Judging",
		CaseResults: []*pb.SubmissionCaseResult{
			{
				Case:   "test00",
				Status: "AC",
				Time:   1.0,
				Memory: 1,
			},
		},
		SubmissionId: id,
	}); err != nil {
		t.Fatal(err)
	}

	if _, err = client.FinishJudgeTask(judgeCtx, &pb.FinishJudgeTaskRequest{
		JudgeName:    "judge-test",
		Status:       "AC",
		SubmissionId: id,
	}); err != nil {
		t.Fatal(err)
	}

	_, err = client.Rejudge(judgeCtx, &pb.RejudgeRequest{
		Id: resp.Id,
	})
	if err != nil {
		t.Fatal("Failed to rejudge")
	}
	_, err = client.Rejudge(judgeCtx, &pb.RejudgeRequest{
		Id: resp.Id,
	})
	if err == nil {
		t.Fatal("Success to rejudge")
	}
	t.Log(err)
}

func TestAdmin(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	ctx := loginAsAdmin(t, client)
	resp, err := client.UserInfo(ctx, &pb.UserInfoRequest{})
	if err != nil {
		t.Fatal("Failed UserInfo")
	}
	if !resp.User.IsAdmin {
		t.Fatal("isAdmin(admin) = False")
	}
}

func TestNotAdmin(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	ctx := loginAsTester(t, client)
	resp, err := client.UserInfo(ctx, &pb.UserInfoRequest{})
	if err != nil {
		t.Fatal("Failed UserInfo")
	}
	if resp.User.IsAdmin {
		t.Fatal("isAdmin(tester) = True")
	}
}

func TestUserList(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	ctx := loginAsAdmin(t, client)
	_, err := client.UserList(ctx, &pb.UserListRequest{})
	if err != nil {
		t.Fatal("Failed UserList")
	}
}

func TestNotAdminUserList(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	ctx := loginAsTester(t, client)
	_, err := client.UserList(ctx, &pb.UserListRequest{})
	if err == nil {
		t.Fatal("Success UserList with tester")
	}
	t.Log(err)
}

func TestCreateUser(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	ctx := context.Background()
	resp, err := client.Register(ctx, &pb.RegisterRequest{
		Name:     uuid.New().String(),
		Password: "password",
	})
	if err != nil {
		t.Fatal("Failed to Register")
	}
	ctx = clientutil.ContextWithToken(ctx, resp.Token)
	_, err = client.UserInfo(ctx, &pb.UserInfoRequest{})
	if err != nil {
		t.Fatal("Failed to UserInfo")
	}
}

func TestChangeUserInfo(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	// admin add bob
	ctx := context.Background()
	aliceCtx := loginAsAdmin(t, client)
	bobName := uuid.New().String()
	regResp, err := client.Register(ctx, &pb.RegisterRequest{
		Name:     bobName,
		Password: "password",
	})
	bobCtx := clientutil.ContextWithToken(ctx, regResp.Token)
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
	if !resp.User.IsAdmin {
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
	if resp.User.IsAdmin {
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
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	// admin add bob
	ctx := loginAsAdmin(t, client)

	_, err := client.ChangeUserInfo(ctx, &pb.ChangeUserInfoRequest{
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
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	// admin add bob
	ctx := context.Background()

	aliceCtx := loginAsTester(t, client)

	bobName := uuid.New().String()
	regResp, err := client.Register(ctx, &pb.RegisterRequest{
		Name:     bobName,
		Password: "password",
	})
	if err != nil {
		t.Fatal("Failed to Register")
	}
	bobCtx := clientutil.ContextWithToken(context.Background(), regResp.Token)

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
	if resp.User.IsAdmin {
		t.Fatal("Promote to admin")
	}
}

func TestOtherJudge(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	judgeCtx := loginAsAdmin(t, client)
	id := submitSomething(t, client)
	resp, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
		JudgeName: "judge-test",
	})
	if err != nil {
		t.Fatal("Failed to PopJudgeTask:", err)
	}

	if id != resp.SubmissionId {
		t.Fatalf("ID is differ, %v vs %v", id, resp.SubmissionId)
	}

	_, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
		JudgeName:    "judge-other",
		Status:       "Judging",
		SubmissionId: id,
	})
	if err == nil {
		t.Fatal("Success to SyncJudgeTaskStatus")
	}
	t.Log(err)
}

func TestJudgeSyncAfterFinished(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	judgeCtx := loginAsAdmin(t, client)
	id := submitSomething(t, client)

	resp, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
		JudgeName: "judge-test",
	})
	if err != nil {
		t.Fatal("Failed to PopJudgeTask:", err)
	}

	if id != resp.SubmissionId {
		t.Fatalf("ID is differ, %v vs %v", id, resp.SubmissionId)
	}

	_, err = client.FinishJudgeTask(judgeCtx, &pb.FinishJudgeTaskRequest{
		JudgeName:    "judge-test",
		Status:       "AC",
		SubmissionId: id,
	})

	if err != nil {
		t.Fatal("JudgeSync Failed:", err)
	}

	_, err = client.FinishJudgeTask(judgeCtx, &pb.FinishJudgeTaskRequest{
		JudgeName:    "judge-test",
		Status:       "AC",
		SubmissionId: id,
	})
	if err == nil {
		t.Fatal("Success to FinishJudgeTask Twice")
	}
	t.Log(err)
	_, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
		JudgeName:    "judge-test",
		Status:       "AC",
		SubmissionId: id,
	})
	if err == nil {
		t.Fatal("Success to SyncJudgeTaskStatus after finished")
	}
	t.Log(err)
}

func TestSimulateJudge(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	judgeCtx := loginAsAdmin(t, client)
	id := submitSomething(t, client)

	resp, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
		JudgeName: "judge-test",
	})
	if err != nil {
		t.Fatal("Failed to PopJudgeTask:", err)
	}

	if id != resp.SubmissionId {
		t.Fatalf("ID is differ, %v vs %v", id, resp.SubmissionId)
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
	}); err != nil {
		t.Fatal("Failed to SyncJudgeTaskStatus:", err)
	}

	sub := testFetchSubmission(t, id, client)
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
		Time:         1.234,
		Memory:       5678,
		CaseResults:  cases[2:3],
		SubmissionId: id,
	}); err != nil {
		t.Fatal("Failed to SyncJudgeTaskStatus:", err)
	}

	sub = testFetchSubmission(t, id, client)
	if sub.Overview.Status != "TLE" {
		t.Fatal("Status is not changed")
	}
	if math.Abs(sub.Overview.Time-1.234) >= 0.00001 {
		t.Fatal("Time is not changed:", sub.Overview.Time)
	}
	if sub.Overview.Memory != 5678 {
		t.Fatal("Memory is not changed")
	}
	assertEqualCases(t, cases[0:3], sub.CaseResults)

	if _, err = client.FinishJudgeTask(judgeCtx, &pb.FinishJudgeTaskRequest{
		JudgeName:    "judge-test",
		Status:       "TLE",
		Time:         2.345,
		Memory:       6789,
		SubmissionId: id,
		CaseVersion:  "test-version",
	}); err != nil {
		t.Fatal("Failed to SyncJudgeTaskStatus:", err)
	}

	sub = testFetchSubmission(t, id, client)
	if sub.Overview.Status != "TLE" {
		t.Fatal("Status is not changed")
	}
	if math.Abs(sub.Overview.Time-2.345) >= 0.00001 {
		t.Fatal("Time is not changed:", sub.Overview.Time)
	}
	if sub.Overview.Memory != 6789 {
		t.Fatal("Memory is not changed")
	}
}

func TestSimulateRejudge(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	judgeCtx := loginAsAdmin(t, client)
	id := submitSomething(t, client)

	for i := 0; i < 3; i++ {
		log.Printf("Start %v/3", i+1)
		if i > 0 {
			if _, err := client.Rejudge(judgeCtx, &pb.RejudgeRequest{
				Id: id,
			}); err != nil {
				t.Fatal("Failed to Rejudge:", err)
			}
		}
		resp, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
			JudgeName: "judge-test",
		})
		if err != nil {
			t.Fatal("Failed to PopJudgeTask:", err)
		}

		if id != resp.SubmissionId {
			t.Fatalf("ID is differ, %v vs %v", id, resp.SubmissionId)
		}

		if _, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
			JudgeName: "judge-test",
			Status:    "Judging",
			CaseResults: []*pb.SubmissionCaseResult{
				{
					Case:   "test00",
					Status: "AC",
					Time:   1.0,
					Memory: 1,
				},
			},
			SubmissionId: id,
		}); err != nil {
			t.Fatal("Failed to SyncJudgeTaskStatus:", err)
		}

		if _, err = client.FinishJudgeTask(judgeCtx, &pb.FinishJudgeTaskRequest{
			JudgeName:    "judge-test",
			Status:       "AC",
			SubmissionId: id,
		}); err != nil {
			t.Fatal("Failed to FinishJudgeTaskStatus:", err)
		}
	}
}

func TestSimulateHack(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	judgeCtx := loginAsAdmin(t, client)
	id := submitSomething(t, client)

	resp, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
		JudgeName: "judge-test",
	})
	if err != nil {
		t.Fatal("Failed to PopJudgeTask:", err)
	}

	if id != resp.SubmissionId {
		t.Fatalf("ID is differ, %v vs %v", id, resp.SubmissionId)
	}

	if _, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
		JudgeName: "judge-test",
		Status:    "Judging",
		CaseResults: []*pb.SubmissionCaseResult{
			{
				Case:   "test00",
				Status: "AC",
				Time:   1.0,
				Memory: 1,
			},
		},
		SubmissionId: id,
	}); err != nil {
		t.Fatal("Failed to SyncJudgeTaskStatus:", err)
	}

	if _, err = client.FinishJudgeTask(judgeCtx, &pb.FinishJudgeTaskRequest{
		JudgeName:    "judge-test",
		Status:       "AC",
		SubmissionId: id,
	}); err != nil {
		t.Fatal("Failed to FinishJudgeTaskStatus:", err)
	}

	// AC -> WA
	if _, err := client.Rejudge(judgeCtx, &pb.RejudgeRequest{
		Id: id,
	}); err != nil {
		t.Fatal("Failed to Rejudge:", err)
	}

	resp, err = client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
		JudgeName: "judge-test",
	})
	if err != nil {
		t.Fatal("Failed to PopJudgeTask:", err)
	}

	if id != resp.SubmissionId {
		t.Fatalf("ID is differ, %v vs %v", id, resp.SubmissionId)
	}

	if testFetchSubmission(t, id, client).Overview.Hacked {
		t.Fatal("Hacked should not be true")
	}

	if _, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
		JudgeName: "judge-test",
		Status:    "Judging",
		CaseResults: []*pb.SubmissionCaseResult{
			{
				Case:   "test00",
				Status: "WA",
				Time:   1.0,
				Memory: 1,
			},
		},
		SubmissionId: id,
	}); err != nil {
		t.Fatal("Failed to SyncJudgeTaskStatus:", err)
	}

	if _, err = client.FinishJudgeTask(judgeCtx, &pb.FinishJudgeTaskRequest{
		JudgeName:    "judge-test",
		Status:       "WA",
		SubmissionId: id,
	}); err != nil {
		t.Fatal("Failed to FinishJudgeTaskStatus:", err)
	}

	if !testFetchSubmission(t, id, client).Overview.Hacked {
		t.Fatal("Hacked should be true")
	}

	// WA -> TLE
	if _, err := client.Rejudge(judgeCtx, &pb.RejudgeRequest{
		Id: id,
	}); err != nil {
		t.Fatal("Failed to Rejudge:", err)
	}

	resp, err = client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
		JudgeName: "judge-test",
	})
	if err != nil {
		t.Fatal("Failed to PopJudgeTask:", err)
	}

	if id != resp.SubmissionId {
		t.Fatalf("ID is differ, %v vs %v", id, resp.SubmissionId)
	}

	if !testFetchSubmission(t, id, client).Overview.Hacked {
		t.Fatal("Hacked should be true")
	}

	if _, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
		JudgeName: "judge-test",
		Status:    "Judging",
		CaseResults: []*pb.SubmissionCaseResult{
			{
				Case:   "test00",
				Status: "TLE",
				Time:   1.0,
				Memory: 1,
			},
		},
		SubmissionId: id,
	}); err != nil {
		t.Fatal("Failed to SyncJudgeTaskStatus:", err)
	}

	if _, err = client.FinishJudgeTask(judgeCtx, &pb.FinishJudgeTaskRequest{
		JudgeName:    "judge-test",
		Status:       "TLE",
		SubmissionId: id,
	}); err != nil {
		t.Fatal("Failed to FinishJudgeTaskStatus:", err)
	}

	if !testFetchSubmission(t, id, client).Overview.Hacked {
		t.Fatal("Hacked should be true")
	}

	subList, err := client.SubmissionList(judgeCtx, &pb.SubmissionListRequest{
		Hacked: true,
		Skip:   0,
		Limit:  1000,
	})
	if err != nil {
		t.Fatal("Failed to SubmissionList", err)
	}
	if subList.Count == 0 {
		t.Fatal("Cannot get hacked submission")
	}
	for _, sub := range subList.Submissions {
		if !sub.Hacked {
			t.Fatal("List unhacked submission")
		}
	}
}

func TestSimulateJudgeDown(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	judgeCtx := loginAsAdmin(t, client)
	id := submitSomething(t, client)

	for i := 0; i < 3; i++ {
		log.Printf("Start %v/3", i+1)
		resp, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
			JudgeName:    "judge-test",
			ExpectedTime: durationpb.New(2 * time.Second),
		})
		if err != nil {
			t.Fatal(err)
		}
		if id != resp.SubmissionId {
			t.Fatalf("ID is differ, %v vs %v", id, resp.SubmissionId)
		}
		time.Sleep(time.Second * 4)
	}
}

func TestParallelJudge(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	judgeCtx := loginAsAdmin(t, client)
	ids := make([]int, 100)
	tasks := make([]int, 100)
	for i := 0; i < 100; i++ {
		ids[i] = -1
		tasks[i] = -1
	}
	var g errgroup.Group
	for i := 0; i < 100; i++ {
		i := i
		g.Go(func() error {
			ids[i] = int(submitSomething(t, client))
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		i := i
		g.Go(func() error {
			resp, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
				JudgeName:    "judge-test",
				ExpectedTime: durationpb.New(time.Minute),
			})
			if err != nil {
				return err
			}
			if resp.SubmissionId == -1 {
				return errors.New("Cannot fetch tasks")
			}
			log.Print("Returned ", i, resp.SubmissionId)
			tasks[i] = int(resp.SubmissionId)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
	sort.Ints(ids)
	sort.Ints(tasks)
	t.Log(ids)
	t.Log(tasks)
	for i := 0; i < 100; i++ {
		if ids[i] != tasks[i] {
			t.Fatal(i)
		}
	}
}

func TestSimulateParallelRejudge(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	judgeCtx := loginAsAdmin(t, client)

	ids := make([]int, 50)
	for i := 0; i < 50; i++ {
		ids[i] = -1
	}
	var g errgroup.Group
	for i := 0; i < 50; i++ {
		i := i
		g.Go(func() error {
			ids[i] = int(submitSomething(t, client))
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
	sort.Ints(ids)
	t.Log(ids)

	for phase := 0; phase < 3; phase++ {
		log.Printf("Start %v/3", phase+1)
		if phase > 0 {
			for _, id := range ids {
				id := id
				g.Go(func() error {
					if _, err := client.Rejudge(judgeCtx, &pb.RejudgeRequest{
						Id: int32(id),
					}); err != nil {
						return err
					}
					return nil
				})
			}
			if err := g.Wait(); err != nil {
				t.Fatal(err)
			}
		}
		tasks := make([]int, 50)
		for i := 0; i < 50; i++ {
			tasks[i] = -1
		}
		for i := 0; i < 50; i++ {
			i := i
			g.Go(func() error {
				resp, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
					JudgeName:    "judge-test",
					ExpectedTime: durationpb.New(2 * time.Second),
				})
				if err != nil {
					return err
				}

				id := resp.SubmissionId
				if id == -1 {
					return errors.New("Cannot fetch task")
				}

				if _, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
					JudgeName: "judge-test",
					Status:    "Judging",
					CaseResults: []*pb.SubmissionCaseResult{
						{
							Case:   "test00",
							Status: "AC",
							Time:   1.0,
							Memory: 1,
						},
					},
					SubmissionId: id,
				}); err != nil {
					return err
				}

				if _, err = client.FinishJudgeTask(judgeCtx, &pb.FinishJudgeTaskRequest{
					JudgeName:    "judge-test",
					Status:       "AC",
					SubmissionId: id,
				}); err != nil {
					return err
				}
				tasks[i] = int(id)
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			t.Fatal(err)
		}
		time.Sleep(4 * time.Second)

		for i := 0; i < 50; i++ {
			g.Go(func() error {
				resp, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
					JudgeName:    "judge-test",
					ExpectedTime: durationpb.New(2 * time.Second),
				})
				if err != nil {
					return err
				}

				id := resp.SubmissionId
				if id != -1 {
					return errors.New(fmt.Sprint("Fetch new task", id))
				}
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			t.Fatal(err)
		}

		sort.Ints(tasks)
		t.Log(tasks)
		for i := 0; i < 50; i++ {
			if ids[i] != tasks[i] {
				t.Fatal(i)
			}
		}
	}
}

func TestChangeProblemInfo(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	ctx := loginAsAdmin(t, client)

	oldProblem, err := client.ProblemInfo(ctx, &pb.ProblemInfoRequest{
		Name: "aplusb",
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.ChangeProblemInfo(ctx, &pb.ChangeProblemInfoRequest{
		Name:        "aplusb",
		Title:       "dummy-title",
		TimeLimit:   123.0,
		Statement:   "dummy-statement",
		CaseVersion: "dummy-version",
	}); err != nil {
		t.Fatal(err)
	}

	problem, err := client.ProblemInfo(ctx, &pb.ProblemInfoRequest{
		Name: "aplusb",
	})
	if err != nil {
		t.Fatal(err)
	}

	if problem.Title != "dummy-title" {
		t.Fatal("Title is not changed")
	}
	if problem.TimeLimit != 123.0 {
		t.Fatalf("TimeLimit is not changed: %v", problem.TimeLimit)
	}
	if problem.Statement != "dummy-statement" {
		t.Fatal("statement is not changed")
	}
	if problem.CaseVersion != "dummy-version" {
		t.Fatal("CaseVersion is not changed: ", problem.CaseVersion)
	}

	if _, err := client.ChangeProblemInfo(ctx, &pb.ChangeProblemInfoRequest{
		Name:        "aplusb",
		Title:       oldProblem.Title,
		TimeLimit:   oldProblem.TimeLimit,
		Statement:   oldProblem.Statement,
		CaseVersion: oldProblem.CaseVersion,
	}); err != nil {
		t.Fatal(err)
	}

	nowProblem, err := client.ProblemInfo(ctx, &pb.ProblemInfoRequest{
		Name: "aplusb",
	})
	if err != nil {
		t.Fatal(err)
	}

	if nowProblem.Title != oldProblem.Title {
		t.Fatal("Title is not recovered")
	}
	if nowProblem.TimeLimit != oldProblem.TimeLimit {
		t.Fatalf("TimeLimit is not recovered: %v vs %v", nowProblem.TimeLimit, problem.TimeLimit)
	}
	if nowProblem.Statement != oldProblem.Statement {
		t.Fatal("Statement is not changed")
	}
	if nowProblem.CaseVersion != oldProblem.CaseVersion {
		t.Fatal("CaseVersion is not changed")
	}
}

func TestChangeProblemInfoByTester(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	ctx := loginAsTester(t, client)

	_, err := client.ProblemInfo(ctx, &pb.ProblemInfoRequest{
		Name: "aplusb",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.ChangeProblemInfo(ctx, &pb.ChangeProblemInfoRequest{
		Name:        "aplusb",
		Title:       "dummy-title",
		TimeLimit:   123.0,
		Statement:   "dummy-statement",
		CaseVersion: "dummy-version",
	})
	if err == nil {
		t.Fatal("success to ChangeProblemInfo")
	}
	t.Log(err)
}

func TestCreateProblem(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	ctx := loginAsAdmin(t, client)

	name := uuid.New().String()
	if _, err := client.ChangeProblemInfo(ctx, &pb.ChangeProblemInfoRequest{
		Name:        name,
		Title:       "dummy-title-x",
		TimeLimit:   1234.0,
		Statement:   "dummy-statement-x",
		CaseVersion: "dummy-version-x",
	}); err != nil {
		t.Fatal("Failed to create problem")
	}

	problem, err := client.ProblemInfo(ctx, &pb.ProblemInfoRequest{
		Name: name,
	})
	if err != nil {
		t.Fatal(err)
	}

	if problem.Title != "dummy-title-x" {
		t.Fatal("Title is invalid")
	}
	if problem.TimeLimit != 1234.0 {
		t.Fatalf("TimeLimit is invalid: %v", problem.TimeLimit)
	}
	if problem.Statement != "dummy-statement-x" {
		t.Fatal("statement is invalid")
	}
	if problem.CaseVersion != "dummy-version-x" {
		t.Fatal("CaseVersion is invalid")
	}
}

func TestProblemCategories(t *testing.T) {
	client, close := createAPIClient(t, createTestDB(t))
	defer close()

	ctx := loginAsAdmin(t, client)
	testData := []*pb.ProblemCategory{
		{
			Title:    "a",
			Problems: []string{"x", "y"},
		},
		{
			Title:    "b",
			Problems: []string{"z"},
		},
	}

	for phase := 0; phase <= 1; phase++ {

		if _, err := client.ChangeProblemCategories(ctx, &pb.ChangeProblemCategoriesRequest{
			Categories: testData,
		}); err != nil {
			t.Fatal(err)
		}
		res, err := client.ProblemCategories(ctx, &pb.ProblemCategoriesRequest{})
		if err != nil {
			t.Fatal(err)
		}
		expect := testData
		actual := res.Categories

		if len(expect) != len(actual) {
			t.Fatal("len is differ")
		}

		for i := 0; i < len(expect); i++ {
			if expect[i].Title != actual[i].Title {
				t.Fatal("title is differ")
			}
			if !reflect.DeepEqual(expect[i].Problems, actual[i].Problems) {
				t.Fatal("problems is differ")
			}
		}

		testData = append(testData, &pb.ProblemCategory{
			Title:    "c",
			Problems: []string{"w"},
		})
	}
}
