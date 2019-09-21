package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

func login(c *gin.Context, name string, password string) bool {
	var user User
	if err := db.Where("name = ?", name).First(&user).Error; err != nil {
		return false
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Passhash), []byte(password)); err != nil {
		return false
	}
	session := sessions.Default(c)
	session.Set("user", user)
	session.Save()
	return true
}

func logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
}

func getUser(c *gin.Context) User {
	session := sessions.Default(c)
	user, ok := session.Get("user").(User)
	if !ok {
		return User{}
	}
	return user
}

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

func htmlWithUser(c *gin.Context, code int, name string, obj gin.H) {
	obj["User"] = getUser(c)
	c.HTML(code, name, obj)
}

func problemList(ctx *gin.Context) {
	var problems = make([]Problem, 0)
	db.Select("name, title").Find(&problems)
	var titlemap = make(map[string]string)
	for _, problem := range problems {
		titlemap[problem.Name] = problem.Title
	}
	htmlWithUser(ctx, 200, "problemlist.html", gin.H{
		"titlemap": titlemap,
		"list":     list,
	})
}

func problemInfo(ctx *gin.Context) {
	name := ctx.Param("name")
	var problem Problem
	db.Select("name, title, statement, timelimit").Where("name = ?", name).First(&problem)
	htmlWithUser(ctx, 200, "problem.html", gin.H{
		"User":    getUser(ctx),
		"Problem": problem,
	})
}

func checkLang(lang string) bool {
	langs := []string{"cpp", "rust", "d", "java"}
	for _, s := range langs {
		if lang == s {
			return true
		}
	}
	return false
}

func submit(ctx *gin.Context) {
	type SubmitForm struct {
		Source  *multipart.FileHeader `form:"source" binding:"required"`
		Problem string                `form:"problem" binding:"required"`
		Lang    string                `form:"lang" binding:"required"`
	}
	var submitForm SubmitForm
	if err := ctx.ShouldBind(&submitForm); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	file, err := submitForm.Source.Open()
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	src, err := ioutil.ReadAll(file)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if !checkLang(submitForm.Lang) {
		ctx.Abort()
		return
	}
	submission := Submission{
		ProblemName: submitForm.Problem,
		Lang:        submitForm.Lang,
		Status:      "WJ",
		Source:      string(src),
		MaxTime:     -1,
		MaxMemory:   -1,
		UserName:    getUser(ctx).getName(),
	}
	db.Create(&submission)

	task := Task{}
	task.Submission = submission.ID
	db.Create(&task)

	ctx.Redirect(http.StatusFound, "/submission/"+strconv.Itoa(submission.ID))
}

func submissionInfo(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Fatal(err)
	}
	var submission Submission
	db.
		Preload("Problem", func(db *gorm.DB) *gorm.DB {
			return db.Select("name, title, testhash")
		}).
		Where("id = ?", id).First(&submission)
	var results []SubmissionTestcaseResult
	db.Where("submission = ?", id).Find(&results)
	rejudge := adminOrSubmitter(ctx, submission)
	htmlWithUser(ctx, 200, "submitinfo.html", gin.H{
		"Submission": submission,
		"Results":    results,
		"Rejudge":    rejudge,
	})
}

func submitList(ctx *gin.Context) {
	type SubmitFilter struct {
		Page    int    `form:"page,default=1" binding:"gte=1,lte=1000"`
		Problem string `form:"problem" binding:"lte=100"`
		Status  string `form:"status" binding:"lte=100"`
	}
	var submitFilter SubmitFilter
	if err := ctx.ShouldBind(&submitFilter); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	var submissions = make([]Submission, 0)
	db.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("name")
		}).
		Preload("Problem", func(db *gorm.DB) *gorm.DB {
			return db.Select("name, title, testhash")
		}).
		Limit(100).
		Offset((submitFilter.Page - 1) * 100).
		Order("id desc").
		Select("id, user_name, problem_name, lang, status, testhash, max_time, max_memory").
		Where(&Submission{ProblemName: submitFilter.Problem, Status: submitFilter.Status}).
		//		Where("problem_name = ?", submitFilter.Problem).
		Find(&submissions)
	count := 0
	db.Table("submissions").Count(&count)
	htmlWithUser(ctx, 200, "submitlist.html", gin.H{
		"Submissions": submissions,
		"NowPage":     submitFilter.Page,
		"NumPage":     (count + 99) / 100,
	})
}

func registerGet(ctx *gin.Context) {
	htmlWithUser(ctx, 200, "register.html", gin.H{
		"Name": "",
	})
}

