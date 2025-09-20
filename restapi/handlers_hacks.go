package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/yosupo06/library-checker-judge/database"
	restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
)

const (
	hackTaskPriority          = 10
	testCaseTextLengthLimit   = 1024 * 1024
	testCaseSourceLengthLimit = 1024 * 1024
)

func (s *server) PostHack(ctx context.Context, request restapi.PostHackRequestObject) (restapi.PostHackResponseObject, error) {
	if request.Body == nil {
		return nil, newHTTPError(http.StatusBadRequest, "invalid request")
	}
	body := request.Body
	if body.Submission < 0 {
		return nil, newHTTPError(http.StatusBadRequest, "invalid submission id")
	}

	txt := []byte(nil)
	if body.TestCaseTxt != nil {
		txt = append(txt, (*body.TestCaseTxt)...)
	}
	cpp := []byte(nil)
	if body.TestCaseCpp != nil {
		cpp = append(cpp, (*body.TestCaseCpp)...)
	}

	if len(txt) > testCaseTextLengthLimit {
		return nil, newHTTPError(http.StatusBadRequest, "test case is too long")
	}
	if len(cpp) > testCaseSourceLengthLimit {
		return nil, newHTTPError(http.StatusBadRequest, "test case generator is too long")
	}

	uid, err := s.uidFromContext(ctx)
	if err != nil || uid == "" {
		return nil, newHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var userName string
	if user, err := database.FetchUserFromUID(s.db, uid); err == nil && user != nil {
		userName = user.Name
	}

	if _, err := database.FetchSubmission(s.db, body.Submission); err != nil {
		if errors.Is(err, database.ErrNotExist) {
			return nil, newHTTPError(http.StatusNotFound, "submission not found")
		}
		return nil, newHTTPError(http.StatusInternalServerError, "failed to fetch submission")
	}

	hack := database.Hack{
		HackTime:     time.Now(),
		SubmissionID: body.Submission,
		TestCaseTxt:  txt,
		TestCaseCpp:  cpp,
		Status:       "WJ",
	}
	if userName != "" {
		hack.UserName = sql.NullString{String: userName, Valid: true}
	}

	id, err := database.SaveHack(s.db, hack)
	if err != nil {
		msg := err.Error()
		status := http.StatusInternalServerError
		if strings.Contains(msg, "must contain") || strings.Contains(msg, "must not") {
			status = http.StatusBadRequest
		}
		return nil, newHTTPError(status, "hack creation failed")
	}

	if err := database.PushHackTask(s.db, database.HackData{ID: id}, hackTaskPriority); err != nil {
		return nil, newHTTPError(http.StatusInternalServerError, "enqueue failed")
	}

	return restapi.PostHack200JSONResponse(restapi.HackResponse{Id: id}), nil
}

func (s *server) GetHackInfo(_ context.Context, request restapi.GetHackInfoRequestObject) (restapi.GetHackInfoResponseObject, error) {
	h, err := database.FetchHack(s.db, request.Id)
	if err != nil {
		if errors.Is(err, database.ErrNotExist) {
			return nil, newHTTPError(http.StatusNotFound, "not found")
		}
		return nil, newHTTPError(http.StatusInternalServerError, "failed to fetch hack")
	}

	overview := restapi.HackOverview{
		Id:           h.ID,
		SubmissionId: h.SubmissionID,
		Status:       h.Status,
		HackTime:     h.HackTime,
	}
	if h.User != nil && h.User.Name != "" {
		name := h.User.Name
		overview.UserName = &name
	} else if h.UserName.Valid {
		name := h.UserName.String
		overview.UserName = &name
	}
	if h.Time.Valid {
		v := float32(h.Time.Int32) / 1000.0
		overview.Time = &v
	}
	if h.Memory.Valid {
		v := h.Memory.Int64
		overview.Memory = &v
	}

	resp := restapi.HackInfoResponse{Overview: overview}
	if len(h.TestCaseTxt) > 0 {
		b := append([]byte(nil), h.TestCaseTxt...)
		resp.TestCaseTxt = &b
	}
	if len(h.TestCaseCpp) > 0 {
		b := append([]byte(nil), h.TestCaseCpp...)
		resp.TestCaseCpp = &b
	}
	if len(h.Stderr) > 0 {
		b := append([]byte(nil), h.Stderr...)
		resp.Stderr = &b
	}
	if len(h.JudgeOutput) > 0 {
		b := append([]byte(nil), h.JudgeOutput...)
		resp.JudgeOutput = &b
	}

	return restapi.GetHackInfo200JSONResponse(resp), nil
}

func (s *server) GetHackList(_ context.Context, request restapi.GetHackListRequestObject) (restapi.GetHackListResponseObject, error) {
	skip := 0
	if request.Params.Skip != nil {
		skip = int(*request.Params.Skip)
	}
	limit := 100
	if request.Params.Limit != nil {
		limit = int(*request.Params.Limit)
	}
	if limit > 1000 {
		return nil, newHTTPError(http.StatusBadRequest, "limit must not be greater than 1000")
	}

	user := ""
	if request.Params.User != nil {
		user = *request.Params.User
	}
	status := ""
	if request.Params.Status != nil {
		status = *request.Params.Status
	}
	order := ""
	if request.Params.Order != nil {
		order = *request.Params.Order
	}

	hacks, err := database.FetchHackList(s.db, skip, limit, user, status, order)
	if err != nil {
		return nil, newHTTPError(http.StatusInternalServerError, "failed to fetch hacks")
	}
	count, err := database.CountHacks(s.db, user, status)
	if err != nil {
		return nil, newHTTPError(http.StatusInternalServerError, "failed to count hacks")
	}

	overviews := make([]restapi.HackOverview, 0, len(hacks))
	for _, h := range hacks {
		overview := restapi.HackOverview{
			Id:           h.ID,
			SubmissionId: h.SubmissionID,
			Status:       h.Status,
			HackTime:     h.HackTime,
		}
		if h.User != nil && h.User.Name != "" {
			name := h.User.Name
			overview.UserName = &name
		} else if h.UserName.Valid {
			name := h.UserName.String
			overview.UserName = &name
		}
		if h.Time.Valid {
			v := float32(h.Time.Int32) / 1000.0
			overview.Time = &v
		}
		if h.Memory.Valid {
			v := h.Memory.Int64
			overview.Memory = &v
		}
		overviews = append(overviews, overview)
	}

	resp := restapi.HackListResponse{Hacks: overviews, Count: int32(count)}
	return restapi.GetHackList200JSONResponse(resp), nil
}
