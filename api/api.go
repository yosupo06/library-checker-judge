package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"sort"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	pb "github.com/yosupo06/library-checker-judge/api/proto"
)

func (s *server) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if in.Name == "" {
		return nil, errors.New("Empty userName")
	}
	if in.Password == "" {
		return nil, errors.New("Empty password")
	}
	passHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 10)
	if err != nil {
		return nil, errors.New("Bcrypt broken")
	}
	user := User{
		Name:     in.Name,
		Passhash: string(passHash),
	}
	if err := db.Create(&user).Error; err != nil {
		return nil, errors.New("This username are already registered")
	}
	token, err := issueToken(in.Name)
	if err != nil {
		return nil, errors.New("Broken")
	}
	return &pb.RegisterResponse{
		Token: token,
	}, nil
}

func (s *server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	var user User
	if err := db.Where("name = ?", in.Name).Take(&user).Error; err != nil {
		return nil, errors.New("Invalid username")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Passhash), []byte(in.Password)); err != nil {
		return nil, errors.New("Invalid password")
	}

	token, err := issueToken(in.Name)
	if err != nil {
		return nil, err
	}
	return &pb.LoginResponse{
		Token: token,
	}, nil
}

func (s *server) UserInfo(ctx context.Context, in *pb.UserInfoRequest) (*pb.UserInfoResponse, error) {
	if getUserName(ctx) == "" {
		return nil, errors.New("Not login")
	}
	return &pb.UserInfoResponse{
		IsAdmin: isAdmin(ctx),
		User: &pb.User{
			Name:    getUserName(ctx),
			IsAdmin: isAdmin(ctx),
		},
	}, nil
}

