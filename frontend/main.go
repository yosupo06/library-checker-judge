package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	_ "github.com/lib/pq"
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

var db *gorm.DB

func gormConnect() *gorm.DB {
	host := getEnv("POSTGRE_HOST", "127.0.0.1")
	port := getEnv("POSTGRE_PORT", "5432")
	user := getEnv("POSTGRE_USER", "postgres")
	pass := getEnv("POSTGRE_PASS", "passwd")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=librarychecker password=%s sslmode=disable",
		host, port, user, pass)

	db, err := gorm.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

type Problem struct {
	Name      string
	Statement template.HTML
}

type Submission struct {
	Id        int
	Problem   string
	Lang      string
	Status    string
	Source    string
	Maxtime   int
	Maxmemory int
}

type Task struct {
	Submission int
}

func problemList(ctx *gin.Context) {
	var problems = make([]Problem, 0)
	db.Find(&problems)
	ctx.HTML(200, "problemlist.html", gin.H{
		"problems": problems,
	})
}

func problemInfo(ctx *gin.Context) {
	name := ctx.Param("name")
	var problem Problem
	db.Where("name = ?", name).First(&problem)
	ctx.HTML(200, "problem.html", gin.H{
		"Problem": problem,
	})
}

func submit(ctx *gin.Context) {
	fileheader, err := ctx.FormFile("source")
	if err != nil {
		log.Fatal(err)
	}
	problem := ctx.PostForm("problem")
	file, err := fileheader.Open()
	if err != nil {
		log.Fatal(err)
	}
	src, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	submission := Submission{}
	submission.Problem = problem
	submission.Lang = "cpp"
	submission.Status = "WJ"
	submission.Source = string(src)
	submission.Maxtime = -1
	submission.Maxmemory = -1
	db.Create(&submission)

	task := Task{}
	task.Submission = submission.Id
	db.Create(&task)

	ctx.Redirect(http.StatusFound, "/submission/"+strconv.Itoa(submission.Id))
}

func submissionInfo(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Fatal(err)
	}
	var submission Submission
	db.Where("id = ?", id).First(&submission)
	ctx.HTML(200, "submitinfo.html", gin.H{
		"Submission": submission,
	})
}

func submitList(ctx *gin.Context) {
	var submissions = make([]Submission, 0)
	db.Order("id desc").Find(&submissions)
	ctx.HTML(200, "submitlist.html", gin.H{
		"Submissions": submissions,
	})
}

func main() {
	db = gormConnect()
	defer db.Close()
	db.AutoMigrate(Problem{})
	db.AutoMigrate(Submission{})
	db.AutoMigrate(Task{})

	router := gin.Default()
	router.LoadHTMLGlob("templates/*.html")
	router.Static("/public", "./public")

	router.GET("/", problemList)
	router.GET("/problem/:name", problemInfo)
	router.POST("/submit", submit)
	router.GET("/submission/:id", submissionInfo)
	router.GET("/submissions", submitList)

	router.Run(":8080")
}
