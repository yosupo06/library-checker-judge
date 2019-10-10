package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"github.com/BurntSushi/toml"
	_ "github.com/lib/pq"
)

type Result struct {
	ReturnCode int     `json:"returncode"`
	Time       float64 `json:"time"`
	Memory     int     `json:"memory"`
}

type Execute struct {
	exec.Cmd
	tl      float64
	overlay bool
}

func SafeRun(cmd *exec.Cmd, tl float64, overlay bool) (Result, error) {
	newCmd := []string{cmd.Path}
	newCmd = append(newCmd, cmd.Args...)
	newCmd = append(newCmd, "--tl", strconv.FormatFloat(tl, 'f', 4, 64))
	if overlay {
		newCmd = append(newCmd, "--overlay")
	}
	tmpfile, err := ioutil.TempFile("", "result")
	if err != nil {
		return Result{}, err
	}
	newCmd = append(newCmd, "--result", tmpfile.Name())
	cmd.Path = "./v2/executor.py"
	cmd.Args = newCmd
	err = cmd.Run()
	if err != nil {
		return Result{ReturnCode: -1, Time: -1, Memory: -1}, errors.New("Fail Tmpfile")
	}
	raw, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		return Result{}, err
	}
	result := Result{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return Result{}, err
	}
	return result, nil
}

// func ExecuteStr(cmd, dir string, tl float64, overlay bool, stdin, stdout, stderr io.Reader) (Result, error) {
// 	return Execute(strings.Fields(cmd), dir, tl, overlay)
// }

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

type Judge struct {
	dir       string
	tl        float64
	lang      Lang
	CaseNames []string
}

func NewJudge(dir, lang string) *Judge {
	judge := new(Judge)
	judge.dir = dir
	judge.lang = langs[lang]
	return judge
}

func (j *Judge) CompileSource() (Result, error) {
	compile := strings.Fields(j.lang.Compile)
	cmd := exec.Command(compile[0], compile[1:]...)
	cmd.Dir = j.dir	
	return SafeRun(cmd, 30.0, false)
}

func (j *Judge) CompileChecker() (Result, error) {
	compile := strings.Fields(langs["checker"].Compile)
	cmd := exec.Command(compile[0], compile[1:]...)
	cmd.Dir = j.dir	
	return SafeRun(cmd, 30.0, false)
}

func (j *Judge) TestCase(caseID int) (Result, error) {
	if caseID < 0 || len(j.CaseNames) <= caseID {
		return Result{}, errors.New("Invalid case ID")
	}
	cn := j.CaseNames[caseID]
	
	of := path.Join(j.dir, "out", cn)
	Execute(j.lang.Exec, j.dir, j.tl, true)
	return ExecuteStr(j.lang.Compile, j.dir, j.tl, false)
}
