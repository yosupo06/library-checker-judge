package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"sort"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/lib/pq"
	"gorm.io/gorm"

	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"github.com/yosupo06/library-checker-judge/database"
)

func FetchUserStatistics(db *gorm.DB, userName string) (map[string]pb.SolvedStatus, error) {
	type Result struct {
		ProblemName string
		LatestAC    bool
	}
	var results = make([]Result, 0)
	if err := db.
		Model(&database.Submission{}).
		Joins("left join problems on submissions.problem_name = problems.name").
		Select("problem_name, bool_or(submissions.test_cases_version=problems.test_cases_version) as latest_ac").
		Where("status = 'AC' and user_name = ?", userName).
		Group("problem_name").
		Find(&results).Error; err != nil {
		log.Print(err)
		return nil, errors.New("failed sql query")
	}
	stats := make(map[string]pb.SolvedStatus)
	for _, result := range results {
		if result.LatestAC {
			stats[result.ProblemName] = pb.SolvedStatus_LATEST_AC
		} else {
			stats[result.ProblemName] = pb.SolvedStatus_AC
		}
	}
	return stats, nil
}

func (s *server) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	token, err := s.authTokenManager.Register(s.db, in.Name, in.Password)
	if err != nil {
		return nil, err
	}
	return &pb.RegisterResponse{
		Token: token,
	}, nil
}

func (s *server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, err := s.authTokenManager.Login(s.db, in.Name, in.Password)
	if err != nil {
		return nil, err
	}
	return &pb.LoginResponse{
		Token: token,
	}, nil
}

func (s *server) UserInfo(ctx context.Context, in *pb.UserInfoRequest) (*pb.UserInfoResponse, error) {
	name := ""
	currentUserName := getCurrentUserName(ctx)
	myName := currentUserName
	if in.Name != "" {
		name = in.Name
	} else {
		name = myName
	}
	if name == "" {
		return nil, errors.New("empty name")
	}
	user, err := database.FetchUser(s.db, name)
	if user == nil || err != nil {
		return nil, errors.New("invalid user name")
	}
	stats, err := FetchUserStatistics(s.db, name)
	if err != nil {
		return nil, errors.New("failed to fetch statistics")
	}
	respUser := &pb.User{
		Name:        name,
		IsAdmin:     user.Admin,
		LibraryUrl:  user.LibraryURL,
		IsDeveloper: user.IsDeveloper,
	}

	resp := &pb.UserInfoResponse{
		IsAdmin: user.Admin,
		User:    respUser,
	}
	resp.SolvedMap = make(map[string]pb.SolvedStatus)
	for key, value := range stats {
		resp.SolvedMap[key] = value
	}
	return resp, nil
}

func (s *server) UserList(ctx context.Context, in *pb.UserListRequest) (*pb.UserListResponse, error) {
	currentUserName := getCurrentUserName(ctx)
	currentUser, _ := database.FetchUser(s.db, currentUserName)
	if currentUser.Name == "" {
		return nil, errors.New("not login")
	}
	if currentUser == nil || !currentUser.Admin {
		return nil, errors.New("must be admin")
	}
	users := []database.User{}
	if err := s.db.Select("name, admin").Find(&users).Error; err != nil {
		return nil, errors.New("failed to get users")
	}
	res := &pb.UserListResponse{}
	for _, user := range users {
		res.Users = append(res.Users, &pb.User{
			Name:    user.Name,
			IsAdmin: user.Admin,
		})
	}
	return res, nil
}

func (s *server) ChangeUserInfo(ctx context.Context, in *pb.ChangeUserInfoRequest) (*pb.ChangeUserInfoResponse, error) {
	type NewUserInfo struct {
		Email      string `validate:"omitempty,email,lt=50"`
		LibraryURL string `validate:"omitempty,url,lt=200"`
	}
	name := in.User.Name
	currentUserName := getCurrentUserName(ctx)
	currentUser, _ := database.FetchUser(s.db, currentUserName)

	if currentUser == nil || currentUser.Name == "" {
		return nil, errors.New("not login")
	}
	if name == "" {
		return nil, errors.New("requested name is empty")
	}
	if name != currentUser.Name && !currentUser.Admin {
		return nil, errors.New("permission denied")
	}
	if name == currentUser.Name && currentUser.Admin && !in.User.IsAdmin {
		return nil, errors.New("cannot remove myself from admin group")
	}

	userInfo := &NewUserInfo{
		LibraryURL: in.User.LibraryUrl,
	}
	if err := validator.New().Struct(userInfo); err != nil {
		return nil, err
	}

	if err := database.UpdateUser(s.db, database.User{
		Name:        in.User.Name,
		Admin:       in.User.IsAdmin,
		LibraryURL:  userInfo.LibraryURL,
		IsDeveloper: in.User.IsDeveloper,
	}); err != nil {
		return nil, err
	}

	return &pb.ChangeUserInfoResponse{}, nil
}

