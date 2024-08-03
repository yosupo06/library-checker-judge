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

const (
	TEST_CASE_TXT_LENGTH_LIMIT = 256 * 1024
	TEST_CASE_CPP_LENGTH_LIMIT = 256 * 1024
)

func (s *server) Hack(ctx context.Context, in *pb.HackRequest) (*pb.HackResponse, error) {
	txt := in.GetTxt()
	cpp := in.GetCpp()
	if len(txt) > TEST_CASE_TXT_LENGTH_LIMIT {
		return nil, errors.New("test case is too long")
	}
	if len(cpp) > TEST_CASE_CPP_LENGTH_LIMIT {
		return nil, errors.New("test case generator is too long")
	}
	currentUserName := s.currentUserName(ctx)
	currentUser, _ := database.FetchUserFromName(s.db, currentUserName)

	submission := in.GetSubmission()

	if _, err := database.FetchSubmission(s.db, submission); err != nil {
		return nil, err
	}

	name := ""
	if currentUser != nil {
		name = currentUser.Name
	}

	h := database.Hack{
		HackTime:     time.Now(),
		SubmissionID: submission,
		TestCaseTxt:  txt,
		TestCaseCpp:  cpp,
		UserName:     sql.NullString{String: name, Valid: name != ""},
		Status:       "WJ",
	}

	id, err := database.SaveHack(s.db, h)
	if err != nil {
		return nil, err
	}

	slog.Info("Create new hack", "id", id)

	if err := database.PushHackTask(s.db, id, HACK_PRIORITY); err != nil {
		return nil, err
	}

	return &pb.HackResponse{Id: id}, nil
}

func (s *server) HackInfo(ctx context.Context, in *pb.HackInfoRequest) (*pb.HackInfoResponse, error) {
	h, err := database.FetchHack(s.db, in.Id)
	if err != nil {
		return nil, err
	}
	overView := pb.HackOverview{
		Id:           h.ID,
		SubmissionId: h.SubmissionID,
		Status:       h.Status,
		HackTime:     toProtoTimestamp(h.HackTime),
	}
	if h.Time.Valid {
		time := float64(h.Time.Int32) / 1000.0
		overView.Time = &time
	}
	if h.Memory.Valid {
		memory := h.Memory.Int64
		overView.Memory = &memory
	}
	response := pb.HackInfoResponse{
		Overview:   &overView,
		CheckerOut: h.CheckerOut,
	}
	response.TestCase = &pb.HackInfoResponse_Cpp{
		Cpp: h.TestCaseCpp,
	}
	return &response, nil
}
