package main

import (
	"time"

	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"github.com/yosupo06/library-checker-judge/database"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoSubmissionOverview(s database.SubmissionOverView) pb.SubmissionOverview {
	return pb.SubmissionOverview{
		Id:             int32(s.ID),
		ProblemName:    s.Problem.Name,
		ProblemTitle:   s.Problem.Title,
		UserName:       s.UserName.String,
		Lang:           s.Lang,
		IsLatest:       s.TestCasesVersion == s.Problem.TestCasesVersion,
		Status:         s.Status,
		Time:           float64(s.MaxTime) / 1000.0,
		Memory:         int64(s.MaxMemory),
		SubmissionTime: toProtoTimestamp(s.SubmissionTime),
	}
}

func toProtoTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	} else {
		return timestamppb.New(t)
	}
}

func toProtoProblemInfo(p database.Problem) *pb.ProblemInfoResponse {
	return &pb.ProblemInfoResponse{
		Title:            p.Title,
		TimeLimit:        float64(p.Timelimit) / 1000.0,
		SourceUrl:        p.SourceUrl,
		Version:          p.Version,
		TestcasesVersion: p.TestCasesVersion,
	}
}

func toProtoUser(user *database.User) *pb.User {
	if user == nil {
		return nil
	}
	return &pb.User{
		Name:        user.Name,
		LibraryUrl:  user.LibraryURL,
		IsDeveloper: user.IsDeveloper,
	}
}
