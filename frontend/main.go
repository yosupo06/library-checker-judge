package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"

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

type Submittion struct {
	Id        int
	Problem   string
	Lang      string
	Status    string
	Source    string
	Maxtime   int
	Maxmemory int
}

type Task struct {
	Submittion int
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
	file, err := fileheader.Open()
	if err != nil {
		log.Fatal(err)
	}
	src, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	submittion := Submittion{}
	submittion.Problem = "unionfind"
	submittion.Lang = "cpp"
	submittion.Status = "WJ"
	submittion.Source = string(src)
	submittion.Maxtime = -1
	submittion.Maxmemory = -1
	db.Create(&submittion)

	task := Task{}
	task.Submittion = submittion.Id
	db.Create(&task)

	ctx.HTML(200, "submit.html", gin.H{})
	//	ctx.Redirect(http.StatusPermanentRedirect, "/submittions")
}

func submitList(ctx *gin.Context) {
	var submittions = make([]Submittion, 0)
	db.Find(&submittions)
	fmt.Println(submittions)
	ctx.HTML(200, "submitlist.html", gin.H{
		"Submittions": submittions,
	})
}

func main() {
	db = gormConnect()
	defer db.Close()
	db.AutoMigrate(Problem{})
	db.AutoMigrate(Submittion{})
	db.AutoMigrate(Task{})

	router := gin.Default()
	router.LoadHTMLGlob("templates/*.html")
	router.Static("/public", "./public")

	router.GET("/", problemList)
	router.GET("/problem/:name", problemInfo)
	router.POST("/submit", submit)
	router.GET("/submittions", submitList)

	router.Run(":8080")
}
