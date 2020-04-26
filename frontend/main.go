package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/gin-gonic/gin"

	"crypto/tls"
	"crypto/x509"

	_ "github.com/lib/pq"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
)

var client pb.LibraryCheckerServiceClient

func langList(ctx *gin.Context) ([]*pb.Lang, error) {
	list, err := client.LangList(ctx, &pb.LangListRequest{})
	if err != nil {
		return nil, err
	}
	return list.Langs, nil
}

func getUserName(c *gin.Context) string {
	user := c.Value("user")
	if name, ok := user.(string); ok {
		return name
	}
	return ""
}

func htmlWithUser(c *gin.Context, code int, name string, obj gin.H) {
	obj["User"] = getUserName(c)
	c.HTML(code, name, obj)
}

func problemList(ctx *gin.Context) {
	res, err := client.ProblemList(ctx, &pb.ProblemListRequest{})
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	type ProblemInfo struct {
		Title  string
		Solved int
	}
	var titlemap = make(map[string]*ProblemInfo)
	for _, problem := range res.Problems {
		if _, ok := titlemap[problem.Name]; !ok {
			titlemap[problem.Name] = &ProblemInfo{}
		}
		titlemap[problem.Name].Title = problem.Title
	}

	userName := getUserName(ctx)
	if userName != "" {
		subs, err := client.SubmissionList(ctx, &pb.SubmissionListRequest{
			Skip:   0,
			Limit:  1000,
			Status: "AC",
			User:   userName,
		})

		if err == nil {
			for _, sub := range subs.Submissions {
				if sub.IsLatest {
					titlemap[sub.ProblemName].Solved = 2
				} else if titlemap[sub.ProblemName].Solved == 0 {
					titlemap[sub.ProblemName].Solved = 1
				}
			}
		}
	}

	htmlWithUser(ctx, 200, "problemlist.html", gin.H{
		"titlemap": titlemap,
		"list":     list,
	})
}

func problemInfo(ctx *gin.Context) {
	name := ctx.Param("name")
	problem, err := client.ProblemInfo(ctx, &pb.ProblemInfoRequest{Name: name})
	if err != nil {
		ctx.AbortWithError(http.StatusServiceUnavailable, err)
		return
	}
	Langs, err := langList(ctx)
	if err != nil {
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	htmlWithUser(ctx, 200, "problem.html", gin.H{
		"Name":      name,
		"Statement": template.HTML(problem.Statement),
		"Problem":   problem,
		"Langs":     Langs,
	})
}

func checkLang(ctx *gin.Context, lang string) (bool, error) {
	langs, err := langList(ctx)
	if err != nil {
		return false, err
	}
	for _, s := range langs {
		if lang == s.Id {
			return true, nil
		}
	}
	return false, nil
}

func submit(ctx *gin.Context) {
	type SubmitForm struct {
		SourceText string `form:"source_text"`
		Problem    string `form:"problem" binding:"required"`
		Lang       string `form:"lang" binding:"required"`
	}
	var submitForm SubmitForm
	if err := ctx.ShouldBind(&submitForm); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	src := submitForm.SourceText
	if src == "" {
		file, err := ctx.FormFile("source")
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}
		f, err := file.Open()
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}
		srcByte, err := ioutil.ReadAll(f)
		src = string(srcByte)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}
	response, err := client.Submit(ctx, &pb.SubmitRequest{
		Problem: submitForm.Problem,
		Source:  src,
		Lang:    submitForm.Lang,
	})

	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/submission/"+strconv.Itoa(int(response.Id)))
}

func submissionInfo(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	sub, err := client.SubmissionInfo(ctx, &pb.SubmissionInfoRequest{
		Id: int32(id),
	})
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	admin := false
	if getUserName(ctx) != "" {
		user, err := client.UserInfo(ctx, &pb.UserInfoRequest{})
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
		}
		admin = user.IsAdmin
	}
	htmlWithUser(ctx, 200, "submitinfo.html", gin.H{
		"Overview":   sub.Overview,
		"CanRejudge": sub.CanRejudge,
		"Results":    sub.CaseResults,
		"Source":     sub.Source,
		"Admin":      admin,
	})
}

