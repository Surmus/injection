package injection

import (
	"net/http"
	"reflect"
)

type HandlerFunc func(interface{})

type Provider interface{}

type Handler interface{}

type HttpHandlerFunc func(http.ResponseWriter, *http.Request)

type registeredProvider func(ctxValues resolvedValues) interface{}

func staticValueRegisterProvider(value reflect.Value) registeredProvider {
	return func(ctxValues resolvedValues) interface{} {
		return value.Interface()
	}
}

type Controller interface {
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
