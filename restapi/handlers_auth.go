package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yosupo06/library-checker-judge/database"
	restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
	"gorm.io/gorm"
)

// PostRegister handles POST /auth/register
func (s *server) PostRegister(w http.ResponseWriter, r *http.Request) {
	var req restapi.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	uid, err := s.uidFromRequest(r)
	if err != nil || uid == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := database.RegisterUser(s.db, req.Name, uid); err != nil {
		http.Error(w, "register failed", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(restapi.RegisterResponse{})
}

// GetCurrentUserInfo handles GET /auth/current_user
func (s *server) GetCurrentUserInfo(w http.ResponseWriter, r *http.Request) {
	uid, err := s.uidFromRequest(r)
	if err != nil || uid == "" {
		// Return empty user (not logged in)
		_ = json.NewEncoder(w).Encode(restapi.CurrentUserInfoResponse{})
		return
	}
	user, err := database.FetchUserFromUID(s.db, uid)
	if err != nil || user == nil {
		_ = json.NewEncoder(w).Encode(restapi.CurrentUserInfoResponse{})
		return
	}
	resp := restapi.CurrentUserInfoResponse{User: &restapi.User{
		Name:        user.Name,
		LibraryUrl:  user.LibraryURL,
		IsDeveloper: user.IsDeveloper,
	}}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// PatchCurrentUserInfo handles PATCH /auth/current_user
func (s *server) PatchCurrentUserInfo(w http.ResponseWriter, r *http.Request) {
	var req restapi.ChangeCurrentUserInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.User.Name == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	uid, err := s.uidFromRequest(r)
	if err != nil || uid == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := database.UpdateUser(s.db, database.User{
		Name:        req.User.Name,
		UID:         uid,
		LibraryURL:  req.User.LibraryUrl,
		IsDeveloper: req.User.IsDeveloper,
	}); err != nil {
		http.Error(w, "update failed", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(restapi.ChangeCurrentUserInfoResponse{})
}

// GetUserInfo handles GET /users/{name}
func (s *server) GetUserInfo(w http.ResponseWriter, r *http.Request, name string) {
	if name == "" {
		http.Error(w, "empty name", http.StatusBadRequest)
		return
	}
	user, err := database.FetchUserFromName(s.db, name)
	if err != nil || user == nil {
		http.Error(w, "invalid user name", http.StatusBadRequest)
		return
	}
	// Build solved_map similar to gRPC implementation
	stats, err := fetchUserStatisticsREST(s.db, name)
	if err != nil {
		http.Error(w, "failed to fetch statistics", http.StatusInternalServerError)
		return
	}
	resp := restapi.UserInfoResponse{
		User: restapi.User{
			Name:        user.Name,
			LibraryUrl:  user.LibraryURL,
			IsDeveloper: user.IsDeveloper,
		},
		SolvedMap: stats,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// PatchUserInfo handles PATCH /users/{name}
func (s *server) PatchUserInfo(w http.ResponseWriter, r *http.Request, name string) {
	var req restapi.ChangeUserInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	uid, err := s.uidFromRequest(r)
	if err != nil || uid == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	current, err := database.FetchUserFromUID(s.db, uid)
	if err != nil || current == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if current.Name != name {
		http.Error(w, "permission denied", http.StatusForbidden)
		return
	}
	// Only allow updating library_url and is_developer
	u := database.User{
		Name:        name,
		LibraryURL:  req.User.LibraryUrl,
		IsDeveloper: req.User.IsDeveloper,
	}
	if err := database.UpdateUser(s.db, u); err != nil {
		http.Error(w, "update failed", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(restapi.ChangeUserInfoResponse{})
}

func fetchUserStatisticsREST(db *gorm.DB, userName string) (map[string]string, error) {
	type Result struct {
		ProblemName string
		LatestAC    bool
	}
	var results = make([]Result, 0)
	if err := db.
		Model(&database.Submission{}).
		Joins("left join problems on submissions.problem_name = problems.name").
		Select("problem_name, bool_or(submissions.test_cases_version=problems.test_cases_version) as latest_ac").
		Where("status = 'AC' and user_name = ?", userName).
		Group("problem_name").
		Find(&results).Error; err != nil {
		return nil, errors.New("failed sql query")
	}
	stats := make(map[string]string)
	for _, result := range results {
		if result.LatestAC {
			stats[result.ProblemName] = "LATEST_AC"
		} else {
			stats[result.ProblemName] = "AC"
		}
	}
	return stats, nil
}
