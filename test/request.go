package test

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
)

// Request encapsulates http request and response logger testing util for more convenient web integration testing
type Request struct {
	Response *httptest.ResponseRecorder
	request  *http.Request
}

func newRequest(request *http.Request) *Request {
	return &Request{Response: httptest.NewRecorder(), request: request}
}

// Do executes mock http request
func (r *Request) Do(e *gin.Engine) *Request {
	e.ServeHTTP(r.Response, r.request)

	return r
}
