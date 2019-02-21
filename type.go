package injection

import (
	"reflect"
)

// Provider should be function returning one value, used for registering value providers with Injector RegisterProviders method.
// See injector_test.go file for examples
type Provider interface{}

type singletonProvider struct {
	provider Provider
}

// NewSingletonProvider instructs the Injector to resolve given provider param to resolve only once
// All successive resolve request return the value resolved at the first time
func NewSingletonProvider(provider Provider) *singletonProvider {
	return &singletonProvider{provider: provider}
}

// Handler should be function with void return type, used for registering http handlers with Injector Handle and Use methods.
// See injector_test.go file for examples
type Handler interface{}

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
// See injector_test.go file for examples
type Controller interface {
	// Routes should return mapping of http route endpoint / Controller methods
	// example: map[string][]string{"/test-http-path": {"PostMethodName"}}
	// "/test-http-path" being http request endpoint and "PostMethodName" as request handler for POST http method
	Routes() map[string][]string

	// Middleware is used to tell injector which methods use middleware in their request handling chain
	// should return map of Controller method / middleware functions slice
	// example: map[string][]Handler{"PostMethodName": {func() { fmt.Println("middleware executed") }}}
	Middleware() map[string][]Handler
}

// BaseController can be embedded into Controller implementations in order to skip implementing Middleware interface method
type BaseController struct{}

// Middleware tells injector that Controller routes do not use any middleware
func (c *BaseController) Middleware() map[string][]Handler {
	return make(map[string][]Handler)
}

type resolvedValues map[reflect.Type]reflect.Value

type typedProvider struct {
	kind  reflect.Type
	value registeredProvider
}

func newTypedProvider(providerType reflect.Type, provider registeredProvider) *typedProvider {
	return &typedProvider{kind: providerType, value: provider}
}

type controllerRoute struct {
	route      string
	methodName string
}
