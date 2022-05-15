package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/google/shlex"
	_ "github.com/lib/pq"
)

type Result struct {
	ReturnCode int     `json:"returncode"`
	Time       float64 `json:"time"`
	Memory     int     `json:"memory"`
	Tle        bool    `json:"tle"`
	Stderr     []byte
}

func SafeRun(cmd *exec.Cmd, tl float64, overlay bool) (Result, error) {
	newArg := []string{}
	newArg = append(newArg, "--tl", strconv.FormatFloat(tl, 'f', 4, 64))
	if overlay {
		newArg = append(newArg, "--overlay")
	}
	tmpfile, err := ioutil.TempFile("", "result")
	if err != nil {
		return Result{}, err
	}
	newArg = append(newArg, "--result", tmpfile.Name())
	newArg = append(newArg, "--")
	newArg = append(newArg, cmd.Args...)

	if cmd.Path, err = exec.LookPath("executor"); err != nil {
		return Result{}, err
	}
	cmd.Args = append([]string{"executor"}, newArg...)
	// add stderr
	os := &outputStripper{N: 1 << 11}
	if cmd.Stderr != nil {
		cmd.Stderr = io.MultiWriter(cmd.Stderr, os)
	} else {
		cmd.Stderr = os
	}

	err = cmd.Run()
	if err != nil && cmd.ProcessState.ExitCode() != 124 {
		return Result{ReturnCode: -1, Time: -1, Memory: -1}, err
	}
	raw, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		return Result{}, err
	}
	result := Result{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return Result{}, err
	}
	result.Stderr = os.Bytes()
	log.Println("execute: ", cmd.Args)
	log.Printf("stderr: %s\n", string(result.Stderr))
	return result, nil
}

type Judge struct {
	dir  string
	tl   float64
	lang Lang

	checkerCompiled bool
	sourceCompiled  bool
}

func NewJudge(lang string, tl float64) (*Judge, error) {
	tempdir, err := ioutil.TempDir("", "judge")
	if err != nil {
		return nil, err
	}

	log.Println("New judge:", tempdir)

	judge := new(Judge)

	judge.lang = langs[lang]
	judge.tl = tl
	judge.dir = tempdir

	return judge, nil
}

func (j *Judge) Close() error {
	return os.RemoveAll(j.dir)
}

func (j *Judge) checkerPath() string {
	return path.Join(j.dir, "checker")
}
func (j *Judge) sourcePath() string {
	return path.Join(j.dir, "source")
}

func (j *Judge) CompileChecker(checker io.Reader) (Result, error) {
	// create dir for checker
	if err := os.Mkdir(j.checkerPath(), 0755); err != nil {
		return Result{}, err
	}
	if err := os.Chmod(j.checkerPath(), 0777); err != nil {
		return Result{}, err
	}

	tempChecker, err := os.Create(path.Join(j.checkerPath(), "checker.cpp"))
	if err != nil {
		return Result{}, err
	}
	if _, err = io.Copy(tempChecker, checker); err != nil {
		return Result{}, err
	}
	testlib, err := os.Open("testlib.h")
	if err != nil {
		return Result{}, err
	}
	tempTestlib, err := os.Create(path.Join(j.checkerPath(), "testlib.h"))
	if err != nil {
		return Result{}, err
	}
	if _, err = io.Copy(tempTestlib, testlib); err != nil {
		return Result{}, err
	}

	compile, err := shlex.Split(langs["checker"].Compile)
	if err != nil {
		return Result{}, err
	}

	cmd := exec.Command(compile[0], compile[1:]...)
	cmd.Dir = j.checkerPath()
	cmd.Stdout = os.Stdout
	cmd.Stderr = nil

	result, err := SafeRun(cmd, 30.0, false)
	if err != nil {
		return Result{}, err
	}
	j.checkerCompiled = true
	return result, err
}

