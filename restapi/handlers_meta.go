package main

import (
	"context"
	"net/http"

	"github.com/yosupo06/library-checker-judge/database"
	"github.com/yosupo06/library-checker-judge/langs"
	restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
)

// GetLangList handles GET /langs
func (s *server) GetLangList(_ context.Context, _ restapi.GetLangListRequestObject) (restapi.GetLangListResponseObject, error) {
	var ls []restapi.Lang
	for _, l := range langs.LANGS {
		ls = append(ls, restapi.Lang{Id: l.ID, Name: l.Name, Version: l.Version})
	}
	resp := restapi.LangListResponse{Langs: ls}
	return restapi.GetLangList200JSONResponse(resp), nil
}

// GetProblemCategories handles GET /categories
func (s *server) GetProblemCategories(_ context.Context, _ restapi.GetProblemCategoriesRequestObject) (restapi.GetProblemCategoriesResponseObject, error) {
	cats, err := database.FetchProblemCategories(s.db)
	if err != nil {
		return nil, newHTTPError(http.StatusInternalServerError, "failed to fetch categories")
	}
	result := make([]restapi.ProblemCategory, 0, len(cats))
	for _, c := range cats {
		result = append(result, restapi.ProblemCategory{Title: c.Title, Problems: c.Problems})
	}
	resp := restapi.ProblemCategoriesResponse{Categories: result}
	return restapi.GetProblemCategories200JSONResponse(resp), nil
}

// GetMonitoring handles GET /monitoring
func (s *server) GetMonitoring(_ context.Context, _ restapi.GetMonitoringRequestObject) (restapi.GetMonitoringResponseObject, error) {
	data, err := database.FetchMonitoringData(s.db)
	if err != nil {
		return nil, newHTTPError(http.StatusInternalServerError, "failed to fetch monitoring data")
	}
	resp := restapi.MonitoringResponse{
		TotalUsers:       int32(data.TotalUsers),
		TotalSubmissions: int32(data.TotalSubmissions),
		TaskQueue: restapi.TaskQueueInfo{
			PendingTasks: int32(data.TaskQueue.PendingTasks),
			RunningTasks: int32(data.TaskQueue.RunningTasks),
			TotalTasks:   int32(data.TaskQueue.TotalTasks),
		},
	}
	return restapi.GetMonitoring200JSONResponse(resp), nil
}
