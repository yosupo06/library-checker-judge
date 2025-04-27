package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"sort"
	"time"

	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"github.com/yosupo06/library-checker-judge/database"
	"github.com/yosupo06/library-checker-judge/langs"
)

func (s *server) Submit(ctx context.Context, in *pb.SubmitRequest) (*pb.SubmitResponse, error) {
	if in.Source == "" {
		return nil, errors.New("empty Source")
	}
	if len(in.Source) > 1024*1024 {
		return nil, errors.New("too large Source")
	}
	if _, ok := langs.GetLang(in.Lang); !ok {
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

func (s *server) SubmissionInfo(ctx context.Context, in *pb.SubmissionInfoRequest) (*pb.SubmissionInfoResponse, error) {
	currentUserName := s.currentUserName(ctx)
	currentUser, _ := database.FetchUserFromName(s.db, currentUserName)

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

	overview := toProtoSubmissionOverview(database.ToSubmissionOverView(sub))

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