func (j *Judge) CompileSource(source io.Reader) (Result, error) {
	// create dir for source
	if err := os.Mkdir(j.sourcePath(), 0755); err != nil {
		return Result{}, err
	}
	if err := os.Chmod(j.sourcePath(), 0777); err != nil {
		return Result{}, err
	}

	tempSrc, err := os.Create(path.Join(j.sourcePath(), j.lang.Source))
	if err != nil {
		log.Print("error ", path.Join(j.sourcePath(), j.lang.Source))
		return Result{}, err
	}
	if _, err = io.Copy(tempSrc, source); err != nil {
		return Result{}, err
	}

	compile, err := shlex.Split(j.lang.Compile)
	if err != nil {
		return Result{}, err
	}

	cmd := exec.Command(compile[0], compile[1:]...)
	cmd.Dir = j.sourcePath()
	cmd.Stdout = os.Stdout
	cmd.Stderr = nil

	result, err := SafeRun(cmd, 30.0, false)
	if err != nil {
		return Result{}, err
	}
	j.sourceCompiled = true
	return result, err
}

func (j *Judge) TestCase(inFile io.Reader, expectFile io.Reader) (CaseResult, error) {
	input, err := os.Create(path.Join(j.checkerPath(), "input.in"))
	if err != nil {
		return CaseResult{}, err
	}
	if _, err = io.Copy(input, inFile); err != nil {
		return CaseResult{}, err
	}
	if _, err = input.Seek(0, 0); err != nil {
		return CaseResult{}, err
	}

	expect, err := os.Create(path.Join(j.checkerPath(), "expect.out"))
	if err != nil {
		return CaseResult{}, err
	}
	if _, err = io.Copy(expect, expectFile); err != nil {
		return CaseResult{}, err
	}
	if err = expect.Close(); err != nil {
		return CaseResult{}, err
	}

	actual, err := os.Create(path.Join(j.checkerPath(), "actual.out"))
	if err != nil {
		return CaseResult{}, err
	}

	arg := strings.Fields(j.lang.Exec)
	cmd := exec.Command(arg[0], arg[1:]...)
	cmd.Dir = j.sourcePath()
	cmd.Stdin = input
	cmd.Stdout = actual
	result, err := SafeRun(cmd, j.tl, true)

	if err != nil {
		return CaseResult{}, err
	}

	if result.Tle {
		//timeout
		return CaseResult{Status: "TLE", Result: result}, nil
	}

	if cmd.ProcessState.ExitCode() != 0 {
		return CaseResult{Status: "Broken", Result: result}, errors.New("executor return non 0, 124 code")
	}

	if result.ReturnCode != 0 {
		return CaseResult{Status: "RE", Result: result}, nil
	}
	actual.Close()

	// run checker
	cmd = exec.Command("./checker", "input.in", "actual.out", "expect.out")
	cmd.Dir = j.checkerPath()
	checkerResult, err := SafeRun(cmd, j.tl, true)
	if err != nil {
		return CaseResult{}, err
	}
	if checkerResult.Tle {
		return CaseResult{Status: "ITLE", Result: result}, nil
	}
	if cmd.ProcessState.ExitCode() != 0 {
		return CaseResult{Status: "Broken", Result: result}, errors.New("executor return non 0, 124 code")
	}
	if checkerResult.ReturnCode == 1 {
		return CaseResult{Status: "WA", Result: result}, nil
	}
	if checkerResult.ReturnCode == 2 {
		return CaseResult{Status: "PE", Result: result}, nil
	}
	if checkerResult.ReturnCode == 3 {
		return CaseResult{Status: "Fail", Result: result}, nil
	}
	if checkerResult.ReturnCode != 0 {
		return CaseResult{Status: "Unknown", Result: result}, nil
	}
	return CaseResult{Status: "AC", Result: result}, nil
}

type CaseResult struct {
	CaseName string
	Status   string
	Result
}

func AggregateResults(results []CaseResult) CaseResult {
	ans := CaseResult{
		Status: "AC",
		Result: Result{ReturnCode: -1, Time: -1, Memory: -1},
	}
	for _, res := range results {
		if res.Status != "AC" {
			ans.Status = res.Status
		}
		if ans.Time < res.Time {
			ans.Time = res.Time
		}
		if ans.Memory < res.Memory {
			ans.Memory = res.Memory
		}
	}
	return ans
}
