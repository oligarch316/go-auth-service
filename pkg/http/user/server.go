package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/oligarch316/go-auth-service/pkg/http"
	"github.com/oligarch316/go-auth-service/pkg/model"
)

// Server TODO.
type Server struct {
	Servelet *httpsvc.Servelet

	SignupValidater, UserValidater interface {
		Validate(*http.Request) (string, error)
	}

	Store interface {
		CreateUserAndDeleteInvite(inviteID, name, password string, mData model.UserUpdate) (model.User, error)

		CreateInvites(ownerID string, count int) ([]model.Invite, error)
		DeleteInvite(id string) error
		LookupInvites(ownerID string) ([]model.Invite, error)
		ReadInvite(id string) (model.Invite, error)

		DeleteUser(id string) error
		ReadUser(id string) (model.User, error)
		UpdateUser(id string, mData model.UserUpdate) error
	}
}

type userResponseBody struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Admin       bool   `json:"admin"`
}

// HandleUserCreate TODO.
func (s *Server) HandleUserCreate() httpsvc.Route {
	info := httpsvc.RouteInfo{
		Name:        "usercreate",
		Description: "create a new user (signup)",
		Method:      http.MethodPost,
		MetricTag:   "user_create",
	}

	type requestBody struct {
		Name        string  `json:"name"`
		Password    string  `json:"password"`
		DisplayName *string `json:"displayName"`
	}

	handle := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// Validate signup token
		inviteID, err := s.SignupValidater.Validate(r)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusUnauthorized, err, "failed to validate signup token"))
			return
		}

		var reqBody requestBody

		// Decode request body
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusBadRequest, err, "failed to load request body"))
			return
		}

		// Create user and delete invite
		user, err := s.Store.CreateUserAndDeleteInvite(
			inviteID,
			reqBody.Name,
			reqBody.Password,
			model.UserUpdate{DisplayName: reqBody.DisplayName},
		)

		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.StoreError(err, "failed to create user"))
			return
		}

		// Encode response body
		bytes, err := json.Marshal(userResponseBody{
			ID:          user.ID,
			Name:        user.Name,
			DisplayName: user.DisplayName,
			Admin:       user.Admin,
		})

		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.EncodeResponseError(err))
			return
		}

		// Respond
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write(bytes)
	}

	return httpsvc.Route{RouteInfo: info, Handle: handle}
}

// HandleUserRead TODO.
func (s *Server) HandleUserRead(userIDParamName string) httpsvc.Route {
	info := httpsvc.RouteInfo{
		Name:        "userread",
		Description: fmt.Sprintf("user data for id '%s'", userIDParamName),
		Method:      http.MethodGet,
		MetricTag:   "user_read",
	}

	handle := func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Validate user token
		userID, err := s.UserValidater.Validate(r)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusUnauthorized, err, "failed to validate user token"))
			return
		}

		// Confirm id match
		if userID != params.ByName(userIDParamName) {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(
				http.StatusForbidden,
				errors.New("user id does not match url parameter"),
				"failed to confirm user id",
			),
			)
			return
		}

		// Read user data
		user, err := s.Store.ReadUser(userID)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.StoreError(err, "failed to read user"))
			return
		}

		// Encode response body
		bytes, err := json.Marshal(userResponseBody{
			ID:          user.ID,
			Name:        user.Name,
			DisplayName: user.DisplayName,
			Admin:       user.Admin,
		})

		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.EncodeResponseError(err))
			return
		}

		// Respond
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}

	return httpsvc.Route{RouteInfo: info, Handle: handle}
}

