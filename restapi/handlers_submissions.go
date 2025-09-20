package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/yosupo06/library-checker-judge/database"
	"github.com/yosupo06/library-checker-judge/langs"
	restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
)

// PostSubmit handles POST /submit
func (s *server) PostSubmit(ctx context.Context, request restapi.PostSubmitRequestObject) (restapi.PostSubmitResponseObject, error) {
	if request.Body == nil {
		return nil, newHTTPError(http.StatusBadRequest, "invalid json")
	}
	body := request.Body
	if body.Problem == "" || body.Source == "" || body.Lang == "" {
		return nil, newHTTPError(http.StatusBadRequest, "missing required fields")
	}
	if len(body.Source) == 0 || len(body.Source) > 1024*1024 {
		return nil, newHTTPError(http.StatusBadRequest, "invalid source length")
	}
	if _, err := database.FetchProblem(s.db, body.Problem); err != nil {
		return nil, newHTTPError(http.StatusBadRequest, "unknown problem")
	}
	if _, ok := langs.GetLang(body.Lang); !ok {
		return nil, newHTTPError(http.StatusBadRequest, "unknown language")
	}

	var userName sql.NullString
	if req, ok := httpRequestFromContext(ctx); ok && s.authClient != nil {
		if token := parseBearerToken(req); token != "" {
			uid := s.authClient.parseUID(req.Context(), token)
			if uid != "" {
				if user, err := database.FetchUserFromUID(s.db, uid); err != nil {
					return nil, newHTTPError(http.StatusInternalServerError, "failed to fetch user")
				} else if user != nil {
					userName = sql.NullString{String: user.Name, Valid: true}
				}
			}
		}
	}

	sub := database.Submission{
		SubmissionTime: time.Now(),
		ProblemName:    body.Problem,
		Lang:           body.Lang,
		Status:         "WJ",
		Source:         body.Source,
		MaxTime:        -1,
		MaxMemory:      -1,
		UserName:       userName,
	}
	id, err := database.SaveSubmission(s.db, sub)
	if err != nil {
		return nil, newHTTPError(http.StatusInternalServerError, "submit failed")
	}
	tleKnockout := false
	if body.TleKnockout != nil {
		tleKnockout = *body.TleKnockout
	}
	if err := database.PushSubmissionTask(s.db, database.SubmissionData{ID: id, TleKnockout: tleKnockout}, 45); err != nil {
		return nil, newHTTPError(http.StatusInternalServerError, "enqueue failed")
	}

	return restapi.PostSubmit200JSONResponse(restapi.SubmitResponse{Id: id}), nil
}

// GetSubmissionInfo handles GET /submissions/{id}
func (s *server) GetSubmissionInfo(_ context.Context, request restapi.GetSubmissionInfoRequestObject) (restapi.GetSubmissionInfoResponseObject, error) {
	sub, err := database.FetchSubmission(s.db, request.Id)
	if err != nil {
		if err == database.ErrNotExist {
			return nil, newHTTPError(http.StatusNotFound, "not found")
		}
		return nil, newHTTPError(http.StatusInternalServerError, "failed to fetch submission")
	}
	cases, err := database.FetchTestcaseResults(s.db, request.Id)
	if err != nil {
		return nil, newHTTPError(http.StatusInternalServerError, "failed to fetch cases")
	}

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
		CanRejudge:   false,
	}
	if len(cr) > 0 {
		resp.CaseResults = &cr
	}
	return restapi.GetSubmissionInfo200JSONResponse(resp), nil
}

// GetSubmissionList handles GET /submissions
func (s *server) GetSubmissionList(_ context.Context, request restapi.GetSubmissionListRequestObject) (restapi.GetSubmissionListResponseObject, error) {
	skip := 0
	limit := 100
	if request.Params.Skip != nil {
		skip = int(*request.Params.Skip)
	}
	if request.Params.Limit != nil {
		limit = int(*request.Params.Limit)
	}
	if limit > 1000 {
		return nil, newHTTPError(http.StatusBadRequest, "limit must not be greater than 1000")
	}
	order := ""
	if request.Params.Order != nil {
		order = string(*request.Params.Order)
	}
	var dbOrder []database.SubmissionOrder
	switch order {
	case "", "-id":
		dbOrder = []database.SubmissionOrder{database.ID_DESC}
	case "+time":
		dbOrder = []database.SubmissionOrder{database.MAX_TIME_ASC, database.ID_DESC}
	default:
		return nil, newHTTPError(http.StatusBadRequest, "unknown sort order")
	}
	problem := deref(request.Params.Problem)
	status := deref(request.Params.Status)
	lang := deref(request.Params.Lang)
	user := deref(request.Params.User)
	dedup := false
	if request.Params.DedupUser != nil {
		dedup = *request.Params.DedupUser
	}
	list, count, err := database.FetchSubmissionList(s.db, problem, status, lang, user, dedup, dbOrder, skip, limit)
	if err != nil {
		return nil, newHTTPError(http.StatusInternalServerError, "failed to fetch submissions")
	}
	overviews := make([]restapi.SubmissionOverview, 0, len(list))
	for _, sub := range list {
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
		overviews = append(overviews, overview)
	}
	resp := restapi.SubmissionListResponse{Submissions: overviews, Count: int32(count)}
	return restapi.GetSubmissionList200JSONResponse(resp), nil
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}
