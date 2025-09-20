package main

import (
	"context"
	"net/http"

	"github.com/yosupo06/library-checker-judge/database"
	restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
)

// GetProblems handles GET /problems
func (s *server) GetProblems(ctx context.Context, _ restapi.GetProblemsRequestObject) (restapi.GetProblemsResponseObject, error) {
	rows, err := database.FetchProblemList(s.db)
	if err != nil {
		return nil, newHTTPError(http.StatusInternalServerError, "failed to fetch problems")
	}
	problems := make([]restapi.Problem, 0, len(rows))
	for _, p := range rows {
		problems = append(problems, restapi.Problem{Name: p.Name, Title: p.Title})
	}
	resp := restapi.ProblemListResponse{Problems: problems}
	return restapi.GetProblems200JSONResponse(resp), nil
}

// GetProblemInfo handles GET /problems/{name}
func (s *server) GetProblemInfo(ctx context.Context, request restapi.GetProblemInfoRequestObject) (restapi.GetProblemInfoResponseObject, error) {
	if request.Name == "" {
		return nil, newHTTPError(http.StatusBadRequest, "missing problem name")
	}
	p, err := database.FetchProblem(s.db, request.Name)
	if err != nil {
		if err == database.ErrNotExist {
			return nil, newHTTPError(http.StatusNotFound, "problem not found")
		}
		return nil, newHTTPError(http.StatusInternalServerError, "failed to fetch problem")
	}
	resp := restapi.ProblemInfoResponse{
		Title:            p.Title,
		SourceUrl:        p.SourceUrl,
		TimeLimit:        float32(p.Timelimit) / 1000.0,
		Version:          p.Version,
		TestcasesVersion: p.TestCasesVersion,
		OverallVersion:   p.OverallVersion,
	}
	return restapi.GetProblemInfo200JSONResponse(resp), nil
}