// HandleUserUpdate TODO.
func (s *Server) HandleUserUpdate(userIDParamName string) httpsvc.Route {
	info := httpsvc.RouteInfo{
		Name:        "userupdate",
		Description: fmt.Sprintf("update user data for id '%s'", userIDParamName),
		Method:      http.MethodPatch,
		MetricTag:   "user_update",
	}

	type requestBody struct {
		DisplayName *string `json:"displayName"`
	}

	handle := func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Validate user token
		userID, err := s.UserValidater.Validate(r)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusUnauthorized, err, "failed to validate user token"))
			return
		}

		// Confirm id match
		if userID != params.ByName(userIDParamName) {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(
				http.StatusForbidden,
				errors.New("user id does not match url parameter"),
				"failed to confirm user id",
			),
			)
			return
		}

		var reqBody requestBody

		// Decode request body
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusBadRequest, err, "failed to load request body"))
			return
		}

		// Perform update
		if err = s.Store.UpdateUser(userID, model.UserUpdate{DisplayName: reqBody.DisplayName}); err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.StoreError(err, "failed to update user"))
			return
		}

		// Read user data
		user, err := s.Store.ReadUser(userID)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.StoreError(err, "failed to read user"))
			return
		}

		// Encode response body
		bytes, err := json.Marshal(userResponseBody{
			ID:          user.ID,
			Name:        user.Name,
			DisplayName: user.DisplayName,
			Admin:       user.Admin,
		})

		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.EncodeResponseError(err))
			return
		}

		// Respond
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}

	return httpsvc.Route{RouteInfo: info, Handle: handle}
}

// HandleUserDelete TODO.
// TODO: What about orphaned invites after this operation ???
func (s *Server) HandleUserDelete(userIDParamName string) httpsvc.Route {
	info := httpsvc.RouteInfo{
		Name:        "userdelete",
		Description: fmt.Sprintf("delete user with id '%s'", userIDParamName),
		Method:      http.MethodDelete,
		MetricTag:   "user_delete",
	}

	handle := func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Validate user token
		userID, err := s.UserValidater.Validate(r)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusUnauthorized, err, "failed to validate user token"))
			return
		}

		// Confirm id match
		if userID != params.ByName(userIDParamName) {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(
				http.StatusForbidden,
				errors.New("user id does not match url parameter"),
				"failed to confirm user id",
			),
			)
			return
		}

		// Perform delete
		if err = s.Store.DeleteUser(userID); err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.StoreError(err, "failed to delete user"))
			return
		}

		// Respond
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNoContent)
	}

	return httpsvc.Route{RouteInfo: info, Handle: handle}
}

type inviteResponseBody struct {
	ID      string `json:"id"`
	OwnerID string `json:"ownerID"`
}

// HandleInviteCreate TODO.
func (s *Server) HandleInviteCreate() httpsvc.Route {
	info := httpsvc.RouteInfo{
		Name:        "invitecreate",
		Description: "create a new invite",
		Method:      http.MethodPost,
		MetricTag:   "invite_create",
	}

	type (
		requestBody struct {
			Count   int    `json:"count"`
			OwnerID string `json:"ownerID"`
		}

		responseBody struct {
			Invites []inviteResponseBody `json:"invites"`
		}
	)

	handle := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// Validate user token
		userID, err := s.UserValidater.Validate(r)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusUnauthorized, err, "failed to validate user token"))
			return
		}

		// Read user data
		user, err := s.Store.ReadUser(userID)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.StoreError(err, "failed to read user"))
			return
		}

		// Assert user has admin privilages
		if !user.Admin {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusForbidden, errors.New("user is not an admin"), "failed to confirm admin privilages"))
			return
		}

		var reqBody requestBody

		// Decode request body
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusBadRequest, err, "failed to load request body"))
			return
		}

		// Create invites
		invites, err := s.Store.CreateInvites(reqBody.OwnerID, reqBody.Count)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.StoreError(err, "failed to create invites"))
			return
		}

		// Create response body
		respBody := responseBody{Invites: make([]inviteResponseBody, len(invites))}
		for i, invite := range invites {
			respBody.Invites[i] = inviteResponseBody{ID: invite.ID, OwnerID: invite.OwnerID}
		}

		// Encode response body
		bytes, err := json.Marshal(respBody)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.EncodeResponseError(err))
			return
		}

		// Respond
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write(bytes)
	}

	return httpsvc.Route{RouteInfo: info, Handle: handle}
}

