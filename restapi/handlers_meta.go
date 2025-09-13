package main

import (
    "encoding/json"
    "net/http"

    "github.com/yosupo06/library-checker-judge/database"
    "github.com/yosupo06/library-checker-judge/langs"
    restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
)

// GetLangList handles GET /langs
func (s *server) GetLangList(w http.ResponseWriter, r *http.Request) {
    var ls []restapi.Lang
    for _, l := range langs.LANGS {
        ls = append(ls, restapi.Lang{Id: l.ID, Name: l.Name, Version: l.Version})
    }
    resp := restapi.LangListResponse{Langs: ls}
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(resp)
}

// GetProblemCategories handles GET /categories
func (s *server) GetProblemCategories(w http.ResponseWriter, r *http.Request) {
    cats, err := database.FetchProblemCategories(s.db)
    if err != nil {
        http.Error(w, "failed to fetch categories", http.StatusInternalServerError)
        return
    }
    result := make([]restapi.ProblemCategory, 0, len(cats))
    for _, c := range cats {
        result = append(result, restapi.ProblemCategory{Title: c.Title, Problems: c.Problems})
    }
    resp := restapi.ProblemCategoriesResponse{Categories: result}
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(resp)
}