func (s *server) UserList(ctx context.Context, in *pb.UserListRequest) (*pb.UserListResponse, error) {
	if getUserName(ctx) == "" {
		return nil, errors.New("Not login")
	}
	if !isAdmin(ctx) {
		return nil, errors.New("Must be admin")
	}
	users := []User{}
	if err := db.Select("name, admin").Find(&users).Error; err != nil {
		return nil, errors.New("Failed to get users")
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
	if getUserName(ctx) == "" {
		return nil, errors.New("Not login")
	}
	if !isAdmin(ctx) {
		return nil, errors.New("Must be admin")
	}
	if getUserName(ctx) == in.User.Name && !in.User.IsAdmin {
		return nil, errors.New("Cannot remove myself from admin")
	}

	tx := db.Begin()
	var user User
	if err := tx.Where("name = ?", in.User.Name).First(&user).Error; err != nil {
		// invalid user name
		tx.Rollback()
		log.Println(err)
		return nil, errors.New("No User")
	}
	if err := tx.Model(&User{}).Where("name = ?", in.User.Name).Updates(
		map[string]interface{}{
			"admin": in.User.IsAdmin,
		}).Error; err != nil {
		tx.Rollback()
		log.Println(err)
		return nil, errors.New("Failed to update user")
	}
	tx.Commit()
	return &pb.ChangeUserInfoResponse{}, nil
}

func (s *server) ProblemInfo(ctx context.Context, in *pb.ProblemInfoRequest) (*pb.ProblemInfoResponse, error) {
	name := in.Name
	if name == "" {
		return nil, errors.New("Empty problem name")
	}
	var problem Problem
	if err := db.Select("name, title, statement, timelimit").Where("name = ?", name).First(&problem).Error; err != nil {
		return nil, errors.New("Failed to get problem")
	}
	return &pb.ProblemInfoResponse{
		Title:     problem.Title,
		Statement: problem.Statement,
		TimeLimit: problem.Timelimit / 1000.0,
	}, nil
}

func (s *server) ProblemList(ctx context.Context, in *pb.ProblemListRequest) (*pb.ProblemListResponse, error) {
	problems := []Problem{}
	if err := db.Select("name, title").Find(&problems).Error; err != nil {
		return nil, errors.New("Fetch problems failed")
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
		return nil, errors.New("Empty Source")
	}
	if len(in.Source) > 1024*1024 {
		return nil, errors.New("Too large Source")
	}
	ok := false
	for _, lang := range langs {
		if lang.Id == in.Lang {
			ok = true
			break
		}
	}
	if !ok {
		return nil, errors.New("Unknown Lang")
	}
	if _, err := s.ProblemInfo(ctx, &pb.ProblemInfoRequest{
		Name: in.Problem,
	}); err != nil {
		return nil, errors.New("Unknown problem")
	}
	user := getUserName(ctx)
	submission := Submission{
		ProblemName: in.Problem,
		Lang:        in.Lang,
		Status:      "WJ",
		Source:      in.Source,
		MaxTime:     -1,
		MaxMemory:   -1,
		UserName:    sql.NullString{String: user, Valid: user != ""},
	}

	if err := db.Create(&submission).Error; err != nil {
		return nil, errors.New("Submit failed")
	}

	task := Task{}
	task.Submission = submission.ID
	if err := db.Create(&task).Error; err != nil {
		return nil, errors.New("Inserting to judge queue is failed")
	}

	log.Println("Submit ", submission.ID)

	return &pb.SubmitResponse{Id: int32(submission.ID)}, nil
}

func (s *server) SubmissionInfo(ctx context.Context, in *pb.SubmissionInfoRequest) (*pb.SubmissionInfoResponse, error) {
	var sub Submission
	sub, err := fetchSubmission(int(in.Id))
	if err != nil {
		return nil, err
	}
	var cases []SubmissionTestcaseResult
	if err := db.Where("submission = ?", in.Id).Find(&cases).Error; err != nil {
		return nil, errors.New("Submission fetch failed")
	}
	overview := &pb.SubmissionOverview{
		Id:           int32(sub.ID),
		ProblemName:  sub.Problem.Name,
		ProblemTitle: sub.Problem.Title,
		UserName:     sub.User.Name,
		Lang:         sub.Lang,
		IsLatest:     sub.Testhash == sub.Problem.Testhash,
		Status:       sub.Status,
		Time:         float64(sub.MaxTime) / 1000.0,
		Memory:       int64(sub.MaxMemory),
	}

	res := &pb.SubmissionInfoResponse{
		Overview:   overview,
		Source:     sub.Source,
		CanRejudge: canRejudge(ctx, overview),
	}

	for _, c := range cases {
		res.CaseResults = append(res.CaseResults, &pb.SubmissionCaseResult{
			Case:   c.Testcase,
			Status: c.Status,
			Time:   float64(c.Time) / 1000.0,
			Memory: int64(c.Memory),
		})
	}
	return res, nil
}

func (s *server) SubmissionList(ctx context.Context, in *pb.SubmissionListRequest) (*pb.SubmissionListResponse, error) {
	if 1000 < in.Limit {
		in.Limit = 1000
	}

	filter := &Submission{
		ProblemName: in.Problem,
		Status:      in.Status,
		UserName:    sql.NullString{String: in.User, Valid: (in.User != "")},
	}

	count := 0
	if err := db.Model(&Submission{}).Where(filter).Count(&count).Error; err != nil {
		return nil, errors.New("Count Query Failed")
	}
	order := ""
	if in.Order == "" || in.Order == "-id" {
		order = "id desc"
	} else if in.Order == "+time" {
		order = "max_time asc"
	} else {
		return nil, errors.New("Unknown Sort Order")
	}

	var submissions = make([]Submission, 0)
	if err := db.Where(filter).Limit(in.Limit).Offset(in.Skip).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("name")
		}).
		Preload("Problem", func(db *gorm.DB) *gorm.DB {
			return db.Select("name, title, testhash")
		}).
		Select("id, user_name, problem_name, lang, status, testhash, max_time, max_memory").
		Order(order).
		Find(&submissions).Error; err != nil {
		return nil, errors.New("Select Query Failed")
	}

	res := pb.SubmissionListResponse{
		Count: int32(count),
	}
	for _, sub := range submissions {
		res.Submissions = append(res.Submissions, &pb.SubmissionOverview{
			Id:           int32(sub.ID),
			ProblemName:  sub.Problem.Name,
			ProblemTitle: sub.Problem.Title,
			UserName:     sub.User.Name,
			Lang:         sub.Lang,
			IsLatest:     sub.Testhash == sub.Problem.Testhash,
			Status:       sub.Status,
			Time:         float64(sub.MaxTime) / 1000.0,
			Memory:       int64(sub.MaxMemory),
		})
	}
	return &res, nil
}

