package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/yosupo06/library-checker-judge/database"
)

func TestToProtoTimestamp(t *testing.T) {
	if toProtoTimestamp(time.Time{}) != nil {
		t.Fatal("toProtoTimestamp(time.Time{}) should returns default value")
	}
}

func TestToProtoSubmissionOverview(t *testing.T) {
	param := database.SubmissionOverView{
		ID:               1,
		SubmissionTime:   time.Now(),
		ProblemName:      "aplusb",
		Problem:          database.Problem{Name: "aplusb", Title: "A + B"},
		Lang:             "C+",
		Status:           "AC",
		TestCasesVersion: "1",
		MaxTime:          1000,
		MaxMemory:        12345,
		UserName:         sql.NullString{Valid: true, String: "user1"},
		User: &database.User{
			Name: "user1",
		},
	}
	submission := toProtoSubmissionOverview(param)
	if submission.Id != param.ID {
		t.Fatal("submission.Id != param.ID")
	}
	if submission.SubmissionTime.String() != toProtoTimestamp(param.SubmissionTime).String() {
		t.Fatal("submission.SubmissionTime != param.SubmissionTime")
	}
	if submission.ProblemName != param.Problem.Name {
		t.Fatal("submission.ProblemName != param.ProblemName")
	}
	if submission.ProblemTitle != param.Problem.Title {
		t.Fatal("submission.ProblemTitle != param.Problem.Title")
	}
	if submission.UserName != param.UserName.String {
		t.Fatal("submission.UserName != param.UserName")
	}
	if submission.Lang != param.Lang {
		t.Fatal("submission.Lang != param.Lang")
	}
	if submission.IsLatest != (param.TestCasesVersion == param.Problem.TestCasesVersion) {
		t.Fatal("submission.IsLatest != param.IsLatest")
	}
	if submission.Status != param.Status {
		t.Fatal("submission.Status != param.Status")
	}
}

func TestToProtoSubmissionOverviewEmptyUser(t *testing.T) {
	param := database.SubmissionOverView{
		ID:          1,
		ProblemName: "aplusb",
		Problem:     database.Problem{Name: "aplusb", Title: "A + B"},
		UserName:    sql.NullString{Valid: false},
		User:        nil,
	}
	submission := toProtoSubmissionOverview(param)
	if submission.UserName != "" {
		t.Fatal("submission.UserName != param.UserName")
	}
}
