package test

import (
	"bytes"
	"github.com/gin-gonic/gin/json"
	"net/http"
)

type RequestBuilder struct {
	method  string
	url     string
	content []byte
	headers map[string]string
}

// NewRequest creates new instance of the RequestBuilder
func NewRequest(url string, method string) *RequestBuilder {
	return &RequestBuilder{url: url, method: method, headers: map[string]string{}}
}

// JsonContent taken value and binds it to request as json
func (builder *RequestBuilder) JsonContent(value interface{}) *RequestBuilder {
	jsonString, _ := json.Marshal(value)

	builder.content = []byte(jsonString)
	builder.Header("Content-Type", "application/json")

	return builder
}

// Header add/replaces http request header to the request being built
func (builder *RequestBuilder) Header(name string, value string) *RequestBuilder {
	builder.headers[name] = value

	return builder
}

// MustBuild composes request from builder variables and panics when encounters error
func (builder *RequestBuilder) MustBuild() *Request {
	request, err := http.NewRequest(builder.method, builder.url, bytes.NewBuffer(builder.content))

	if err != nil {
		panic(err)
	}

	for key, val := range builder.headers {
		request.Header.Set(key, val)
	}

	return newRequest(request)
}
