package main

import (
    "encoding/json"
    "net/http"

    "github.com/yosupo06/library-checker-judge/database"
    restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
)

// GetProblems handles GET /problems
func (s *server) GetProblems(w http.ResponseWriter, r *http.Request) {
    rows, err := database.FetchProblemList(s.db)
    if err != nil {
        http.Error(w, "failed to fetch problems", http.StatusInternalServerError)
        return
    }
    problems := make([]restapi.Problem, 0, len(rows))
    for _, p := range rows {
        problems = append(problems, restapi.Problem{Name: p.Name, Title: p.Title})
    }
    resp := restapi.ProblemListResponse{Problems: problems}
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(resp)
}

// GetProblemInfo handles GET /problems/{name}
func (s *server) GetProblemInfo(w http.ResponseWriter, r *http.Request, name string) {
    if name == "" {
        http.Error(w, "missing problem name", http.StatusBadRequest)
        return
    }
    p, err := database.FetchProblem(s.db, name)
    if err != nil {
        if err == database.ErrNotExist {
            http.Error(w, "problem not found", http.StatusNotFound)
            return
        }
        http.Error(w, "failed to fetch problem", http.StatusInternalServerError)
        return
    }
    resp := restapi.ProblemInfoResponse{
        Title:            p.Title,
        SourceUrl:        p.SourceUrl,
        TimeLimit:        float32(p.Timelimit) / 1000.0,
        Version:          p.Version,
        TestcasesVersion: p.TestCasesVersion,
        OverallVersion:   p.OverallVersion,
    }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(resp)
}
