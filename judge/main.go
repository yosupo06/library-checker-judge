package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/yosupo06/library-checker-judge/api/clientutil"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
)

var client pb.LibraryCheckerServiceClient
var judgeName string
var judgeCtx context.Context

type Problem struct {
	Name      string
	Title     string
	Timelimit float64
	Testhash  string
	Testzip   []byte
}

var casesDir string

func fetchData(db *gorm.DB, problemName string) (string, string, error) {
	problem := Problem{}
	if err := db.Where("name = ?", problemName).Take(&problem).Error; err != nil {
		return "", "", err
	}
	zipPath := path.Join(casesDir, fmt.Sprintf("cases-%s.zip", problem.Testhash))
	data := path.Join(casesDir, fmt.Sprintf("cases-%s", problem.Testhash))
	if _, err := os.Stat(zipPath); err != nil {
		// fetch zip
		zipFile, err := os.Create(zipPath)
		if err != nil {
			return "", "", err
		}
		if _, err = zipFile.Write(problem.Testzip); err != nil {
			return "", "", err
		}
		if err = zipFile.Close(); err != nil {
			return "", "", err
		}
		cmd := exec.Command("unzip", zipPath, "-d", data)
		if err := cmd.Run(); err != nil {
			return "", "", err
		}
	}
	return data, problem.Testhash, nil
}

func getCases(data string) ([]string, error) {
	// write glob code
	matches, err := filepath.Glob(path.Join(data, "in", "*.in"))
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, match := range matches {
		_, name := path.Split(match)
		name = strings.TrimSuffix(name, ".in")
		result = append(result, name)
	}
	return result, nil
}

func execJudge(db *gorm.DB, submissionID int32) error {
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
	caseDir, caseVersion, err := fetchData(db, submission.Overview.ProblemName)
	log.Print("Fetched :", caseVersion)
	if err != nil {
		log.Println("Fail to fetchData")
		return err
	}

	checker, err := os.Open(path.Join(caseDir, "checker.cpp"))
	if err != nil {
		return err
	}
	tempdir, err := ioutil.TempDir("", "judge")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempdir)
	judge, err := NewJudge(tempdir, submission.Overview.Lang, checker, strings.NewReader(submission.Source), problem.TimeLimit)
	if err != nil {
		return err
	}

	if _, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
		JudgeName:    judgeName,
		SubmissionId: submissionID,
		Status:       "Compiling",
	}); err != nil {
		return err
	}
	result, err := judge.CompileChecker()
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
	result, err = judge.CompileSource()
	if err != nil {
		return err
	}
	if result.ReturnCode != 0 {
		if _, err = client.FinishJudgeTask(judgeCtx, &pb.FinishJudgeTaskRequest{
			JudgeName:    judgeName,
			SubmissionId: submissionID,
			Status:       "CE",
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
	cases, err := getCases(caseDir)
	if err != nil {
		return err
	}
	caseResults := []CaseResult{}
	for _, caseName := range cases {
		inFile, err := os.Open(path.Join(caseDir, "in", caseName+".in"))
		if err != nil {
			return err
		}
		outFile, err := os.Open(path.Join(caseDir, "out", caseName+".out"))
		if err != nil {
			return err
		}
		caseResult, err := judge.TestCase(inFile, outFile)
		if err != nil {
			return err
		}
		caseResults = append(caseResults, caseResult)
		if _, err = client.SyncJudgeTaskStatus(judgeCtx, &pb.SyncJudgeTaskStatusRequest{
			JudgeName:    judgeName,
			SubmissionId: submissionID,
			Status:       "Executing",
			CaseResults: []*pb.SubmissionCaseResult{
				{
					Case:   caseName,
					Status: caseResult.Status,
					Time:   caseResult.Time,
					Memory: int64(caseResult.Memory),
				},
			},
		}); err != nil {
			return err
		}
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
	PostgreHost string `toml:"postgre_host"`
	PostgreUser string `toml:"postgre_user"`
	PostgrePass string `toml:"postgre_pass"`
	ApiHost     string `toml:"api_host"`
	ApiUser     string `toml:"api_user"`
	ApiPass     string `toml:"api_pass"`
	Prod        bool   `toml:"prod"`
}

func init() {
	if _, err := toml.DecodeFile("./secret.toml", &secretConfig); err != nil {
		log.Fatal(err)
	}
}

func gormConnect() *gorm.DB {
	connStr := fmt.Sprintf(
		"host=%s port=5432 user=%s dbname=librarychecker password=%s sslmode=disable",
		secretConfig.PostgreHost, secretConfig.PostgreUser, secretConfig.PostgrePass)

	log.Println("Connected to DB:", secretConfig.PostgreHost)
	db, err := gorm.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if !secretConfig.Prod {
		db.LogMode(true)
	}
	db.AutoMigrate(Problem{})
	return db
}

func initClient(conn *grpc.ClientConn) {
	client = pb.NewLibraryCheckerServiceClient(conn)
	ctx := context.Background()
	resp, err := client.Login(ctx, &pb.LoginRequest{
		Name:     secretConfig.ApiUser,
		Password: secretConfig.ApiPass,
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
	log.Printf("Connect to API host: %v", secretConfig.ApiHost)
	conn, err := grpc.Dial(secretConfig.ApiHost, options...)
	if err != nil {
		log.Fatal("Cannot connect to the API server:", err)
	}
	return conn
}

func main() {
	// init directory
	myCasesDir, err := ioutil.TempDir(os.Getenv("CASEDIR"), "case")
	casesDir = myCasesDir
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(casesDir)

	// init DB
	db := gormConnect()
	defer db.Close()

	// init gRPC
	conn := apiConnect()
	defer conn.Close()
	initClient(conn)

	log.Println("Start Pooling")
	for {
		task, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
			JudgeName: judgeName,
		})
		if err != nil {
			time.Sleep(1 * time.Second)
			log.Print("PopJudgeTask error: ", err)
			continue
		}
		if task.SubmissionId == -1 {
			time.Sleep(1 * time.Second)
			continue
		}
		log.Println("Start Judge:", task.SubmissionId)
		err = execJudge(db, task.SubmissionId)
		if err != nil {
			log.Println(err.Error())
			continue
		}
	}
}
