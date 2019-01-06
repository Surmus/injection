package injection

import (
	"net/http"
	"reflect"
)

// Provider should be function returning one value, used for registering value providers with Injector RegisterProviders method.
// See test.type.go file for examples
type Provider interface{}

// Handler should be function with void return type, used for registering http handlers with Injector Handle and Use methods.
// See test.type.go file for examples
type Handler interface{}

type HttpHandlerFunc func(http.ResponseWriter, *http.Request)

type registeredProvider func(ctxValues resolvedValues) interface{}

func staticValueRegisterProvider(value reflect.Value) registeredProvider {
	return func(ctxValues resolvedValues) interface{} {
		return value.Interface()
	}
}

// Controller should contain type methods which are used for http request handling
// request handler methods should be exported(start with capital character)
// request handler http method is determined from the first part of the method name,
// method name should contain http method meant to be used, name parts should be formatted by camelcase standard
// example/supported method names:
// - PostMethodName - handles POST request
// - GetMethodName - handles GET request
// - DeleteMethodName - handles DELETE request
// - PutMethodName - handles PUT request
// - ConnectMethodName - handles CONNECT request
// - HeadMethodName - handles HEAD request
// - OptionsMethodName - handles OPTIONS request
// - PatchMethodName - handles PATCH request
// - TraceMethodName - handles TRACE request
// when method name does not match any known http method, GET is used
// See test.type.go file for examples
type Controller interface {
	// Routes should return mapping of http route endpoint / Controller method
	// example: map[string]string{"/test-http-path": "PostMethodName"}
	// "/test-http-path" being http request endpoint and "PostMethodName" method on the Controller method used as request handler
	Routes() map[string]string
}

type resolvedValues map[reflect.Type]reflect.Value

type typedProvider struct {
	kind  reflect.Type
	value registeredProvider
}

func newTypedProvider(providerType reflect.Type, provider registeredProvider) *typedProvider {
	return &typedProvider{kind: providerType, value: provider}
}
