package token

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/oligarch316/go-auth-service/internal/pkg/claims"
	"github.com/oligarch316/go-auth-service/pkg/http"
	"github.com/oligarch316/go-auth-service/pkg/http/secret/token"
	"github.com/oligarch316/go-auth-service/pkg/model"
	"github.com/oligarch316/go-skeleton/pkg/config/types"
	"go.uber.org/zap/zapcore"
)

// ConfigAudienceNames TODO.
type ConfigAudienceNames struct {
	User   string `json:"user"`
	Signup string `json:"signup"`
}

// DefaultAudienceNamesConfig TODO.
func DefaultAudienceNamesConfig() ConfigAudienceNames {
	return ConfigAudienceNames{
		Signup: claims.DefaultAudienceNameSignup,
		User:   claims.DefaultAudienceNameUser,
	}
}

// MarshalLogObject TODO.
func (can ConfigAudienceNames) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("user", can.User)
	enc.AddString("signup", can.Signup)
	return nil
}

// ConfigServer TODO.
type ConfigServer struct {
	AudienceNames ConfigAudienceNames `json:"audienceNames"`
	IssuerName    string              `json:"issuerName"`
	MaxTTL        ctype.Duration      `json:"maxTTL"`
}

// DefaultServerConfig TODO.
func DefaultServerConfig() ConfigServer {
	return ConfigServer{
		AudienceNames: DefaultAudienceNamesConfig(),
		IssuerName:    claims.DefaultIssuerName,
		MaxTTL:        ctype.Duration{Duration: 24 * time.Hour},
	}
}

// MarshalLogObject TODO.
func (cs ConfigServer) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddDuration("maxTTL", cs.MaxTTL.Duration)
	enc.AddString("issuerName", cs.IssuerName)
	return enc.AddObject("audienceNames", cs.AudienceNames)
}

// Server TODO.
type Server struct {
	ConfigServer

	Servelet *httpsvc.Servelet

	Secret interface {
		Sign(claims interface{}) (string, error)
		Validate(token string, claims interface{}) error
	}

	Store interface {
		LookupUser(name string) (model.User, error)
		ReadInvite(id string) (model.Invite, error)
	}
}

func (s Server) claimsGenFactory(audienceNames ...string) func(string, time.Duration) token.StandardClaims {
	allowedAud := ctype.NewStringSet(audienceNames...)

	return func(id string, ttl time.Duration) token.StandardClaims {
		if ttl <= 0 || ttl > s.MaxTTL.Duration {
			ttl = s.MaxTTL.Duration
		}

		now := time.Now()

		return token.StandardClaims{
			Issuer:     s.IssuerName,
			Subject:    id,
			Audience:   allowedAud,
			Expiration: &token.NumericDate{Time: now.Add(ttl)},
			IssuedAt:   &token.NumericDate{Time: now},
		}
	}
}

func (s Server) claimsValFactory(audienceName string) func(*http.Request) (string, error) {
	return token.Validater{
		Claims: token.ConfigValidater{
			AllowedIssuers: ctype.NewStringSet(s.IssuerName),
			AudienceName:   audienceName,
		},
		Secret: s.Secret,
	}.Validate
}

type tokenResponseBody struct {
	ID         string             `json:"id"`
	Token      string             `json:"token"`
	Expiration *token.NumericDate `json:"expiration"`
}

// HandleUserCreate TODO.
func (s *Server) HandleUserCreate() httpsvc.Route {
	info := httpsvc.RouteInfo{
		Name:        "tokenusercreate",
		Description: "create a new user token (login)",
		Method:      http.MethodPost,
		MetricTag:   "token_user_create",
	}

	type requestBody struct {
		Name     string         `json:"name"`
		Password string         `json:"password"`
		TTL      ctype.Duration `json:"ttl"`
	}

	genClaims := s.claimsGenFactory(s.AudienceNames.User)

	handle := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var reqBody requestBody

		// Decode request body
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusBadRequest, err, "failed to load request body"))
			return
		}

		// Lookup user data by name
		data, err := s.Store.LookupUser(reqBody.Name)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.StoreError(err, "failed to lookup user"))
			return
		}

		// Valiate given password against existing user data
		if err = data.PasswordHash.Compare(reqBody.Password); err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusForbidden, err, "failed to validate password"))
			return
		}

		// Create user claims
		claims := genClaims(data.ID, reqBody.TTL.Duration)

		// Build token from claims
		token, err := s.Secret.Sign(claims)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.InternalError(err, "failed to sign token claims"))
			return
		}

		// Encode response body
		bytes, err := json.Marshal(tokenResponseBody{
			ID:         data.ID,
			Token:      token,
			Expiration: claims.Expiration,
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

// HandleUserRead TODO
func (s *Server) HandleUserRead() httpsvc.Route {
	info := httpsvc.RouteInfo{
		Name:        "tokenuserread",
		Description: "validate attached user token",
		Method:      http.MethodGet,
		MetricTag:   "token_user_read",
	}

	valClaims := s.claimsValFactory(s.AudienceNames.User)

	handle := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// Validate user token
		if _, err := valClaims(r); err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusUnauthorized, err, "failed to validate user token"))
			return
		}

		// Respond
		w.WriteHeader(http.StatusOK)
	}

	return httpsvc.Route{RouteInfo: info, Handle: handle}
}

// HandleSignupCreate TODO
func (s *Server) HandleSignupCreate() httpsvc.Route {
	info := httpsvc.RouteInfo{
		Name:        "tokensignupcreate",
		Description: "create a new signup token",
		Method:      http.MethodPost,
		MetricTag:   "token_signup_create",
	}

	type requestBody struct {
		InviteID string         `json:"inviteID"`
		TTL      ctype.Duration `json:"ttl"`
	}

	var (
		valUserClaims   = s.claimsValFactory(s.AudienceNames.User)
		genSignupClaims = s.claimsGenFactory(s.AudienceNames.Signup)
	)

	handle := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// Validate user token
		userID, err := valUserClaims(r)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusUnauthorized, err, "failed to validate user token"))
			return
		}

		var reqBody requestBody

		// Decode request body
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusBadRequest, err, "failed to load request body"))
			return
		}

		// Lookup invite by id
		inviteData, err := s.Store.ReadInvite(reqBody.InviteID)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.StoreError(err, "failed to read invite"))
			return
		}

		// Ensure invite ownership
		if inviteData.OwnerID != userID {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusForbidden, errors.New("forbidden"), "failed to read invite"))
			return
		}

		// Create signup claims
		signupClaims := genSignupClaims(inviteData.ID, reqBody.TTL.Duration)

		// Build token from claims
		signupToken, err := s.Secret.Sign(signupClaims)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.InternalError(err, "failed to sign signup token claims"))
			return
		}

		// Encode response body
		bytes, err := json.Marshal(tokenResponseBody{
			ID:         inviteData.ID,
			Token:      signupToken,
			Expiration: signupClaims.Expiration,
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

// HandleSignupRead TODO
func (s *Server) HandleSignupRead() httpsvc.Route {
	info := httpsvc.RouteInfo{
		Name:        "tokensignupread",
		Description: "validate attached signup token",
		Method:      http.MethodGet,
		MetricTag:   "token_signup_read",
	}

	valClaims := s.claimsValFactory(s.AudienceNames.Signup)

	handle := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// Validate signup token
		if _, err := valClaims(r); err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusUnauthorized, err, "failed to validate signup token"))
			return
		}

		// Respond
		w.WriteHeader(http.StatusOK)
	}

	return httpsvc.Route{RouteInfo: info, Handle: handle}
}
