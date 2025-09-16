package main

import (
	"database/sql"
	"encoding/json"
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

func (s *server) PostHack(w http.ResponseWriter, r *http.Request) {
	var req restapi.CreateHackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Submission < 0 {
		http.Error(w, "invalid submission id", http.StatusBadRequest)
		return
	}

	txt := []byte(nil)
	if req.TestCaseTxt != nil {
		txt = append(txt, (*req.TestCaseTxt)...)
	}
	cpp := []byte(nil)
	if req.TestCaseCpp != nil {
		cpp = append(cpp, (*req.TestCaseCpp)...)
	}

	if len(txt) > testCaseTextLengthLimit {
		http.Error(w, "test case is too long", http.StatusBadRequest)
		return
	}
	if len(cpp) > testCaseSourceLengthLimit {
		http.Error(w, "test case generator is too long", http.StatusBadRequest)
		return
	}

	uid, err := s.uidFromRequest(r)
	if err != nil || uid == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var userName string
	if user, err := database.FetchUserFromUID(s.db, uid); err == nil && user != nil {
		userName = user.Name
	}

	if _, err := database.FetchSubmission(s.db, req.Submission); err != nil {
		if errors.Is(err, database.ErrNotExist) {
			http.Error(w, "submission not found", http.StatusNotFound)
		} else {
			http.Error(w, "failed to fetch submission", http.StatusInternalServerError)
		}
		return
	}

	hack := database.Hack{
		HackTime:     time.Now(),
		SubmissionID: req.Submission,
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
		http.Error(w, "hack creation failed", status)
		return
	}

	if err := database.PushHackTask(s.db, database.HackData{ID: id}, hackTaskPriority); err != nil {
		http.Error(w, "enqueue failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(restapi.HackResponse{Id: id})
}

func (s *server) GetHackInfo(w http.ResponseWriter, r *http.Request, id int32) {
	h, err := database.FetchHack(s.db, id)
	if err != nil {
		if errors.Is(err, database.ErrNotExist) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "failed to fetch hack", http.StatusInternalServerError)
		}
		return
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

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *server) GetHackList(w http.ResponseWriter, r *http.Request, params restapi.GetHackListParams) {
	skip := 0
	if params.Skip != nil {
		skip = int(*params.Skip)
	}
	limit := 100
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	if limit > 1000 {
		http.Error(w, "limit must not be greater than 1000", http.StatusBadRequest)
		return
	}

	user := ""
	if params.User != nil {
		user = *params.User
	}
	status := ""
	if params.Status != nil {
		status = *params.Status
	}
	order := ""
	if params.Order != nil {
		order = *params.Order
	}

	hacks, err := database.FetchHackList(s.db, skip, limit, user, status, order)
	if err != nil {
		http.Error(w, "failed to fetch hacks", http.StatusInternalServerError)
		return
	}
	count, err := database.CountHacks(s.db, user, status)
	if err != nil {
		http.Error(w, "failed to count hacks", http.StatusInternalServerError)
		return
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

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(restapi.HackListResponse{Hacks: overviews, Count: int32(count)})
}
