package test

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
)

type Request struct {
	Response *httptest.ResponseRecorder
	request  *http.Request
}

func newRequest(request *http.Request) *Request {
	return &Request{Response: httptest.NewRecorder(), request: request}
}

func (r *Request) Do(e *gin.Engine) *Request {
	e.ServeHTTP(r.Response, r.request)

	return r
}
