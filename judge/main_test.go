package main

import (
	"strings"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

func Submit(db *gorm.DB, problem, lang string, srcFile io.Reader) (int, error) {
	src, err := ioutil.ReadAll(srcFile)
	if err != nil {
		return -1, err
	}
	submission := Submission{
		ProblemName: problem,
		Lang:        lang,
		Status:      "WJ",
		Source:      string(src),
		MaxTime:     -1,
		MaxMemory:   -1,
		UserName:    "tester",
	}
	if err = db.Create(&submission).Error; err != nil {
		return -1, err
	}
	return submission.ID, nil
}

var db *gorm.DB

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

	db = gormConnect()
	db.AutoMigrate(Problem{})
	db.AutoMigrate(Submission{})
	db.AutoMigrate(Task{})
	//db.LogMode(true)
	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()
	os.Exit(m.Run())
}

func TestSubmitAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac.cpp")
	if err != nil {
		t.Fatal(err)
	}
	id, err := Submit(db, "aplusb", "cpp", src)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("submit ok ", id)
	err = execJudge(db, Task{id})
	if err != nil {
		t.Fatal(err)
	}
	var submission Submission
	if err = db.
		Preload("Problem", func(db *gorm.DB) *gorm.DB {
			return db.Select("name, title, testhash")
		}).
		Where("id = ?", id).Take(&submission).Error; err != nil {			
		t.Fatal(err)
	}
	if submission.Status != "AC" {
		t.Fatal("Expect status AC, actual ", submission.Status)
	}
	if !(1 <= submission.MaxTime && submission.MaxTime <= 100) {
		t.Fatal("Irregural consume time ", submission.MaxTime)
	}
	if !(1 <= submission.MaxMemory && submission.MaxMemory <= 10_000_000) {
		t.Fatal("Irregural consume memory ", submission.MaxTime)
	}
	if submission.Testhash == "" {
		t.Fatal("You forgot to set testhash")
	}
}

func TestSubmitWA(t *testing.T) {
	src, err := os.Open("test_src/aplusb/wa.cpp")
	if err != nil {
		t.Fatal(err)
	}
	id, err := Submit(db, "aplusb", "cpp", src)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("submit ok ", id)
	err = execJudge(db, Task{id})
	if err != nil {
		t.Fatal(err)
	}
	var submission Submission
	if err = db.
		Preload("Problem", func(db *gorm.DB) *gorm.DB {
			return db.Select("name, title, testhash")
		}).
		Where("id = ?", id).Take(&submission).Error; err != nil {			
		t.Fatal(err)
	}
	if submission.Status != "WA" {
		t.Fatal("Expect status WA, actual ", submission.Status)
	}
	if !(1 <= submission.MaxTime && submission.MaxTime <= 100) {
		t.Fatal("Irregural consume time ", submission.MaxTime)
	}
}

func TestSubmitTLE(t *testing.T) {
	src, err := os.Open("test_src/TLE.cpp")
	if err != nil {
		t.Fatal(err)
	}
	id, err := Submit(db, "aplusb", "cpp", src)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("submit ok ", id)
	err = execJudge(db, Task{id})
	if err != nil {
		t.Fatal(err)
	}
	var submission Submission
	if err = db.
		Preload("Problem", func(db *gorm.DB) *gorm.DB {
			return db.Select("name, title, testhash")
		}).
		Where("id = ?", id).Take(&submission).Error; err != nil {			
		t.Fatal(err)
	}
	if submission.Status != "TLE" {
		t.Fatal("Expect status WA, actual ", submission.Status)
	}
	if !(1900 <= submission.MaxTime && submission.MaxTime <= 2100) {
		t.Fatal("Irregural consume time ", submission.MaxTime)
	}
}

func TestSubmitRE(t *testing.T) {
	src, err := os.Open("test_src/aplusb/re.cpp")
	if err != nil {
		t.Fatal(err)
	}
	id, err := Submit(db, "aplusb", "cpp", src)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("submit ok ", id)
	err = execJudge(db, Task{id})
	if err != nil {
		t.Fatal(err)
	}
	var submission Submission
	if err = db.
		Preload("Problem", func(db *gorm.DB) *gorm.DB {
			return db.Select("name, title, testhash")
		}).
		Where("id = ?", id).Take(&submission).Error; err != nil {			
		t.Fatal(err)
	}
	if submission.Status != "RE" {
		t.Fatal("Expect status WA, actual ", submission.Status)
	}
	if !(1 <= submission.MaxTime && submission.MaxTime <= 100) {
		t.Fatal("Irregural consume time ", submission.MaxTime)
	}
}

