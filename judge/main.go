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

type Submission struct {
	ID          int
	ProblemName string
	Problem     Problem `gorm:"foreignkey:ProblemName"`
	Lang        string
	UserName    string
	Status      string
	Source      string
	Testhash    string
	MaxTime     int
	MaxMemory   int
	JudgePing   time.Time
}

type Task struct {
	Submission int
}

type SubmissionTestcaseResult struct {
	Submission int
	Testcase   string
	Status     string
	Time       int
	Memory     int
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
		Memory:       int64(caseResult.Time),
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

func gormConnect() *gorm.DB {
	var sqlInfo struct {
		PostgreHost string `toml:"postgre_host"`
		PostgreUser string `toml:"postgre_user"`
		PostgrePass string `toml:"postgre_pass"`
	}
	if _, err := toml.DecodeFile("./secret.toml", &sqlInfo); err != nil {
		log.Fatal(err)
	}
	connStr := fmt.Sprintf(
		"host=%s port=5432 user=%s dbname=librarychecker password=%s sslmode=disable",
		sqlInfo.PostgreHost, sqlInfo.PostgreUser, sqlInfo.PostgrePass)

	log.Println("Connected to DB:", sqlInfo.PostgreHost)
	db, err := gorm.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db
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

func apiConnect() (*grpc.ClientConn, pb.LibraryCheckerServiceClient, context.Context) {
	var apiInfo struct {
		ApiHost string `toml:"api_host"`
		ApiUser string `toml:"api_user"`
		ApiPass string `toml:"api_pass"`
	}
	if _, err := toml.DecodeFile("./secret.toml", &apiInfo); err != nil {
		log.Fatal(err)
	}

	options := []grpc.DialOption{grpc.WithBlock(), grpc.WithPerRPCCredentials(&loginCreds{}), grpc.WithTimeout(10 * time.Second)}
	if strings.HasPrefix(apiInfo.ApiHost, "localhost") {
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
	log.Printf("Connect to API host: %v", apiInfo.ApiHost)
	conn, err := grpc.Dial(apiInfo.ApiHost, options...)
	if err != nil {
		log.Fatal("Cannot connect to the API server:", err)
	}
	client = pb.NewLibraryCheckerServiceClient(conn)
	ctx := context.Background()
	resp, err := client.Login(ctx, &pb.LoginRequest{
		Name:     apiInfo.ApiUser,
		Password: apiInfo.ApiPass,
	})

	if err != nil {
		log.Fatal("Cannot login to API Server:", err)
	}
	judgeCtx = context.WithValue(ctx, tokenKey{}, resp.Token)

	return conn, client, judgeCtx
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
	db.AutoMigrate(Problem{})
	db.AutoMigrate(Submission{})
	db.AutoMigrate(Task{})
	db.AutoMigrate(SubmissionTestcaseResult{})
	// db.LogMode(true)

	// init gRPC
	var conn *grpc.ClientConn
	conn, client, judgeCtx = apiConnect()
	defer conn.Close()

	judgeName, err = os.Hostname()
	if err != nil {
		log.Fatal("Cannot get hostname:", err)
	}
	log.Println("Start Pooling JudgeName:", judgeName)
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