// HandleInviteList TODO.
func (s *Server) HandleInviteList() httpsvc.Route {
	info := httpsvc.RouteInfo{
		Name:        "invitelist",
		Description: "list invites owned by user",
		Method:      http.MethodGet,
		MetricTag:   "invite_list",
	}

	type responseBody struct {
		Invites []inviteResponseBody `json:"invites"`
	}

	handle := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// Validate user token
		userID, err := s.UserValidater.Validate(r)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusUnauthorized, err, "failed to validate user token"))
			return
		}

		// Lookup invite data for user id
		invites, err := s.Store.LookupInvites(userID)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.StoreError(err, "failed to lookup invites for user"))
			return
		}

		// Create response body
		respBody := responseBody{Invites: make([]inviteResponseBody, len(invites))}
		for i, invite := range invites {
			respBody.Invites[i] = inviteResponseBody{ID: invite.ID, OwnerID: invite.OwnerID}
		}

		// Encode response body
		bytes, err := json.Marshal(respBody)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.EncodeResponseError(err))
			return
		}

		// Respond
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}

	return httpsvc.Route{RouteInfo: info, Handle: handle}
}

// HandleInviteRead TODO.
func (s *Server) HandleInviteRead(inviteIDParamName string) httpsvc.Route {
	info := httpsvc.RouteInfo{
		Name:        "inviteread",
		Description: fmt.Sprintf("invite data for id '%s'", inviteIDParamName),
		Method:      http.MethodGet,
		MetricTag:   "invite_read",
	}

	handle := func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Validate user token
		userID, err := s.UserValidater.Validate(r)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusUnauthorized, err, "failed to validate user token"))
			return
		}

		// Read invite id from url
		inviteID := params.ByName(inviteIDParamName)
		if inviteID == "" {
			s.Servelet.HandleErr(w, r, httpsvc.URLParamError(inviteIDParamName))
			return
		}

		// Read invite data
		invite, err := s.Store.ReadInvite(inviteID)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.StoreError(err, "failed to read invite data"))
			return
		}

		// Confirm ownership
		if userID != invite.OwnerID {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(
				http.StatusForbidden,
				errors.New("user is not the owner of invite"),
				"failed to confirm invite ownership",
			))
			return
		}

		// Encode response body
		bytes, err := json.Marshal(inviteResponseBody{ID: invite.ID, OwnerID: invite.OwnerID})
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.EncodeResponseError(err))
			return
		}

		// Respond
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}

	return httpsvc.Route{RouteInfo: info, Handle: handle}
}

// HandleInviteDelete TODO.
func (s *Server) HandleInviteDelete(inviteIDParamName string) httpsvc.Route {
	info := httpsvc.RouteInfo{
		Name:        "invitedelete",
		Description: fmt.Sprintf("delete invited with id '%s'", inviteIDParamName),
		Method:      http.MethodDelete,
		MetricTag:   "invite_delete",
	}

	handle := func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Validate user token
		userID, err := s.UserValidater.Validate(r)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusUnauthorized, err, "failed to validate user token"))
			return
		}

		// Read invite id from url
		inviteID := params.ByName(inviteIDParamName)
		if inviteID == "" {
			s.Servelet.HandleErr(w, r, httpsvc.URLParamError(inviteIDParamName))
			return
		}

		// Read invite data
		invite, err := s.Store.ReadInvite(inviteID)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.StoreError(err, "failed to read invite data"))
			return
		}

		// Confirm ownership
		if userID != invite.OwnerID {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(
				http.StatusForbidden,
				errors.New("user is not the owner of invite"),
				"failed to confirm invite ownership",
			))
			return
		}

		// Perform delete
		if err = s.Store.DeleteInvite(inviteID); err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.StoreError(err, "failed to delete invite"))
			return
		}

		// Respond
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNoContent)
	}

	return httpsvc.Route{RouteInfo: info, Handle: handle}
}
