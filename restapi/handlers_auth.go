package main

import (
	"context"
	"net/http"

	"github.com/yosupo06/library-checker-judge/database"
	restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
	"gorm.io/gorm"
)

// PostRegister handles POST /auth/register
func (s *server) PostRegister(ctx context.Context, request restapi.PostRegisterRequestObject) (restapi.PostRegisterResponseObject, error) {
	if request.Body == nil || request.Body.Name == "" {
		return nil, newHTTPError(http.StatusBadRequest, "invalid request")
	}
	uid, err := s.uidFromContext(ctx)
	if err != nil || uid == "" {
		return nil, newHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if err := database.RegisterUser(s.db, request.Body.Name, uid); err != nil {
		return nil, newHTTPError(http.StatusBadRequest, "register failed")
	}
	return restapi.PostRegister200JSONResponse(restapi.RegisterResponse{}), nil
}

// GetCurrentUserInfo handles GET /auth/current_user
func (s *server) GetCurrentUserInfo(ctx context.Context, _ restapi.GetCurrentUserInfoRequestObject) (restapi.GetCurrentUserInfoResponseObject, error) {
	uid, err := s.uidFromContext(ctx)
	if err != nil || uid == "" {
		return restapi.GetCurrentUserInfo200JSONResponse(restapi.CurrentUserInfoResponse{}), nil
	}
	user, err := database.FetchUserFromUID(s.db, uid)
	if err != nil || user == nil {
		return restapi.GetCurrentUserInfo200JSONResponse(restapi.CurrentUserInfoResponse{}), nil
	}
	resp := restapi.CurrentUserInfoResponse{User: &restapi.User{
		Name:        user.Name,
		LibraryUrl:  user.LibraryURL,
		IsDeveloper: user.IsDeveloper,
	}}
	return restapi.GetCurrentUserInfo200JSONResponse(resp), nil
}

// PatchCurrentUserInfo handles PATCH /auth/current_user
func (s *server) PatchCurrentUserInfo(ctx context.Context, request restapi.PatchCurrentUserInfoRequestObject) (restapi.PatchCurrentUserInfoResponseObject, error) {
	if request.Body == nil || request.Body.User.Name == "" {
		return nil, newHTTPError(http.StatusBadRequest, "invalid request")
	}
	uid, err := s.uidFromContext(ctx)
	if err != nil || uid == "" {
		return nil, newHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	user := database.User{
		Name:        request.Body.User.Name,
		UID:         uid,
		LibraryURL:  request.Body.User.LibraryUrl,
		IsDeveloper: request.Body.User.IsDeveloper,
	}
	if err := s.updateUser(s.db, user); err != nil {
		return nil, newHTTPError(http.StatusBadRequest, "update failed")
	}
	return restapi.PatchCurrentUserInfo200JSONResponse(restapi.ChangeCurrentUserInfoResponse{}), nil
}

// GetUserInfo handles GET /users/{name}
func (s *server) GetUserInfo(_ context.Context, request restapi.GetUserInfoRequestObject) (restapi.GetUserInfoResponseObject, error) {
	if request.Name == "" {
		return nil, newHTTPError(http.StatusBadRequest, "empty name")
	}
	user, err := database.FetchUserFromName(s.db, request.Name)
	if err != nil || user == nil {
		return nil, newHTTPError(http.StatusBadRequest, "invalid user name")
	}
	resp := restapi.UserInfoResponse{
		User: restapi.User{
			Name:        user.Name,
			LibraryUrl:  user.LibraryURL,
			IsDeveloper: user.IsDeveloper,
		},
	}
	return restapi.GetUserInfo200JSONResponse(resp), nil
}

// GetUserStatistics handles GET /users/{name}/statistics
func (s *server) GetUserStatistics(_ context.Context, request restapi.GetUserStatisticsRequestObject) (restapi.GetUserStatisticsResponseObject, error) {
	if request.Name == "" {
		return nil, newHTTPError(http.StatusBadRequest, "empty name")
	}
	stats, err := fetchUserStatisticsREST(s.db, request.Name)
	if err != nil {
		return nil, newHTTPError(http.StatusInternalServerError, "failed to fetch statistics")
	}
	resp := restapi.UserSolvedStatisticsResponse{SolvedMap: stats}
	return restapi.GetUserStatistics200JSONResponse(resp), nil
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
