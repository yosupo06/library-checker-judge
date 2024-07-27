package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"github.com/yosupo06/library-checker-judge/database"
)

func (s *server) Hack(ctx context.Context, in *pb.HackRequest) (*pb.HackResponse, error) {
	tc := in.GetTestCase()
	if len(tc) == 0 {
		return nil, errors.New("test case is empty")
	}
	if len(tc) > 1024*1024 {
		return nil, errors.New("test case is too long")
	}
	currentUserName := s.currentUserName(ctx)
	currentUser, _ := database.FetchUserFromName(s.db, currentUserName)

	name := ""
	if currentUser != nil {
		name = currentUser.Name
	}

	h := database.Hack{
		HackTime:     time.Now(),
		SubmissionID: in.GetSubmission(),
		TestCase:     in.GetTestCase(),
		UserName:     sql.NullString{String: name, Valid: name != ""},
	}

	id, err := database.SaveHack(s.db, h)
	if err != nil {
		return nil, err
	}

	slog.Info("Create new hack", "id", id)

	if err := database.PushHackTask(s.db, id, HACK_PRIORITY); err != nil {
		return nil, err
	}

	return &pb.HackResponse{Id: &id}, nil
}

func (s *server) HackInfo(ctx context.Context, in *pb.HackInfoRequest) (*pb.HackInfoResponse, error) {
	h, err := database.FetchHack(s.db, *in.Id)
	if err != nil {
		return nil, err
	}
	time := float64(h.Time) / 1000.0
	memory := int64(h.Memory)
	return &pb.HackInfoResponse{
		Overview: &pb.HackOverview{
			Id:           &h.ID,
			SubmissionId: &h.SubmissionID,
			Status:       &h.Status,
			Time:         &time,
			Memory:       &memory,
			HackTime:     toProtoTimestamp(h.HackTime),
		},
		TestCase:   h.TestCase,
		CheckerOut: h.CheckerOut,
	}, nil
}
