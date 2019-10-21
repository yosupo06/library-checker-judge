package main

import (
	"bytes"
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

	"github.com/BurntSushi/toml"
	_ "github.com/lib/pq"
)

// Save stripped output with strip()
type outputStripper struct {
	N        int
	data     []byte
	overflow bool
}

func (os *outputStripper) Write(b []byte) (n int, err error) {
	if os.N <= 20 {
		return -1, errors.New("N is too small")
	}
	blen := len(b)
	cap := os.N - 20 - len(os.data)

	add := blen
	if cap < add {
		add = cap
		os.overflow = true
	}
	os.data = append(os.data, b[:add]...)
	return blen, nil
}

func (os *outputStripper) Bytes() []byte {
	var buf bytes.Buffer
	buf.Write(os.data)
	if os.overflow {
		buf.Write([]byte(" ... stripped"))
	}
	return buf.Bytes()
}

type Result struct {
	ReturnCode int     `json:"returncode"`
	Time       float64 `json:"time"`
	Memory     int     `json:"memory"`
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

	wd, err := os.Getwd()
	if err != nil {
		return Result{}, err
	}
	cmd.Path = path.Join(wd, "executor.py")
	cmd.Args = append([]string{cmd.Path}, newArg...)

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
	return result, nil
}

type Lang struct {
	Source  string `toml:"source"`
	Compile string `toml:"compile"`
	Exec    string `toml:"exec"`
}

var langs map[string]Lang

func init() {
	var tomlData struct {
		Langs map[string]Lang
	}
	if _, err := toml.DecodeFile("../compiler/langs.toml", &tomlData); err != nil {
		log.Fatal(err)
	}
	langs = tomlData.Langs
	if _, ok := langs["checker"]; !ok {
		log.Fatal("lang file don't have checker")
	}
}

/*
Judge conditition:

dir / checker / checker.cpp
dir / source / main.ext
*/
type Judge struct {
	dir  string
	tl   float64
	lang Lang
}

func NewJudge(lang string, checker, source io.Reader, tl float64) (*Judge, error) {
	judge := new(Judge)
	judge.lang = langs[lang]
	judge.tl = tl

	tempdir, err := ioutil.TempDir("", "hello")
	if err != nil {
		return nil, err
	}
	judge.dir = tempdir

	if err = os.Mkdir(path.Join(tempdir, "checker"), 0755); err != nil {
		return nil, err
	}
	if err = os.Chmod(path.Join(tempdir, "checker"), 0777); err != nil {
		return nil, err
	}
	if err = os.Mkdir(path.Join(tempdir, "source"), 0755); err != nil {
		return nil, err
	}
	if err = os.Chmod(path.Join(tempdir, "source"), 0777); err != nil {
		return nil, err
	}

	tempChecker, err := os.Create(path.Join(tempdir, "checker", "checker.cpp"))
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(tempChecker, checker); err != nil {
		return nil, err
	}

	testlib, err := os.Open("testlib.h")
	if err != nil {
		return nil, err
	}
	tempTestlib, err := os.Create(path.Join(tempdir, "checker", "testlib.h"))
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(tempTestlib, testlib); err != nil {
		return nil, err
	}
	tempSrc, err := os.Create(path.Join(tempdir, "source", judge.lang.Source))
	if err != nil {
		log.Print("error ", path.Join(tempdir, "source", judge.lang.Source))
		return nil, err
	}
	if _, err = io.Copy(tempSrc, source); err != nil {
		return nil, err
	}

	return judge, nil
}

func (j *Judge) CompileSource() (Result, error) {
	compile := strings.Fields(j.lang.Compile)
	cmd := exec.Command(compile[0], compile[1:]...)
	cmd.Dir = path.Join(j.dir, "source")
	cmd.Stdout = os.Stdout
	cmd.Stderr = nil
	return SafeRun(cmd, 30.0, false)
}

func (j *Judge) CompileChecker() (Result, error) {
	compile := strings.Fields(langs["checker"].Compile)
	cmd := exec.Command(compile[0], compile[1:]...)
	cmd.Dir = path.Join(j.dir, "checker")
	cmd.Stdout = os.Stdout
	cmd.Stderr = nil
	return SafeRun(cmd, 30.0, false)
}

type CaseResult struct {
	Status string
	Result
}

func AggregateResults(results []CaseResult) CaseResult {
	ans := CaseResult{"AC", Result{ReturnCode: -1, Time: -1, Memory: -1}}
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

func (j *Judge) TestCase(inFile io.Reader, expectFile io.Reader) (CaseResult, error) {
	input, err := os.Create(path.Join(j.dir, "checker", "input.in"))
	if err != nil {
		return CaseResult{}, err
	}
	if _, err = io.Copy(input, inFile); err != nil {
		return CaseResult{}, err
	}
	if _, err = input.Seek(0, 0); err != nil {
		return CaseResult{}, err
	}

	expect, err := os.Create(path.Join(j.dir, "checker", "expect.out"))
	if err != nil {
		return CaseResult{}, err
	}
	if _, err = io.Copy(expect, expectFile); err != nil {
		return CaseResult{}, err
	}
	if err = expect.Close(); err != nil {
		return CaseResult{}, err
	}

	actual, err := os.Create(path.Join(j.dir, "checker", "actual.out"))
	if err != nil {
		return CaseResult{}, err
	}

	arg := strings.Fields(j.lang.Exec)
	cmd := exec.Command(arg[0], arg[1:]...)
	cmd.Dir = path.Join(j.dir, "source")
	cmd.Stdin = input
	cmd.Stdout = actual
	result, err := SafeRun(cmd, j.tl, true)

	if err != nil {
		return CaseResult{}, err
	}

	if cmd.ProcessState.ExitCode() == 124 {
		//timeout
		return CaseResult{"TLE", result}, nil
	}

	if cmd.ProcessState.ExitCode() != 0 {
		return CaseResult{"Broken", result}, errors.New("executor return non 0, 124 code")
	}

	if result.ReturnCode != 0 {
		return CaseResult{"RE", result}, nil
	}
	actual.Close()

	// run checker
	cmd = exec.Command("./checker", "input.in", "actual.out", "expect.out")
	cmd.Dir = path.Join(j.dir, "checker")
	checkerResult, err := SafeRun(cmd, j.tl, true)
	if err != nil {
		return CaseResult{}, err
	}
	if cmd.ProcessState.ExitCode() == 124 {
		return CaseResult{"ITLE", result}, nil
	}
	if cmd.ProcessState.ExitCode() != 0 {
		return CaseResult{"Broken", result}, errors.New("executor return non 0, 124 code")
	}
	if checkerResult.ReturnCode == 1 {
		return CaseResult{"WA", result}, nil
	}
	if checkerResult.ReturnCode == 2 {
		return CaseResult{"PE", result}, nil
	}
	if checkerResult.ReturnCode == 3 {
		return CaseResult{"Fail", result}, nil
	}
	if checkerResult.ReturnCode != 0 {
		return CaseResult{"Unknown", result}, nil
	}
	return CaseResult{"AC", result}, nil
}