func registerPost(ctx *gin.Context) {
	type UserPass struct {
		Name     string `form:"name" binding:"required,alphanum,gte=3,lte=60"`
		Password string `form:"password" binding:"required,printascii,gte=8,lte=72"`
		Confirm  string `form:"confirm" binding:"eqfield=Password"`
	}
	var userPass UserPass
	if err := ctx.ShouldBind(&userPass); err != nil {
		htmlWithUser(ctx, 200, "register.html", gin.H{
			"Name":  userPass.Name,
			"Error": err.Error(),
		})
		return
	}
	passHash, err := bcrypt.GenerateFromPassword([]byte(userPass.Password), 10)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	user := User{
		Name:     userPass.Name,
		Passhash: string(passHash),
	}
	if err := db.Create(&user).Error; err != nil {
		htmlWithUser(ctx, 200, "register.html", gin.H{
			"Error": "This username are already registered",
		})
	}
	login(ctx, userPass.Name, userPass.Password)
	ctx.Redirect(http.StatusFound, "/")
}

func loginGet(ctx *gin.Context) {
	htmlWithUser(ctx, 200, "login.html", gin.H{
		"Name": "",
	})
}

func loginPost(ctx *gin.Context) {
	type UserPass struct {
		Name     string `form:"name" binding:"required,alphanum,gte=3,lte=60"`
		Password string `form:"password" binding:"required,printascii,gte=8,lte=72"`
	}
	var userPass UserPass
	if err := ctx.ShouldBind(&userPass); err != nil {
		htmlWithUser(ctx, 200, "login.html", gin.H{
			"Name":  userPass.Name,
			"Error": err.Error(),
		})
		return
	}
	if !login(ctx, userPass.Name, userPass.Password) {
		htmlWithUser(ctx, 200, "login.html", gin.H{
			"Name": userPass.Name,
		})
		return
	}
	ctx.Redirect(http.StatusFound, "/")
}

func logoutGet(ctx *gin.Context) {
	logout(ctx)

	ctx.Redirect(http.StatusFound, "/")
}

func adminOrSubmitter(ctx *gin.Context, submission Submission) bool {
	if getUser(ctx).Admin {
		return true
	}
	if !submission.UserName.Valid || submission.UserName.String == "" {
		return false
	}
	return submission.UserName.String == getUser(ctx).Name
}

func rejudge(id int) error {
	tx := db.Begin()
	var submission Submission
	err := tx.Where("id = ?", id).First(&submission).Error
	if err != nil {
		tx.Rollback()
		return errors.New("can not find submissions")
	}
	if time.Now().Sub(submission.JudgePing).Minutes() <= 1 {
		tx.Rollback()
		return errors.New("this submission seems judging now")
	}
	submission.Status = "WJ"
	tx.Save(&submission)
	tx.Commit()

	task := Task{}
	task.Submission = id
	db.Create(&task)
	return nil
}

func getRejudge(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "param error",
		})
		return
	}
	var submission Submission
	err = db.Where("id = ?", id).First(&submission).Error
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "can not find submissions",
		})
		return
	}
	if !adminOrSubmitter(ctx, submission) {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "no authority",
		})
		return
	}
	err = rejudge(id)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	ctx.Redirect(http.StatusFound, fmt.Sprintf("/submission/%d", id))
}

func allRejudge(ctx *gin.Context) {
	var submissions = make([]Submission, 0)
	if !getUser(ctx).Admin {
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}
	db.
		Order("id desc").
		Select("id").
		Find(&submissions)
	for _, s := range submissions {
		err := rejudge(s.ID)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

func main() {
	loadList()
	gob.Register(User{})
	db = gormConnect()
	defer db.Close()
	db.AutoMigrate(Problem{})
	db.AutoMigrate(Submission{})
	db.AutoMigrate(Task{})
	db.AutoMigrate(User{})
	// db.LogMode(true)

	router := gin.Default()
	router.SetFuncMap(template.FuncMap{
		"repeat": func(a, b int) []int {
			var result []int
			for i := a; i <= b; i++ {
				result = append(result, i)
			}
			return result
		},
	})
	router.Use(sessions.Sessions("mysession",
		cookie.NewStore([]byte(getEnv("SESSION_SECRET", "session_secret")))))
	router.LoadHTMLGlob("templates/*.html")
	router.Static("/public", "./public")

	router.GET("/register", registerGet)
	router.POST("/register", registerPost)

	router.GET("/login", loginGet)
	router.POST("/login", loginPost)
	router.GET("/logout", logoutGet)

	router.GET("/", problemList)
	router.GET("/problem/:name", problemInfo)
	router.POST("/submit", submit)
	router.GET("/submission/:id", submissionInfo)
	router.GET("/submissions", submitList)

	router.GET("/rejudge/:id", getRejudge)
	router.GET("/admin/allrejudge", allRejudge)

	router.Run(":8080")
}
