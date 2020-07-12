package httpsvc

import (
	"fmt"
	"net/http"

	"github.com/oligarch316/go-skeleton/pkg/observ"
	"go.uber.org/zap"
)

// Error TODO
type Error struct {
	Status  int
	Message string
	error
}

// Unwrap TODO
func (e Error) Unwrap() error { return e.error }

// NewError TODO
func NewError(status int, err error, format string, a ...interface{}) Error {
	return Error{
		Status:  status,
		Message: fmt.Sprintf(format, a...),
		error:   err,
	}
}

// InternalError TODO
func InternalError(err error, format string, a ...interface{}) Error {
	return NewError(http.StatusInternalServerError, err, format, a...)
}

// EncodeResponseError TODO
func EncodeResponseError(err error) Error {
	return InternalError(err, "failed to encode response")
}

// URLParamError TODO
func URLParamError(paramName string) Error {
	return InternalError(fmt.Errorf("invalid parameter data for '%s'", paramName), "failed to parse url parameter")
}

// StoreError TODO
func StoreError(err error, format string, a ...interface{}) Error {
	var status int

	// TODO: Determine appropriate status code
	status = http.StatusInternalServerError

	return NewError(status, err, format, a...)
}

// Servelet TODO
type Servelet struct{ *observ.Corelet }

// Named TODO
func (s *Servelet) Named(name string) *Servelet {
	return &Servelet{Corelet: s.Corelet.Named(name)}
}

// HandleErr TODO
func (s *Servelet) HandleErr(w http.ResponseWriter, r *http.Request, err Error) {
	s.logErr(r, err)
	s.writeErr(w, err)
}

func (s *Servelet) logErr(r *http.Request, err Error) {
	var logFunc func(string, ...zap.Field)
	if err.Status >= http.StatusInternalServerError {
		logFunc = s.Logger.Error
	} else {
		logFunc = s.Logger.Debug
	}

	logFunc(err.Message,
		zap.String("method", r.Method),
		zap.String("remoteAddr", r.RemoteAddr),
		zap.String("requestURI", r.RequestURI),
		zap.Int("status", err.Status),
		zap.Error(err.error),
	)
}

func (s *Servelet) writeErr(w http.ResponseWriter, err Error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(err.Status)
	fmt.Fprintf(w, `{ "message": "%s", "error": "%s" }`, err.Message, err.Error())
}