func (s *server) Rejudge(ctx context.Context, in *pb.RejudgeRequest) (*pb.RejudgeResponse, error) {
	sub, err := s.SubmissionInfo(ctx, &pb.SubmissionInfoRequest{Id: in.Id})
	if err != nil {
		return nil, err
	}
	if !sub.CanRejudge {
		return nil, errors.New("No permission")
	}
	task := Task{}
	task.Submission = int(in.Id)
	if err := db.Create(&task).Error; err != nil {
		return nil, errors.New("Cannot insert into queue")
	}
	return &pb.RejudgeResponse{}, nil
}

func (s *server) LangList(ctx context.Context, in *pb.LangListRequest) (*pb.LangListResponse, error) {
	return &pb.LangListResponse{Langs: langs}, nil
}

func (s *server) Ranking(ctx context.Context, in *pb.RankingRequest) (*pb.RankingResponse, error) {
	filter := &Submission{
		Status: "AC",
	}

	var submissions = make([]Submission, 0)
	if err := db.
		Select("id, user_name, problem_name, status").Where(filter).Find(&submissions).Error; err != nil {
		return nil, errors.New("Select Query Failed")
	}

	ac := make(map[string]map[string]bool)
	for _, sub := range submissions {
		if !sub.UserName.Valid {
			continue
		}
		userName := sub.UserName.String
		if _, ok := ac[userName]; !ok {
			ac[userName] = make(map[string]bool)
		}
		ac[userName][sub.ProblemName] = true
	}
	stats := make([]*pb.UserStatistics, 0)
	for name, acs := range ac {
		stats = append(stats, &pb.UserStatistics{
			Name:  name,
			Count: int32(len(acs)),
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

func (s *server) PopJudgeTask(ctx context.Context, in *pb.PopJudgeTaskRequest) (*pb.PopJudgeTaskResponse, error) {
	if !isAdmin(ctx) {
		return nil, errors.New("Permission denied")
	}
	if in.JudgeName == "" {
		return nil, errors.New("JudgeName is empty")
	}
	tx := db.Begin()
	task := Task{}
	err := tx.First(&task).Error
	if gorm.IsRecordNotFoundError(err) {
		tx.Rollback()
		return &pb.PopJudgeTaskResponse{
			SubmissionId: -1,
		}, nil
	}
	if err != nil {
		tx.Rollback()
		log.Print(err.Error())
		return nil, errors.New("Connection to db failed")
	}
	tx.Delete(task)
	if err := tx.Commit().Error; err != nil {
		log.Println(err.Error())
		return nil, errors.New("Connection to db failed")
	}
	log.Println("Pop Submission:", task.Submission)

	id := task.Submission
	ok, err := updateSubmissionRegistration(id, in.JudgeName, true)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &pb.PopJudgeTaskResponse{
			SubmissionId: -1,
		}, nil
	}

	return &pb.PopJudgeTaskResponse{
		SubmissionId: int32(task.Submission),
	}, nil
}

func (s *server) SyncJudgeTaskStatus(ctx context.Context, in *pb.SyncJudgeTaskStatusRequest) (*pb.SyncJudgeTaskStatusResponse, error) {
	if !isAdmin(ctx) {
		return nil, errors.New("Permission denied")
	}
	if in.JudgeName == "" {
		return nil, errors.New("JudgeName is empty")
	}
	id := int(in.SubmissionId)
	ok, err := updateSubmissionRegistration(id, in.JudgeName, true)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	if !ok {
		err := errors.New("Call SyncJudgeTaskStatus to non-registered submission")
		log.Println(err, id)
		return nil, err
	}

	for _, testCase := range in.CaseResults {
		if err := db.Create(&SubmissionTestcaseResult{
			Submission: id,
			Testcase:   testCase.Case,
			Status:     testCase.Status,
			Time:       int(testCase.Time * 1000),
			Memory:     int(testCase.Memory),
		}).Error; err != nil {
			log.Println(err)
			return nil, errors.New("DB update failed")
		}
	}
	if err := db.Model(&Submission{
		ID: id,
	}).Update("status", in.Status).Error; err != nil {
		return nil, errors.New("Update Status Failed")
	}
	if in.IsFinished {
		if err := db.Model(&Submission{
			ID: id,
		}).Updates(map[string]interface{}{
			"judge_name": "",
		}).Error; err != nil {
			log.Print(err)
			return nil, errors.New("Failed to clear judge_name")
		}
	}
	return &pb.SyncJudgeTaskStatusResponse{}, nil
}
