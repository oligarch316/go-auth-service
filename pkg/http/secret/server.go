package secret

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/knq/pemutil"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/oligarch316/go-auth-service/pkg/http"
)

// Server TODO
type Server struct {
	Servelet *httpsvc.Servelet
	Set      *jwk.Set
}

type keyListResponse struct {
	KeyIDs []string `json:"keyIDs"`
}

// HandleSetRead TODO
func (s *Server) HandleSetRead() httpsvc.Route {
	info := httpsvc.RouteInfo{
		Name:        "secretsetread",
		Description: "JSON web key (jwk) set",
		Method:      http.MethodGet,
		MetricTag:   "secret_set_read",
	}

	handle := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		bytes, err := json.Marshal(s.Set)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.EncodeResponseError(err))
			return
		}

		cTypeJSON.writeRespHeader(w)
		w.Write(bytes)
	}

	return httpsvc.Route{RouteInfo: info, Handle: handle}
}

// HandleKeyList TODO
func (s *Server) HandleKeyList() httpsvc.Route {
	info := httpsvc.RouteInfo{
		Name:        "secretkeylist",
		Description: "list of JSON web key (jwk) key ids",
		Method:      http.MethodGet,
		MetricTag:   "secret_key_list",
	}

	var resp keyListResponse

	for _, key := range s.Set.Keys {
		resp.KeyIDs = append(resp.KeyIDs, key.KeyID())
	}

	handle := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		bytes, err := json.Marshal(resp)
		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.EncodeResponseError(err))
			return
		}

		cTypeJSON.writeRespHeader(w)
		w.Write(bytes)
	}

	return httpsvc.Route{RouteInfo: info, Handle: handle}
}

// HandleKeyRead TODO
func (s *Server) HandleKeyRead(kidParamName string) httpsvc.Route {
	info := httpsvc.RouteInfo{
		Name:        "secretkeyread",
		Description: fmt.Sprintf("JSON web key (jwk) for id '%s'", kidParamName),
		Method:      http.MethodGet,
		MetricTag:   "secret_key_read",
	}

	negotiateCType := allowedContentTypes(cTypeJSON, cTypePEM)

	handle := func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		var (
			ct  = negotiateCType(r)
			kid = params.ByName(kidParamName)
		)

		if kid == "" {
			s.Servelet.HandleErr(w, r, httpsvc.URLParamError(kidParamName))
			return
		}

		keys := s.Set.LookupKeyID(kid)
		if len(keys) < 1 {
			err := fmt.Errorf("no such key '%s'", kid)
			s.Servelet.HandleErr(w, r, httpsvc.NewError(http.StatusNotFound, err, "failed to load key"))
			return
		}

		var (
			bytes []byte
			err   error
		)

		switch ct {
		case cTypeJSON:
			bytes, err = json.Marshal(keys[0])
		case cTypePEM:
			var item interface{}
			if err = keys[0].Raw(&item); err == nil {
				bytes, err = pemutil.EncodePrimitive(item)
			}
		default:
			err = fmt.Errorf("unexpected content-type '%s'", ct)
		}

		if err != nil {
			s.Servelet.HandleErr(w, r, httpsvc.EncodeResponseError(err))
			return
		}

		ct.writeRespHeader(w)
		w.Write(bytes)
	}

	return httpsvc.Route{RouteInfo: info, Handle: handle}
}
