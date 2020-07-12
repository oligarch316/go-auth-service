package user

import "github.com/oligarch316/go-auth-service/pkg/http"

const (
	// APIVersion TODO.
	APIVersion = "v1"

	// DefaultAddress TODO.
	DefaultAddress = "localhost:8003"
)

const (
	pathUser   = "/user"
	pathInvite = "/invite"

	paramUserID   = "userID"
	paramInviteID = "inviteID"
)

// AddRoutes TODO
func (s *Server) AddRoutes(r *httpsvc.Router) {
	child := r.Child("/%s", APIVersion)

	child.Add(s.HandleUserCreate(), pathUser)
	child.Add(s.HandleUserRead(paramUserID), "/%s/%P", pathUser, paramUserID)
	child.Add(s.HandleUserUpdate(paramUserID), "/%s/%P", pathUser, paramUserID)
	child.Add(s.HandleUserDelete(paramUserID), "/%s/%P", pathUser, paramUserID)

	child.Add(s.HandleInviteCreate(), pathInvite)
	child.Add(s.HandleInviteList(), pathInvite)
	child.Add(s.HandleInviteRead(paramInviteID), "/%s/%P", pathInvite, paramInviteID)
	child.Add(s.HandleInviteDelete(paramInviteID), "/%s/%P", pathInvite, paramInviteID)
}
