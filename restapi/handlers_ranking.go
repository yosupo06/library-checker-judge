package main

import (
	"context"
	"net/http"

	"github.com/yosupo06/library-checker-judge/database"
	restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
)

// GetRanking handles GET /ranking
func (s *server) GetRanking(_ context.Context, request restapi.GetRankingRequestObject) (restapi.GetRankingResponseObject, error) {
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
	results, total, err := database.FetchRanking(s.db, skip, limit)
	if err != nil {
		return nil, newHTTPError(http.StatusInternalServerError, "failed to fetch ranking")
	}
	stats := make([]restapi.UserStatistics, 0, len(results))
	for _, rs := range results {
		stats = append(stats, restapi.UserStatistics{Name: rs.UserName, Count: int32(rs.AcCount)})
	}
	resp := restapi.RankingResponse{Statistics: stats, Count: int32(total)}
	return restapi.GetRanking200JSONResponse(resp), nil
}
