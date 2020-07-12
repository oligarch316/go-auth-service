package httpsvc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"sort"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap/zapcore"
	"gopkg.in/alexcesaro/statsd.v2"
)

const (
	pathMetaRoute = "/route"

	paramRouteName = "name"
)

type pathArg struct{ arg interface{} }

func (pa pathArg) Format(s fmt.State, verb rune) {
	var newFormat string

	if verb == 'P' {
		newFormat = ":%s"
	} else {
		newFormat = string([]rune{'%', verb})
	}

	fmt.Fprintf(s, newFormat, pa.arg)
}

func formatPath(format string, a []interface{}) string {
	args := make([]interface{}, len(a))
	for i, arg := range a {
		args[i] = pathArg{arg}
	}
	return fmt.Sprintf(format, args...)
}

// RouteInfo TODO
type RouteInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Method      string `json:"method"`
	MetricTag   string `json:"metricTag"`
}

// MarshalLogObject TODO
func (ri RouteInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", ri.Name)
	enc.AddString("method", ri.Method)
	enc.AddString("metricTag", ri.MetricTag)
	return nil
}

type routeData struct {
	RouteInfo
	Path string `json:"path"`
}

func (rd routeData) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("path", rd.Path)
	return rd.RouteInfo.MarshalLogObject(enc)
}

// Route TODO
type Route struct {
	RouteInfo
	Handle httprouter.Handle
}

// Router TODO
type Router struct {
	prefix   string
	servelet *Servelet
	router   *httprouter.Router

	dataMap map[string]routeData
}

// NewRouter TODO
func NewRouter(servelet *Servelet, format string, a ...interface{}) *Router {
	return &Router{
		prefix:   path.Join("/", formatPath(format, a)),
		servelet: servelet,
		router:   httprouter.New(),
		dataMap:  make(map[string]routeData),
	}
}

// Child TODO
func (rtr *Router) Child(format string, a ...interface{}) *Router {
	return &Router{
		prefix:   path.Join(rtr.prefix, formatPath(format, a)),
		servelet: rtr.servelet,
		router:   rtr.router,
		dataMap:  rtr.dataMap,
	}
}

// Add TODO
func (rtr *Router) Add(route Route, format string, a ...interface{}) {
	// Format path
	pathStr := path.Join(rtr.prefix, formatPath(format, a))

	// Record route data
	rtr.dataMap[route.Name] = routeData{
		RouteInfo: route.RouteInfo,
		Path:      pathStr,
	}

	// Wrap handler for metric emission
	routeEmitter := rtr.servelet.Emitter.Clone(statsd.Tags("route", route.MetricTag))
	metricHandle := WrapMetrics(routeEmitter, route.Handle)

	// Add to httprouter
	rtr.router.Handle(route.Method, pathStr, metricHandle)
}

// ServeHTTP TODO
func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) { rtr.router.ServeHTTP(w, r) }

// MarshalLogArray TODO
func (rtr Router) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, item := range rtr.dataMap {
		if err := enc.AppendObject(item); err != nil {
			return err
		}
	}
	return nil
}

// HandleRouteList TODO
func (rtr *Router) HandleRouteList() Route {
	info := RouteInfo{
		Name:        "routelist",
		Description: "list of route information",
		Method:      http.MethodGet,
		MetricTag:   "route_list",
	}

	// TODO/BUG
	// Precomputing this means these meta routes aren't themselves included
	// in the route list...
	// Maybe a sync.Once is the answer?
	var resp struct {
		Routes []routeData `json:"routes"`
	}

	for _, item := range rtr.dataMap {
		resp.Routes = append(resp.Routes, item)
	}

	sort.Slice(resp.Routes, func(i, j int) bool { return resp.Routes[i].Name < resp.Routes[j].Name })

	handle := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		bytes, err := json.Marshal(resp)
		if err != nil {
			rtr.servelet.HandleErr(w, r, EncodeResponseError(err))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(bytes)
	}

	return Route{RouteInfo: info, Handle: handle}
}

// HandleRouteRead TODO
func (rtr *Router) HandleRouteRead(paramNameRouteName string) Route {
	info := RouteInfo{
		Name:        "routeread",
		Description: fmt.Sprintf("route information for '%s'", paramNameRouteName),
		Method:      http.MethodGet,
		MetricTag:   "route_read",
	}

	handle := func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		routeName := params.ByName(paramNameRouteName)
		if routeName == "" {
			rtr.servelet.HandleErr(w, r, URLParamError(paramNameRouteName))
			return
		}

		data, ok := rtr.dataMap[routeName]
		if !ok {
			err := fmt.Errorf("no such route '%s'", routeName)
			rtr.servelet.HandleErr(w, r, NewError(http.StatusNotFound, err, "failed to load route data"))
			return
		}

		bytes, err := json.Marshal(data)
		if err != nil {
			rtr.servelet.HandleErr(w, r, EncodeResponseError(err))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(bytes)
	}

	return Route{RouteInfo: info, Handle: handle}
}

// AddMetaRoutes TODO
func (rtr *Router) AddMetaRoutes() {
	rtr.Add(rtr.HandleRouteList(), pathMetaRoute)
	rtr.Add(rtr.HandleRouteRead(paramRouteName), "%s/%P", pathMetaRoute, paramRouteName)
}
