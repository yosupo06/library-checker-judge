package main

import (
	"context"
	"errors"
	"log"
	"sort"

	"github.com/go-playground/validator/v10"
	_ "github.com/lib/pq"
	"gorm.io/gorm"

	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"github.com/yosupo06/library-checker-judge/database"
	"github.com/yosupo06/library-checker-judge/langs"
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

func (s *server) pushTask(_ context.Context, subID, priority int32) error {
	if err := database.PushTask(s.db, database.TaskData{
		TaskType:   database.JUDGE_SUBMISSION,
		Submission: subID,
	}, priority); err != nil {
		return err
	}
	return nil
}

func (s *server) LangList(ctx context.Context, in *pb.LangListRequest) (*pb.LangListResponse, error) {
	var pbLangs []*pb.Lang
	for _, lang := range langs.LANGS {
		pbLangs = append(pbLangs, &pb.Lang{
			Id:      lang.ID,
			Name:    lang.Name,
			Version: lang.Version,
		})
	}
	return &pb.LangListResponse{Langs: pbLangs}, nil
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

func (s *server) ProblemCategories(ctx context.Context, in *pb.ProblemCategoriesRequest) (*pb.ProblemCategoriesResponse, error) {
	categories, err := database.FetchProblemCategories(s.db)
	if err != nil {
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
