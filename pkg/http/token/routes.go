package token

import "github.com/oligarch316/go-auth-service/pkg/http"

const (
	// APIVersion TODO.
	APIVersion = "v1"

	// DefaultAddress TODO.
	DefaultAddress = "localhost:8002"
)

const (
	pathBase = "/token"

	pathUser   = "/user"
	pathSignup = "/signup"
)

// AddRoutes TODO.
func (s *Server) AddRoutes(r *httpsvc.Router) {
	child := r.Child("/%s/%s", APIVersion, pathBase)

	child.Add(s.HandleUserCreate(), pathUser)
	child.Add(s.HandleUserRead(), pathUser)

	child.Add(s.HandleSignupCreate(), pathSignup)
	child.Add(s.HandleSignupRead(), pathSignup)
}