func submitList(ctx *gin.Context) {
	type SubmitFilter struct {
		Page    int    `form:"page,default=1" binding:"gte=1,lte=1000"`
		Problem string `form:"problem" binding:"lte=100"`
		Status  string `form:"status" binding:"lte=100"`
		User    string `form:"user" binding:"lte=100"`
		Order   string `form:"order"`
	}
	var submitFilter SubmitFilter
	if err := ctx.ShouldBind(&submitFilter); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	res, err := client.SubmissionList(ctx, &pb.SubmissionListRequest{
		Skip:    uint32((submitFilter.Page - 1) * 100),
		Limit:   100,
		Problem: submitFilter.Problem,
		Status:  submitFilter.Status,
		User:    submitFilter.User,
		Order:   submitFilter.Order,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	list, err := client.LangList(ctx, &pb.LangListRequest{})
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	langMap := make(map[string]string)
	for _, v := range list.Langs {
		langMap[v.Id] = v.Name
	}
	numPage := int((res.Count + 99) / 100)
	pages := make([]int, 0)
	for i := -5; i <= 5; i++ {
		page := submitFilter.Page + i
		if 1 <= page && page <= numPage {
			pages = append(pages, page)
		}
	}

	htmlWithUser(ctx, 200, "submitlist.html", gin.H{
		"Submissions": res.Submissions,
		"Problem":     submitFilter.Problem,
		"Status":      submitFilter.Status,
		"Pages":       pages,
		"Order":       submitFilter.Order,
		"FilterUser":  submitFilter.User,
		"LangMap":     langMap,
		"NumPage":     numPage,
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
	res, err := client.Register(ctx, &pb.RegisterRequest{
		Name:     userPass.Name,
		Password: userPass.Password,
	})
	if err != nil {
		htmlWithUser(ctx, 200, "register.html", gin.H{
			"Error": "This username are already registered",
		})
		return
	}
	ctx.SetCookie("token", res.Token, 365*24*3600, "", "", false, false)
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
	response, err := client.Login(ctx, &pb.LoginRequest{
		Name:     userPass.Name,
		Password: userPass.Password,
	})

	if err != nil {
		htmlWithUser(ctx, 200, "login.html", gin.H{
			"Name":  userPass.Name,
			"Error": err.Error(),
		})
		return
	}

	ctx.SetCookie("token", response.Token, 365*24*3600, "", "", false, false)
	ctx.Redirect(http.StatusFound, "/")
}

func logoutGet(ctx *gin.Context) {
	ctx.SetCookie("token", "", -1, "", "", false, false)
	ctx.Redirect(http.StatusFound, "/")
}

func getRejudge(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "param error",
		})
		return
	}

	_, err = client.Rejudge(ctx, &pb.RejudgeRequest{
		Id: int32(id),
	})

	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	ctx.Redirect(http.StatusFound, fmt.Sprintf("/submission/%d", id))
}

func helpPage(ctx *gin.Context) {
	langs, err := langList(ctx)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	htmlWithUser(ctx, 200, "help.html", gin.H{
		"Langs": langs,
	})
}

func getRanking(ctx *gin.Context) {
	stats, err := client.Ranking(ctx, &pb.RankingRequest{})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	type UserInfo struct {
		Rank  int
		Name  string
		Count int32
	}
	userInfo := make([]UserInfo, 0)
	nowRank := 1
	back := int32(-1)
	for i, v := range stats.Statistics {
		if back != v.Count {
			nowRank = i + 1
			back = v.Count
		}
		userInfo = append(userInfo, UserInfo{
			Rank:  nowRank,
			Name:  v.Name,
			Count: v.Count,
		})
	}
	htmlWithUser(ctx, 200, "ranking.html", gin.H{
		"Statistics": userInfo,
	})
}

func getUserList(ctx *gin.Context) {
	users, err := client.UserList(ctx, &pb.UserListRequest{})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	htmlWithUser(ctx, 200, "userlist.html", gin.H{
		"Users": users.Users,
	})
}

func getAddAdmin(ctx *gin.Context) {
	name := ctx.Param("name")
	_, err := client.ChangeUserInfo(ctx, &pb.ChangeUserInfoRequest{
		User: &pb.User{
			Name:    name,
			IsAdmin: true,
		},
	})
	if err != nil {
		ctx.AbortWithError(http.StatusServiceUnavailable, err)
		return
	}
	ctx.Redirect(http.StatusFound, "/admin/userlist")
}

func getDelAdmin(ctx *gin.Context) {
	name := ctx.Param("name")
	_, err := client.ChangeUserInfo(ctx, &pb.ChangeUserInfoRequest{
		User: &pb.User{
			Name:    name,
			IsAdmin: false,
		},
	})
	if err != nil {
		ctx.AbortWithError(http.StatusServiceUnavailable, err)
		return
	}
	ctx.Redirect(http.StatusFound, "/admin/userlist")
}

type loginCreds struct{}

func (c *loginCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	dict := map[string]string{}
	if token, ok := ctx.Value("token").(string); ok && token != "" {
		dict["authorization"] = "bearer " + token
	}
	return dict, nil
}

func (c *loginCreds) RequireTransportSecurity() bool {
	return false
}

func grpcDial(local bool, host string) (*grpc.ClientConn, error) {
	options := []grpc.DialOption{grpc.WithBlock(), grpc.WithPerRPCCredentials(&loginCreds{})}
	if local {
		if host == "" {
			host = "localhost:50051"
		}
		options = append(options, grpc.WithInsecure())
	} else {
		if host == "" {
			host = "apiv1.yosupo.jp:443"
		}
		systemRoots, err := x509.SystemCertPool()
		if err != nil {
			log.Fatal(err)
		}
		creds := credentials.NewTLS(&tls.Config{
			RootCAs: systemRoots,
		})
		options = append(options, grpc.WithTransportCredentials(creds))
	}

	return grpc.Dial(host, options...)
}

func main() {
	local := flag.Bool("local", false, "API server is local")
	host := flag.String("host", "", "Hostname of API server")
	flag.Parse()
	conn, err := grpcDial(*local, *host)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client = pb.NewLibraryCheckerServiceClient(conn)
	loadList()

	router := gin.Default()
	router.Use(func(ctx *gin.Context) {
		token, err := ctx.Cookie("token")
		if err == nil {
			ctx.Set("token", token)
			parsed, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
			if err == nil {
				if claims, ok := parsed.Claims.(jwt.MapClaims); ok {
					if user, ok := claims["user"]; ok {
						if name, ok := user.(string); ok {
							ctx.Set("user", name)
						}
					}
				}
			}
		}
		ctx.Next()
	})
	router.SetFuncMap(template.FuncMap{
		"repeat": func(a, b int) []int {
			var result []int
			for i := a; i <= b; i++ {
				result = append(result, i)
			}
			return result
		},
		"time2str": func(a float64) string {
			msec := int(math.Round(a * 1000))
			return fmt.Sprintf("%d ms", msec)
		},
		"mem2str": func(a int64) string {
			if a == -1 {
				return "-1 Mib"
			}
			return fmt.Sprintf("%.2f MiB", float64(a)/1024/1024)
		},
	})
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

	router.GET("/ranking", getRanking)
	router.GET("/help", helpPage)

	router.GET("/admin/userlist", getUserList)
	router.GET("/admin/addadmin/:name", getAddAdmin)
	router.GET("/admin/deladmin/:name", getDelAdmin)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}
