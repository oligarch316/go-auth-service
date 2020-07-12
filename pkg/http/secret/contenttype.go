package secret

import (
	"net/http"

	"github.com/golang/gddo/httputil"
)

const (
	cTypeJSON contentType = "application/json"
	cTypePEM  contentType = "application/x-pem-file"
)

type contentType string

func (ct contentType) writeReqHeader(r *http.Request) {
	r.Header.Set("Accept", string(ct))
}

func (ct contentType) writeRespHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", string(ct)+"; charset=utf-8")
}

func allowedContentTypes(items ...contentType) func(*http.Request) contentType {
	strs := make([]string, len(items))
	for i, item := range items {
		strs[i] = string(item)
	}

	return func(r *http.Request) contentType {
		res := httputil.NegotiateContentType(r, strs, strs[0])
		return contentType(res)
	}
}