func (s *server) ProblemInfo(ctx context.Context, in *pb.ProblemInfoRequest) (*pb.ProblemInfoResponse, error) {
	p, err := database.FetchProblem(s.db, in.Name)

	if err != nil {
		return nil, err
	}

	return toProtoProblemInfo(p), nil
}

func (s *server) ProblemList(ctx context.Context, in *pb.ProblemListRequest) (*pb.ProblemListResponse, error) {
	problems := []database.Problem{}
	if err := s.db.Select("name, title").Find(&problems).Error; err != nil {
		return nil, errors.New("fetch problems failed")
	}

	res := pb.ProblemListResponse{}
	for _, prob := range problems {
		res.Problems = append(res.Problems, &pb.Problem{
			Name:  prob.Name,
			Title: prob.Title,
		})
	}
	return &res, nil
}

func (s *server) Submit(ctx context.Context, in *pb.SubmitRequest) (*pb.SubmitResponse, error) {
	if in.Source == "" {
		return nil, errors.New("empty Source")
	}
	if len(in.Source) > 1024*1024 {
		return nil, errors.New("too large Source")
	}
	ok := false
	for _, lang := range s.langs {
		if lang.Id == in.Lang {
			ok = true
			break
		}
	}
	if !ok {
		return nil, errors.New("unknown Lang")
	}
	if _, err := s.ProblemInfo(ctx, &pb.ProblemInfoRequest{
		Name: in.Problem,
	}); err != nil {
		log.Print(err)
		return nil, errors.New("unknown problem")
	}
	currentUserName := getCurrentUserName(ctx)
	currentUser, _ := database.FetchUser(s.db, currentUserName)

	name := ""
	if currentUser != nil {
		name = currentUser.Name
	}
	submission := database.Submission{
		SubmissionTime: time.Now(),
		ProblemName:    in.Problem,
		Lang:           in.Lang,
		Status:         "WJ",
		Source:         in.Source,
		MaxTime:        -1,
		MaxMemory:      -1,
		UserName:       sql.NullString{String: name, Valid: name != ""},
	}

	id, err := database.SaveSubmission(s.db, submission)
	if err != nil {
		log.Print(err)
		return nil, errors.New("Submit failed")
	}

	log.Println("Submit ", id)

	if err := s.pushTask(ctx, id, 50); err != nil {
		log.Print(err)
		return nil, errors.New("inserting to judge queue is failed")
	}

	return &pb.SubmitResponse{Id: id}, nil
}

func canRejudge(currentUser database.User, submission *pb.SubmissionOverview) bool {
	name := currentUser.Name
	if name == "" {
		return false
	}
	if name == submission.UserName {
		return true
	}
	if !submission.IsLatest && submission.Status == "AC" {
		return true
	}
	if currentUser.Admin {
		return true
	}
	return false
}

func (s *server) SubmissionInfo(ctx context.Context, in *pb.SubmissionInfoRequest) (*pb.SubmissionInfoResponse, error) {
	currentUserName := getCurrentUserName(ctx)
	currentUser, _ := database.FetchUser(s.db, currentUserName)

	var sub database.Submission
	sub, err := database.FetchSubmission(s.db, in.Id)
	if err != nil {
		log.Println("failed to fetch submission:", err)
		return nil, errors.New("failed to fetch submission")
	}
	cases, err := database.FetchTestcaseResults(s.db, in.Id)
	if err != nil {
		log.Println("failed to fetch submission results:", err)
		return nil, errors.New("failed to fetch submission results")
	}

	overview, err := toProtoSubmission(&sub)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	rej := false

	if currentUser != nil {
		rej = canRejudge(*currentUser, overview)
	}

	res := &pb.SubmissionInfoResponse{
		Overview:     overview,
		Source:       sub.Source,
		CompileError: sub.CompileError,
		CanRejudge:   rej,
	}

	sort.Slice(cases, func(i, j int) bool {
		return cases[i].Testcase < cases[j].Testcase
	})

	for _, c := range cases {
		res.CaseResults = append(res.CaseResults, &pb.SubmissionCaseResult{
			Case:       c.Testcase,
			Status:     c.Status,
			Time:       float64(c.Time) / 1000.0,
			Memory:     int64(c.Memory),
			Stderr:     c.Stderr,
			CheckerOut: c.CheckerOut,
		})
	}
	return res, nil
}

