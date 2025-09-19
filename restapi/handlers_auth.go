package main

import (
	"encoding/json"
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
	if err := s.updateUser(s.db, database.User{
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
	resp := restapi.UserInfoResponse{
		User: restapi.User{
			Name:        user.Name,
			LibraryUrl:  user.LibraryURL,
			IsDeveloper: user.IsDeveloper,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// GetUserStatistics handles GET /users/{name}/statistics
func (s *server) GetUserStatistics(w http.ResponseWriter, r *http.Request, name string) {
	if name == "" {
		http.Error(w, "empty name", http.StatusBadRequest)
		return
	}
	stats, err := fetchUserStatisticsREST(s.db, name)
	if err != nil {
		http.Error(w, "failed to fetch statistics", http.StatusInternalServerError)
		return
	}
	resp := restapi.UserSolvedStatisticsResponse{SolvedMap: stats}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func fetchUserStatisticsREST(db *gorm.DB, userName string) (map[string]restapi.SolvedStatus, error) {
	statuses, err := database.FetchUserSolvedStatuses(db, userName)
	if err != nil {
		return nil, err
	}
	stats := make(map[string]restapi.SolvedStatus, len(statuses))
	for _, status := range statuses {
		if status.LatestAC {
			stats[status.ProblemName] = restapi.LATESTAC
		} else {
			stats[status.ProblemName] = restapi.AC
		}
	}
	return stats, nil
}
