package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/yosupo06/library-checker-judge/database"
	"github.com/yosupo06/library-checker-judge/langs"
	restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
)

// PostSubmit handles POST /submit
func (s *server) PostSubmit(w http.ResponseWriter, r *http.Request) {
	var req restapi.SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Problem == "" || req.Source == "" || req.Lang == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}
	// Validate problem exists
	if _, err := database.FetchProblem(s.db, req.Problem); err != nil {
		http.Error(w, "unknown problem", http.StatusBadRequest)
		return
	}
	// Validate language exists
	if _, ok := langs.GetLang(req.Lang); !ok {
		http.Error(w, "unknown language", http.StatusBadRequest)
		return
	}
	// Validate source size (<= 1MB)
	if len(req.Source) == 0 || len(req.Source) > 1024*1024 {
		http.Error(w, "invalid source length", http.StatusBadRequest)
		return
	}
	var userName sql.NullString
	if token := parseBearerToken(r); token != "" && s.authClient != nil {
		uid := s.authClient.parseUID(r.Context(), token)
		if uid != "" {
			if user, err := database.FetchUserFromUID(s.db, uid); err != nil {
				http.Error(w, "failed to fetch user", http.StatusInternalServerError)
				return
			} else if user != nil {
				userName = sql.NullString{String: user.Name, Valid: true}
			}
		}
	}
	// Create submission and associate user when available
	sub := database.Submission{
		SubmissionTime: time.Now(),
		ProblemName:    req.Problem,
		Lang:           req.Lang,
		Status:         "WJ",
		Source:         req.Source,
		MaxTime:        -1,
		MaxMemory:      -1,
		UserName:       userName,
	}
	id, err := database.SaveSubmission(s.db, sub)
	if err != nil {
		http.Error(w, "submit failed", http.StatusInternalServerError)
		return
	}
	tleKnockout := false
	if req.TleKnockout != nil {
		tleKnockout = *req.TleKnockout
	}
	if err := database.PushSubmissionTask(s.db, database.SubmissionData{ID: id, TleKnockout: tleKnockout}, 45); err != nil {
		http.Error(w, "enqueue failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(restapi.SubmitResponse{Id: id})
}

// GetSubmissionInfo handles GET /submissions/{id}
func (s *server) GetSubmissionInfo(w http.ResponseWriter, r *http.Request, id int32) {
	sub, err := database.FetchSubmission(s.db, id)
	if err != nil {
		if err == database.ErrNotExist {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch submission", http.StatusInternalServerError)
		return
	}
	cases, err := database.FetchTestcaseResults(s.db, id)
	if err != nil {
		http.Error(w, "failed to fetch cases", http.StatusInternalServerError)
		return
	}
	// Build overview
	var userNamePtr *string
	if sub.UserName.Valid {
		v := sub.UserName.String
		userNamePtr = &v
	}
	overview := restapi.SubmissionOverview{
		Id:           sub.ID,
		ProblemName:  sub.Problem.Name,
		ProblemTitle: sub.Problem.Title,
		UserName:     userNamePtr,
		Lang:         sub.Lang,
		IsLatest:     sub.TestCasesVersion == sub.Problem.TestCasesVersion,
		Status:       sub.Status,
		Time:         float32(sub.MaxTime) / 1000.0,
		Memory:       sub.MaxMemory,
	}
	if !sub.SubmissionTime.IsZero() {
		t := sub.SubmissionTime
		overview.SubmissionTime = &t
	}
	// Build case results
	cr := make([]restapi.SubmissionCaseResult, 0, len(cases))
	for _, c := range cases {
		var stderr *[]byte
		if len(c.Stderr) > 0 {
			b := c.Stderr
			stderr = &b
		}
		var checker *[]byte
		if len(c.CheckerOut) > 0 {
			b := c.CheckerOut
			checker = &b
		}
		cr = append(cr, restapi.SubmissionCaseResult{
			Case:       c.Testcase,
			Status:     c.Status,
			Time:       float32(c.Time) / 1000.0,
			Memory:     c.Memory,
			Stderr:     stderr,
			CheckerOut: checker,
		})
	}
	var compileErr *[]byte
	if len(sub.CompileError) > 0 {
		b := sub.CompileError
		compileErr = &b
	}
	resp := restapi.SubmissionInfoResponse{
		Overview:     overview,
		Source:       sub.Source,
		CompileError: compileErr,
		CanRejudge:   false, // no auth in REST yet
	}
	if len(cr) > 0 {
		resp.CaseResults = &cr
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// GetSubmissionList handles GET /submissions
func (s *server) GetSubmissionList(w http.ResponseWriter, r *http.Request, params restapi.GetSubmissionListParams) {
	// Defaults
	skip := 0
	limit := 100
	if params.Skip != nil {
		skip = int(*params.Skip)
	}
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	if limit > 1000 {
		http.Error(w, "limit must not be greater than 1000", http.StatusBadRequest)
		return
	}
	order := ""
	if params.Order != nil {
		order = string(*params.Order)
	}
	var dbOrder []database.SubmissionOrder
	switch order {
	case "", "-id":
		dbOrder = []database.SubmissionOrder{database.ID_DESC}
	case "+time":
		dbOrder = []database.SubmissionOrder{database.MAX_TIME_ASC, database.ID_DESC}
	default:
		http.Error(w, "unknown sort order", http.StatusBadRequest)
		return
	}
	problem := deref(params.Problem)
	status := deref(params.Status)
	lang := deref(params.Lang)
	user := deref(params.User)
	dedup := false
	if params.DedupUser != nil {
		dedup = *params.DedupUser
	}
	list, count, err := database.FetchSubmissionList(s.db, problem, status, lang, user, dedup, dbOrder, skip, limit)
	if err != nil {
		http.Error(w, "failed to fetch submissions", http.StatusInternalServerError)
		return
	}
	overviews := make([]restapi.SubmissionOverview, 0, len(list))
	for _, sub := range list {
		var userNamePtr *string
		if sub.UserName.Valid {
			v := sub.UserName.String
			userNamePtr = &v
		}
		ov := restapi.SubmissionOverview{
			Id:           sub.ID,
			ProblemName:  sub.Problem.Name,
			ProblemTitle: sub.Problem.Title,
			UserName:     userNamePtr,
			Lang:         sub.Lang,
			IsLatest:     sub.TestCasesVersion == sub.Problem.TestCasesVersion,
			Status:       sub.Status,
			Time:         float32(sub.MaxTime) / 1000.0,
			Memory:       sub.MaxMemory,
		}
		if !sub.SubmissionTime.IsZero() {
			t := sub.SubmissionTime
			ov.SubmissionTime = &t
		}
		overviews = append(overviews, ov)
	}
	resp := restapi.SubmissionListResponse{Submissions: overviews, Count: int32(count)}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}
