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

const (
	SUBMISSION_PRIORITY           = 50
	ANONYMOUS_SUBMISSION_PRIORITY = 45
	REJUDGE_PRIORITY              = 40
)

func (s *server) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	uid := s.currentUserUID(ctx)

	if uid == "" {
		return nil, errors.New("uid is empty")
	}

	if err := database.RegisterUser(s.db, in.Name, uid); err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{}, nil
}

func (s *server) CurrentUserInfo(ctx context.Context, in *pb.CurrentUserInfoRequest) (*pb.CurrentUserInfoResponse, error) {
	user := s.currentUser(ctx)

	return &pb.CurrentUserInfoResponse{
		User: toProtoUser(user),
	}, nil
}

func (s *server) ChangeCurrentUserInfo(ctx context.Context, in *pb.ChangeCurrentUserInfoRequest) (*pb.ChangeCurrentUserInfoResponse, error) {
	uid := s.currentUserUID(ctx)
	if uid == "" {
		return nil, errors.New("login required")
	}

	if err := database.UpdateUser(s.db, database.User{
		Name:        in.User.Name,
		UID:         uid,
		LibraryURL:  in.User.LibraryUrl,
		IsDeveloper: in.User.IsDeveloper,
	}); err != nil {
		return nil, err
	}

	return &pb.ChangeCurrentUserInfoResponse{}, nil
}

func (s *server) UserInfo(ctx context.Context, in *pb.UserInfoRequest) (*pb.UserInfoResponse, error) {
	name := in.Name
	if name == "" {
		return nil, errors.New("empty name")
	}
	user, err := database.FetchUserFromName(s.db, name)
	if user == nil || err != nil {
		return nil, errors.New("invalid user name")
	}
	stats, err := FetchUserStatistics(s.db, name)
	if err != nil {
		return nil, errors.New("failed to fetch statistics")
	}

	resp := &pb.UserInfoResponse{
		User: toProtoUser(user),
	}
	resp.SolvedMap = make(map[string]pb.SolvedStatus)
	for key, value := range stats {
		resp.SolvedMap[key] = value
	}
	return resp, nil
}

func (s *server) ChangeUserInfo(ctx context.Context, in *pb.ChangeUserInfoRequest) (*pb.ChangeUserInfoResponse, error) {
	type NewUserInfo struct {
		LibraryURL string `validate:"omitempty,url,lt=200"`
	}

	name := in.User.Name
	if name == "" {
		return nil, errors.New("requested name is empty")
	}

	currentUser := s.currentUser(ctx)
	if currentUser == nil {
		return nil, errors.New("not login")
	}
	if name != currentUser.Name {
		return nil, errors.New("permission denied")
	}

	userInfo := &NewUserInfo{
		LibraryURL: in.User.LibraryUrl,
	}
	if err := validator.New().Struct(userInfo); err != nil {
		return nil, err
	}

	if err := database.UpdateUser(s.db, database.User{
		Name:        in.User.Name,
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
		log.Println(err)
		return nil, errors.New("failed to fetch problem")
	}

	if p == nil {
		return nil, errors.New("problem not found")
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
	currentUserName := s.currentUserName(ctx)
	currentUser, _ := database.FetchUserFromName(s.db, currentUserName)

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

	priority := ANONYMOUS_SUBMISSION_PRIORITY
	if currentUser != nil {
		priority = SUBMISSION_PRIORITY
	}
	if err := s.pushTask(ctx, id, int32(priority)); err != nil {
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
	return false
}

func (s *server) SubmissionInfo(ctx context.Context, in *pb.SubmissionInfoRequest) (*pb.SubmissionInfoResponse, error) {
	currentUserName := s.currentUserName(ctx)
	currentUser, _ := database.FetchUserFromName(s.db, currentUserName)

	var sub *database.Submission
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

	overview := toProtoSubmissionOverview(database.ToSubmissionOverView(*sub))

	rej := false
	if currentUser != nil {
		rej = canRejudge(*currentUser, &overview)
	}

	res := &pb.SubmissionInfoResponse{
		Overview:     &overview,
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

func (s *server) pushTask(_ context.Context, subID, priority int32) error {
	if err := database.PushTask(s.db, database.TaskData{
		TaskType:   database.JUDGE_SUBMISSION,
		Submission: subID,
	}, priority); err != nil {
		return err
	}
	return nil
}

func (s *server) SubmissionList(ctx context.Context, in *pb.SubmissionListRequest) (*pb.SubmissionListResponse, error) {
	if 1000 < in.Limit {
		return nil, errors.New("limit must not greater than 1000")
	}

	var order []database.SubmissionOrder
	if in.Order == "" || in.Order == "-id" {
		order = []database.SubmissionOrder{database.ID_DESC}
	} else if in.Order == "+time" {
		order = []database.SubmissionOrder{database.MAX_TIME_ASC, database.ID_DESC}
	} else {
		return nil, errors.New("unknown sort order")
	}

	list, count, err := database.FetchSubmissionList(s.db, in.Problem, in.Status, in.Lang, in.User, in.DedupUser, order, int(in.Skip), int(in.Limit))
	if err != nil {
		return nil, err
	}

	res := pb.SubmissionListResponse{
		Count: int32(count),
	}
	for _, sub := range list {
		protoSub := toProtoSubmissionOverview(sub)
		res.Submissions = append(res.Submissions, &protoSub)
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

	if err := s.pushTask(ctx, in.Id, REJUDGE_PRIORITY); err != nil {
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
	if err := json.Unmarshal([]byte(*data), &categories); err != nil {
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
