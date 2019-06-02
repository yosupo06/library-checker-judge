package main

import (
	"fmt"
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
	Name string
}

func problemList(ctx *gin.Context) {
	var problems = make([]Problem, 0)
	db.Find(&problems)
	fmt.Println(problems)
	ctx.HTML(200, "problemlist.html", gin.H{
		"problems": problems,
	})

}

func problemInfo(ctx *gin.Context) {
	name := ctx.Param("name")
	ctx.HTML(200, "problem.html", gin.H{
		"Name": name,
	})
}

func submit(ctx *gin.Context) {
	file, err := ctx.FormFile("source")
	fmt.Println(file, err)
}

func main() {
	db = gormConnect()
	defer db.Close()
	db.AutoMigrate(Problem{})

	router := gin.Default()
	router.LoadHTMLGlob("templates/*.html")
	router.Static("/public", "./public")

	router.GET("/", problemList)
	router.GET("/problem/:name", problemInfo)
	router.POST("/submit", submit)

	router.Run(":8080")
}
