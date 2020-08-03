package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/yosupo06/library-checker-judge/api/clientutil"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"google.golang.org/grpc"
)

func Submit(t *testing.T, problem, lang string, srcFile io.Reader) int32 {
	src, err := ioutil.ReadAll(srcFile)
	if err != nil {
		t.Fatal("Cannot read file:", err)
	}

	resp, err := client.Submit(judgeCtx, &pb.SubmitRequest{
		Problem: problem,
		Lang:    lang,
		Source:  string(src),
	})
	t.Log("Submit: ", resp.Id)

	if err != nil {
		t.Fatal("Submit Failed:", err)
	}
	return resp.Id
}

func TestMain(m *testing.M) {
	myCasesDir, err := ioutil.TempDir("", "case")
	casesDir = myCasesDir
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := os.RemoveAll(casesDir); err != nil {
			panic(err)
		}
	}()

	options := []grpc.DialOption{grpc.WithBlock(), grpc.WithPerRPCCredentials(&clientutil.LoginCreds{}), grpc.WithInsecure(), grpc.WithTimeout(3 * time.Second)}
	conn, err := grpc.Dial("localhost:50051", options...)
	if err != nil {
		panic(err)
	}
	initClient(conn)
	defer func() {
		if err := conn.Close(); err != nil {
			panic(err)
		}
	}()

	minioClient = minioConnect()

	os.Exit(m.Run())
}

func runJudge(t *testing.T, id int32) *pb.SubmissionInfoResponse {
	task, err := client.PopJudgeTask(judgeCtx, &pb.PopJudgeTaskRequest{
		JudgeName: judgeName,
	})
	if err != nil {
		t.Fatal("PopJudgeTask error: ", err)
	}
	if task.SubmissionId != id {
		t.Fatalf("Differ ID %v vs %v", task.SubmissionId, id)
	}
	log.Println("Start Judge:", id)
	err = execJudge(id)
	if err != nil {
		t.Fatal(err.Error())
	}
	resp, err := client.SubmissionInfo(judgeCtx, &pb.SubmissionInfoRequest{
		Id: id,
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	return resp
}

func checkStatus(t *testing.T, sub *pb.SubmissionInfoResponse, expect string) {
	overview := sub.Overview

	if overview.Status != expect {
		t.Fatalf("Expect status %v, actual %v", expect, overview.Status)
	}
}

func checkTime(t *testing.T, sub *pb.SubmissionInfoResponse, expectLower float64, expectUpper float64) {
	overview := sub.Overview
	if !(expectLower <= overview.Time && overview.Time <= expectUpper) {
		t.Fatalf("Irregural consume time expect [%f, %f], actual %v", expectLower, expectUpper, overview.Time)
	}
}

func checkMemory(t *testing.T, sub *pb.SubmissionInfoResponse, expectLower int64, expectUpper int64) {
	overview := sub.Overview
	if !(expectLower <= overview.Memory && overview.Memory <= expectUpper) {
		t.Fatalf("Irregural consume time expect [%v, %v], actual %v", expectLower, expectUpper, overview.Time)
	}
}

func TestSubmitAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac.cpp")
	if err != nil {
		t.Fatal(err)
	}
	id := Submit(t, "aplusb", "cpp", src)
	submission := runJudge(t, id)
	overview := submission.Overview

	checkStatus(t, submission, "AC")
	checkTime(t, submission, 0.001, 0.100)
	checkMemory(t, submission, 1, 10_000_000)
	if !overview.IsLatest {
		t.Fatal("Not latest")
	}
}

func TestRejudge(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac.cpp")
	if err != nil {
		t.Fatal(err)
	}
	id := Submit(t, "aplusb", "cpp", src)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("submit ok ", id)
	for i := 0; i < 3; i++ {
		if i > 0 {
			client.Rejudge(judgeCtx, &pb.RejudgeRequest{
				Id: id,
			})
		}
		submission := runJudge(t, id)
		overview := submission.Overview
		checkStatus(t, submission, "AC")
		checkTime(t, submission, 0.001, 0.100)
		checkMemory(t, submission, 1, 10_000_000)
		if !overview.IsLatest {
			t.Fatal("Not latest")
		}
	}
}

func TestSubmitWA(t *testing.T) {
	src, err := os.Open("test_src/aplusb/wa.cpp")
	if err != nil {
		t.Fatal(err)
	}
	id := Submit(t, "aplusb", "cpp", src)
	submission := runJudge(t, id)
	checkStatus(t, submission, "WA")
	checkTime(t, submission, 0.001, 0.100)
}

func TestSubmitTLE(t *testing.T) {
	src, err := os.Open("test_src/TLE.cpp")
	if err != nil {
		t.Fatal(err)
	}
	id := Submit(t, "aplusb", "cpp", src)
	submission := runJudge(t, id)
	checkStatus(t, submission, "TLE")
	checkTime(t, submission, 1.900, 2.100)
}

func TestSubmitRE(t *testing.T) {
	src, err := os.Open("test_src/aplusb/re.cpp")
	if err != nil {
		t.Fatal(err)
	}
	id := Submit(t, "aplusb", "cpp", src)
	submission := runJudge(t, id)
	checkStatus(t, submission, "RE")
	checkTime(t, submission, 0.001, 0.100)
}

func TestSubmitCE(t *testing.T) {
	id := Submit(t, "aplusb", "cpp", strings.NewReader("The answer is 42..."))
	submission := runJudge(t, id)
	checkStatus(t, submission, "CE")
}

func TestSubmitRustAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac.rs")
	if err != nil {
		t.Fatal(err)
	}
	id := Submit(t, "aplusb", "rust", src)
	submission := runJudge(t, id)
	overview := submission.Overview

	checkStatus(t, submission, "AC")
	checkTime(t, submission, 0.001, 0.100)
	checkMemory(t, submission, 1, 10_000_000)
	if !overview.IsLatest {
		t.Fatal("Not latest")
	}
}

func TestSubmitHaskellAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac.hs")
	if err != nil {
		t.Fatal(err)
	}
	id := Submit(t, "aplusb", "haskell", src)
	submission := runJudge(t, id)
	overview := submission.Overview

	checkStatus(t, submission, "AC")
	checkTime(t, submission, 0.001, 0.100)
	checkMemory(t, submission, 1, 10_000_000)
	if !overview.IsLatest {
		t.Fatal("Not latest")
	}
}

func TestSubmitHaskellStackAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac_stack.hs")
	if err != nil {
		t.Fatal(err)
	}
	id := Submit(t, "aplusb", "haskell", src)
	submission := runJudge(t, id)
	overview := submission.Overview

	checkStatus(t, submission, "AC")
	checkTime(t, submission, 0.001, 0.100)
	checkMemory(t, submission, 1, 10_000_000)
	if !overview.IsLatest {
		t.Fatal("Not latest")
	}
}

func TestSubmitCSharpAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac.cs")
	if err != nil {
		t.Fatal(err)
	}
	id := Submit(t, "aplusb", "csharp", src)
	submission := runJudge(t, id)
	overview := submission.Overview

	checkStatus(t, submission, "AC")
	checkTime(t, submission, 0.001, 1.000)
	checkMemory(t, submission, 1, 100_000_000)
	if !overview.IsLatest {
		t.Fatal("Not latest")
	}
}

func TestSubmitPythonAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac_numpy.py")
	if err != nil {
		t.Fatal(err)
	}
	id := Submit(t, "aplusb", "python3", src)
	submission := runJudge(t, id)
	overview := submission.Overview

	checkStatus(t, submission, "AC")
	checkTime(t, submission, 0.001, 0.500)
	if !overview.IsLatest {
		t.Fatal("Not latest")
	}
}

func TestSubmitPyPyAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac.py")
	if err != nil {
		t.Fatal(err)
	}
	id := Submit(t, "aplusb", "pypy3", src)
	submission := runJudge(t, id)
	overview := submission.Overview

	checkStatus(t, submission, "AC")
	checkTime(t, submission, 0.001, 0.500)
	if !overview.IsLatest {
		t.Fatal("Not latest")
	}
}

func TestSubmitDlangAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac.d")
	if err != nil {
		t.Fatal(err)
	}
	id := Submit(t, "aplusb", "d", src)
	submission := runJudge(t, id)
	overview := submission.Overview

	checkStatus(t, submission, "AC")
	checkTime(t, submission, 0.001, 0.100)
	if !overview.IsLatest {
		t.Fatal("Not latest")
	}
}

func TestSubmitJavaAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac.java")
	if err != nil {
		t.Fatal(err)
	}
	id := Submit(t, "aplusb", "java", src)
	submission := runJudge(t, id)
	overview := submission.Overview

	checkStatus(t, submission, "AC")
	checkTime(t, submission, 0.001, 1.000)
	if !overview.IsLatest {
		t.Fatal("Not latest")
	}
}

func TestSubmitGoAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac.go")
	if err != nil {
		t.Fatal(err)
	}
	id := Submit(t, "aplusb", "go", src)
	submission := runJudge(t, id)
	overview := submission.Overview

	checkStatus(t, submission, "AC")
	checkTime(t, submission, 0.001, 0.100)
	if !overview.IsLatest {
		t.Fatal("Not latest")
	}
}

func TestSubmitCommonLispAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac.lisp")
	if err != nil {
		t.Fatal(err)
	}
	id := Submit(t, "aplusb", "lisp", src)
	submission := runJudge(t, id)
	overview := submission.Overview

	checkStatus(t, submission, "AC")
	checkTime(t, submission, 0.001, 0.500)
	if !overview.IsLatest {
		t.Fatal("Not latest")
	}
}
