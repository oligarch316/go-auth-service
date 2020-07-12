package secret

import "github.com/oligarch316/go-auth-service/pkg/http"

const (
	// APIVersion TODO
	APIVersion = "v1"

	// DefaultAddress TODO.
	DefaultAddress = "localhost:8001"
)

const (
	pathBase = "/secret"

	pathSet = "/set"
	pathKey = "/key"

	paramKeyID = "keyID"
)

// AddRoutes TODO
func (s *Server) AddRoutes(r *httpsvc.Router) {
	child := r.Child("/%s/%s", APIVersion, pathBase)

	child.Add(s.HandleSetRead(), pathSet)
	child.Add(s.HandleKeyList(), pathKey)
	child.Add(s.HandleKeyRead(paramKeyID), "/%s/%P", pathKey, paramKeyID)
}
