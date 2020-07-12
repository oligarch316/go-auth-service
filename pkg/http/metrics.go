package httpsvc

import (
	"io"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/alexcesaro/statsd.v2"
)

type latencyRecorder time.Time

func (lr latencyRecorder) emit(c *statsd.Client) {
	dur := time.Now().Sub(time.Time(lr))
	c.Timing("requests.latency", int(dur/time.Millisecond))
}

type payloadRecorder struct {
	io.ReadCloser
	size int
}

func (pr payloadRecorder) emit(c *statsd.Client) {
	if pr.size > 0 {
		c.Histogram("requests.payload_size", pr.size)
	}
}

func (pr *payloadRecorder) Read(p []byte) (int, error) {
	n, err := pr.ReadCloser.Read(p)
	pr.size = pr.size + n
	return n, err
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr statusRecorder) tag(c *statsd.Client) *statsd.Client {
	var className string

	switch {
	case sr.status < 100:
		className = "unknown"
	case sr.status < 200:
		className = "1XX"
	case sr.status < 300:
		className = "2XX"
	case sr.status < 400:
		className = "3XX"
	case sr.status < 500:
		className = "4XX"
	case sr.status < 600:
		className = "5XX"
	default:
		className = "unknown"
	}

	return c.Clone(statsd.Tags("status_class", className))
}

func (sr *statusRecorder) WriteHeader(statusCode int) {
	sr.status = statusCode
	sr.ResponseWriter.WriteHeader(statusCode)
}

func (sr *statusRecorder) Write(data []byte) (int, error) {
	if sr.status == 0 {
		sr.status = http.StatusOK
	}
	return sr.ResponseWriter.Write(data)
}

// WrapMetrics TODO
func WrapMetrics(client *statsd.Client, handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		var (
			latencyR = latencyRecorder(time.Now())
			payloadR = &payloadRecorder{ReadCloser: r.Body}
			statusR  = &statusRecorder{ResponseWriter: w}
		)

		r.Body = payloadR

		handle(statusR, r, params)

		latencyR.emit(statusR.tag(client))
		payloadR.emit(client)
	}
}