func (s *server) pushTask(ctx context.Context, subID, priority int32) error {
	if err := database.PushTask(s.db, subID, priority); err != nil {
		return err
	}
	return nil
}

func (s *server) SubmissionList(ctx context.Context, in *pb.SubmissionListRequest) (*pb.SubmissionListResponse, error) {
	if 1000 < in.Limit {
		in.Limit = 1000
	}

	filter := &database.Submission{
		ProblemName: in.Problem,
		Status:      in.Status,
		Lang:        in.Lang,
		UserName:    sql.NullString{String: in.User, Valid: (in.User != "")},
		Hacked:      in.Hacked,
	}

	count := int64(0)
	if err := s.db.Model(&database.Submission{}).Where(filter).Count(&count).Error; err != nil {
		return nil, errors.New("count query failed")
	}
	order := ""
	if in.Order == "" || in.Order == "-id" {
		order = "id desc"
	} else if in.Order == "+time" {
		order = "max_time asc"
	} else {
		return nil, errors.New("unknown sort order")
	}

	var submissions = make([]database.Submission, 0)
	if err := s.db.Where(filter).Limit(int(in.Limit)).Offset(int(in.Skip)).
		Preload("User").
		Preload("Problem").
		Order(order).
		Find(&submissions).Error; err != nil {
		return nil, errors.New("select query failed")
	}

	res := pb.SubmissionListResponse{
		Count: int32(count),
	}
	for _, sub := range submissions {
		protoSub, err := toProtoSubmission(&sub)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		res.Submissions = append(res.Submissions, protoSub)
	}
	return &res, nil
}

func (s *server) Rejudge(ctx context.Context, in *pb.RejudgeRequest) (*pb.RejudgeResponse, error) {
	sub, err := s.SubmissionInfo(ctx, &pb.SubmissionInfoRequest{Id: in.Id})
	if err != nil {
		return nil, err
	}
	if !sub.CanRejudge {
		return nil, errors.New("no permission")
	}

	if err := s.pushTask(ctx, in.Id, 40); err != nil {
		log.Print("rejudge failed:", err)
		return nil, errors.New("rejudge failed")
	}
	return &pb.RejudgeResponse{}, nil
}

func (s *server) LangList(ctx context.Context, in *pb.LangListRequest) (*pb.LangListResponse, error) {
	return &pb.LangListResponse{Langs: s.langs}, nil
}

func (s *server) Ranking(ctx context.Context, in *pb.RankingRequest) (*pb.RankingResponse, error) {
	type Result struct {
		UserName string
		AcCount  int
	}
	var results = make([]Result, 0)
	if err := s.db.
		Model(&database.Submission{}).
		Select("user_name, count(distinct problem_name) as ac_count").
		Where("status = 'AC' and user_name is not null").
		Group("user_name").
		Find(&results).Error; err != nil {
		log.Print(err)
		return nil, errors.New("failed sql query")
	}
	stats := make([]*pb.UserStatistics, 0)
	for _, result := range results {
		stats = append(stats, &pb.UserStatistics{
			Name:  result.UserName,
			Count: int32(result.AcCount),
		})
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count != stats[j].Count {
			return stats[i].Count > stats[j].Count
		}
		return stats[i].Name < stats[j].Name
	})
	res := pb.RankingResponse{
		Statistics: stats,
	}
	return &res, nil
}

type Category struct {
	Title    string   `json:"title"`
	Problems []string `json:"problems"`
}

func (s *server) ProblemCategories(ctx context.Context, in *pb.ProblemCategoriesRequest) (*pb.ProblemCategoriesResponse, error) {
	data, err := database.FetchMetadata(s.db, "problem_categories")
	if err != nil {
		return nil, err
	}
	var categories []Category
	if json.Unmarshal([]byte(*data), &categories); err != nil {
		return nil, err
	}

	var result []*pb.ProblemCategory

	for _, c := range categories {
		result = append(result, &pb.ProblemCategory{
			Title:    c.Title,
			Problems: c.Problems,
		})
	}
	return &pb.ProblemCategoriesResponse{
		Categories: result,
	}, nil
}
