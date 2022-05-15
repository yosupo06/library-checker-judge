package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	_ "github.com/lib/pq"
	"github.com/yosupo06/library-checker-judge/api/clientutil"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
)

// gRPC
var client pb.LibraryCheckerServiceClient
var judgeName string
var judgeCtx context.Context
var testCaseFetcher TestCaseFetcher

func execJudge(submissionID int32) (err error) {
	submission, err := client.SubmissionInfo(judgeCtx, &pb.SubmissionInfoRequest{
		Id: submissionID,
	})
	if err != nil {
		return err
	}
	problem, err := client.ProblemInfo(judgeCtx, &pb.ProblemInfoRequest{
		Name: submission.Overview.ProblemName,
	})
	log.Println("Submission info:", submissionID, submission.Overview.ProblemTitle)

	log.Println("Fetch data")
	if _, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
		JudgeName:    judgeName,
		SubmissionId: submissionID,
		Status:       "Fetching",
	}); err != nil {
		return err
	}

	caseVersion := problem.CaseVersion
	testCases, err := testCaseFetcher.Fetch(submission.Overview.ProblemName, caseVersion)
	log.Print("Fetched :", caseVersion)
	if err != nil {
		log.Println("Fail to fetchData")
		return err
	}

	checker, err := os.Open(testCases.CheckerPath())
	if err != nil {
		return err
	}
	judge, err := NewJudge(submission.Overview.Lang, problem.TimeLimit)
	if err != nil {
		return err
	}
	defer judge.Close()

	defer func() {
		if err != nil {
			// error detected, try to change status into IE
			client.FinishJudgeTask(judgeCtx, &pb.FinishJudgeTaskRequest{
				JudgeName:    judgeName,
				SubmissionId: submissionID,
				Status:       "IE",
				CaseVersion:  caseVersion,
			})
		}
	}()

	if _, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
		JudgeName:    judgeName,
		SubmissionId: submissionID,
		Status:       "Compiling",
	}); err != nil {
		return err
	}
	result, err := judge.CompileChecker(checker)
	if err != nil {
		return err
	}
	if result.ReturnCode != 0 {
		if _, err = client.FinishJudgeTask(judgeCtx, &pb.FinishJudgeTaskRequest{
			JudgeName:    judgeName,
			SubmissionId: submissionID,
			Status:       "ICE",
			CaseVersion:  caseVersion,
		}); err != nil {
			return err
		}
		return nil
	}
	result, err = judge.CompileSource(strings.NewReader(submission.Source))
	if err != nil {
		return err
	}
	if result.ReturnCode != 0 {
		if _, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
			JudgeName:    judgeName,
			SubmissionId: submissionID,
			Status:       "CE",
			CompileError: result.Stderr,
		}); err != nil {
			return err
		}

		if _, err = client.FinishJudgeTask(judgeCtx, &pb.FinishJudgeTaskRequest{
			JudgeName:    judgeName,
			SubmissionId: submissionID,
			CaseVersion:  caseVersion,
		}); err != nil {
			return err
		}
		return nil
	}
	if _, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
		JudgeName:    judgeName,
		SubmissionId: submissionID,
		Status:       "Executing",
	}); err != nil {
		return err
	}

	cases, err := testCases.CaseNames()
	if err != nil {
		return err
	}
	unsendCases := []CaseResult{}
	sendCase := func() error {
		cases := []*pb.SubmissionCaseResult{}
		for _, caseResult := range unsendCases {
			cases = append(cases, &pb.SubmissionCaseResult{
				Case:   caseResult.CaseName,
				Status: caseResult.Status,
				Time:   caseResult.Time,
				Memory: int64(caseResult.Memory),
			})
		}
		if _, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
			JudgeName:    judgeName,
			SubmissionId: submissionID,
			Status:       "Executing",
			CaseResults:  cases,
		}); err != nil {
			return err
		}
		unsendCases = []CaseResult{}
		return nil
	}
	caseResults := []CaseResult{}
	lastSend := time.Time{}
	addCase := func(caseResult *CaseResult) error {
		if caseResult != nil {
			caseResults = append(caseResults, *caseResult)
			unsendCases = append(unsendCases, *caseResult)
		}
		if lastSend.Add(time.Second).Before(time.Now()) {
			lastSend = time.Now()
			if err := sendCase(); err != nil {
				return err
			}
		}
		return nil
	}
	for _, caseName := range cases {
		inFile, err := os.Open(testCases.InFilePath(caseName))
		if err != nil {
			return err
		}
		outFile, err := os.Open(testCases.OutFilePath(caseName))
		if err != nil {
			return err
		}
		caseResult, err := judge.TestCase(inFile, outFile)
		if err != nil {
			return err
		}
		caseResult.CaseName = caseName
		if err := addCase(&caseResult); err != nil {
			return err
		}
	}
	if err := sendCase(); err != nil {
		return err
	}
	caseResult := AggregateResults(caseResults)
	if _, err = client.FinishJudgeTask(judgeCtx, &pb.FinishJudgeTaskRequest{
		JudgeName:    judgeName,
		SubmissionId: submissionID,
		Status:       caseResult.Status,
		Time:         caseResult.Time,
		Memory:       int64(caseResult.Memory),
		CaseVersion:  caseVersion,
	}); err != nil {
		return err
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

var secretConfig struct {
	APIHost     string `toml:"api_host"`
	APIUser     string `toml:"api_user"`
	APIPass     string `toml:"api_pass"`
	MinioHost   string `toml:"minio_host"`
	MinioAccess string `toml:"minio_access"`
	MinioSecret string `toml:"minio_secret"`
	MinioBucket string `toml:"minio_bucket"`
	Prod        bool   `toml:"prod"`
}

func init() {
	if _, err := toml.DecodeFile("./secret.toml", &secretConfig); err != nil {
		log.Fatal(err)
	}
}

func initClient(conn *grpc.ClientConn) {
	client = pb.NewLibraryCheckerServiceClient(conn)
	ctx := context.Background()
	resp, err := client.Login(ctx, &pb.LoginRequest{
		Name:     secretConfig.APIUser,
		Password: secretConfig.APIPass,
	})

	if err != nil {
		log.Fatal("Cannot login to API Server:", err)
	}
	judgeCtx = clientutil.ContextWithToken(ctx, resp.Token)

	judgeName, err = os.Hostname()
	if err != nil {
		log.Fatal("Cannot get hostname:", err)
	}
	log.Print("JudgeName: ", judgeName)
}

func apiConnect() *grpc.ClientConn {
	options := []grpc.DialOption{grpc.WithBlock(), grpc.WithPerRPCCredentials(&clientutil.LoginCreds{}), grpc.WithTimeout(10 * time.Second)}
	if !secretConfig.Prod {
		log.Print("local mode")
		options = append(options, grpc.WithInsecure())
	} else {
		systemRoots, err := x509.SystemCertPool()
		if err != nil {
			log.Fatal(err)
		}
		creds := credentials.NewTLS(&tls.Config{
			RootCAs: systemRoots,
		})
		options = append(options, grpc.WithTransportCredentials(creds))
	}
	log.Printf("Connect to API host: %v", secretConfig.APIHost)
	conn, err := grpc.Dial(secretConfig.APIHost, options...)
	if err != nil {
		log.Fatal("Cannot connect to the API server:", err)
	}
	return conn
}

func main() {
	var err error

	// init gRPC
	conn := apiConnect()
	defer conn.Close()
	initClient(conn)

	testCaseFetcher, err = NewTestCaseFetcher(
		secretConfig.MinioHost,
		secretConfig.MinioAccess,
		secretConfig.MinioSecret,
		secretConfig.Prod,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer testCaseFetcher.Close()

	log.Println("Start Pooling")
	for {
		task, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
			JudgeName: judgeName,
		})
		if err != nil {
			time.Sleep(3 * time.Second)
			log.Print("PopJudgeTask error: ", err)
			continue
		}
		if task.SubmissionId == -1 {
			time.Sleep(3 * time.Second)
			continue
		}
		log.Println("Start Judge:", task.SubmissionId)
		err = execJudge(task.SubmissionId)
		if err != nil {
			log.Println(err.Error())
			continue
		}
	}
}