func TestSubmitCE(t *testing.T) {
	id, err := Submit(db, "aplusb", "cpp", strings.NewReader("The answer is 42..."))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("submit ok ", id)
	err = execJudge(db, Task{id})
	if err != nil {
		t.Fatal(err)
	}
	var submission Submission
	if err = db.
		Preload("Problem", func(db *gorm.DB) *gorm.DB {
			return db.Select("name, title, testhash")
		}).
		Where("id = ?", id).Take(&submission).Error; err != nil {			
		t.Fatal(err)
	}
	if submission.Status != "CE" {
		t.Fatal("Expect status CE, actual ", submission.Status)
	}
}



func TestSubmitRustAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac.rs")
	if err != nil {
		t.Fatal(err)
	}
	id, err := Submit(db, "aplusb", "rust", src)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("submit ok ", id)
	err = execJudge(db, Task{id})
	if err != nil {
		t.Fatal(err)
	}
	var submission Submission
	if err = db.
		Preload("Problem", func(db *gorm.DB) *gorm.DB {
			return db.Select("name, title, testhash")
		}).
		Where("id = ?", id).Take(&submission).Error; err != nil {			
		t.Fatal(err)
	}
	if submission.Status != "AC" {
		t.Fatal("Expect status AC, actual ", submission.Status)
	}
	if !(1 <= submission.MaxTime && submission.MaxTime <= 100) {
		t.Fatal("Irregural consume time ", submission.MaxTime)
	}
	if !(1 <= submission.MaxMemory && submission.MaxMemory <= 10_000_000) {
		t.Fatal("Irregural consume memory ", submission.MaxTime)
	}
}

func TestSubmitPythonAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac.py")
	if err != nil {
		t.Fatal(err)
	}
	id, err := Submit(db, "aplusb", "pypy3", src)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("submit ok ", id)
	err = execJudge(db, Task{id})
	if err != nil {
		t.Fatal(err)
	}
	var submission Submission
	if err = db.
		Preload("Problem", func(db *gorm.DB) *gorm.DB {
			return db.Select("name, title, testhash")
		}).
		Where("id = ?", id).Take(&submission).Error; err != nil {			
		t.Fatal(err)
	}
	if submission.Status != "AC" {
		t.Fatal("Expect status AC, actual ", submission.Status)
	}
	if !(1 <= submission.MaxTime && submission.MaxTime <= 100) {
		t.Fatal("Irregural consume time ", submission.MaxTime)
	}
	if !(1 <= submission.MaxMemory && submission.MaxMemory <= 10_000_000) {
		t.Fatal("Irregural consume memory ", submission.MaxTime)
	}
}

func TestSubmitDlangAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac.d")
	if err != nil {
		t.Fatal(err)
	}
	id, err := Submit(db, "aplusb", "d", src)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("submit ok ", id)
	err = execJudge(db, Task{id})
	if err != nil {
		t.Fatal(err)
	}
	var submission Submission
	if err = db.
		Preload("Problem", func(db *gorm.DB) *gorm.DB {
			return db.Select("name, title, testhash")
		}).
		Where("id = ?", id).Take(&submission).Error; err != nil {			
		t.Fatal(err)
	}
	if submission.Status != "AC" {
		t.Fatal("Expect status AC, actual ", submission.Status)
	}
	if !(1 <= submission.MaxTime && submission.MaxTime <= 100) {
		t.Fatal("Irregural consume time ", submission.MaxTime)
	}
	if !(1 <= submission.MaxMemory && submission.MaxMemory <= 10_000_000) {
		t.Fatal("Irregural consume memory ", submission.MaxTime)
	}
}

func TestSubmitJavaAC(t *testing.T) {
	src, err := os.Open("test_src/aplusb/ac.java")
	if err != nil {
		t.Fatal(err)
	}
	id, err := Submit(db, "aplusb", "java", src)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("submit ok ", id)
	err = execJudge(db, Task{id})
	if err != nil {
		t.Fatal(err)
	}
	var submission Submission
	if err = db.
		Preload("Problem", func(db *gorm.DB) *gorm.DB {
			return db.Select("name, title, testhash")
		}).
		Where("id = ?", id).Take(&submission).Error; err != nil {			
		t.Fatal(err)
	}
	if submission.Status != "AC" {
		t.Fatal("Expect status AC, actual ", submission.Status)
	}
}
