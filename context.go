package sgf

import (
	"net/http"
)

type response struct {
	HttpResp http.ResponseWriter
}
type Context struct {
	Response response
	Request  *http.Request
	Action   string
	Latency  int64
}

func (r response) SetHeaders(headers map[string]string, status_code int) {
	for k, v := range headers {
		r.HttpResp.Header().Set(k, v)
	}
	if 0 < status_code {
		r.HttpResp.WriteHeader(status_code)
	}
}
func (r response) End(data string) {
	panic("200" + data)
}
func (r response) Write(data string) {
	r.HttpResp.Write([]byte(data))
}
