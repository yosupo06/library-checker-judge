package main

import (
    "encoding/json"
    "net/http"

    "github.com/yosupo06/library-checker-judge/database"
    restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
)

// GetRanking handles GET /ranking
func (s *server) GetRanking(w http.ResponseWriter, r *http.Request, params restapi.GetRankingParams) {
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
    results, total, err := database.FetchRanking(s.db, skip, limit)
    if err != nil {
        http.Error(w, "failed to fetch ranking", http.StatusInternalServerError)
        return
    }
    stats := make([]restapi.UserStatistics, 0, len(results))
    for _, rs := range results {
        stats = append(stats, restapi.UserStatistics{Name: rs.UserName, Count: int32(rs.AcCount)})
    }
    resp := restapi.RankingResponse{Statistics: stats, Count: int32(total)}
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(resp)
}
