package injection

import (
	"net/http"
	"reflect"
)

type Http interface {
	HandleFunc(pattern string, handler Handler) (Http, error)
}

type httpImpl struct {
	responseWriterType reflect.Type
	requestType        reflect.Type
	*Injector
}

func NewHttpInjector() Http {
	return &httpImpl{Injector: newInjector()}
}

func (r *httpImpl) registerContextProvider() {
	var responseWriter http.ResponseWriter
	var request http.Request

	r.responseWriterType = reflect.TypeOf(responseWriter)
	r.requestType = reflect.TypeOf(request)

	r.providers[r.responseWriterType] = nil
	r.providers[r.requestType] = nil
}

func (r *httpImpl) HandleFunc(pattern string, handler Handler) (Http, error) {
	httpHandler, err := r.registerHandlerFunctions(handler)

	http.HandleFunc(pattern, httpHandler)

	return r, err
}

func (r *httpImpl) RegisterController(controller Controller) (err error) {
	defer func() {
		e := recover()

		if injectErr, ok := e.(Error); ok {
			err = injectErr
			return
		}

		if e != nil {
			panic(e)
		}
	}()

	ctrlVal := reflect.ValueOf(controller)
	ctrlType := ctrlVal.Type()

	ctrlFieldProviders := r.controllerProviders(ctrlVal)

	for route, handlerMethodName := range controller.Routes() {
		if method := ctrlVal.MethodByName(handlerMethodName); !method.IsValid() {
			panic(newUnknownHttpHandlerMethodName(ctrlType, handlerMethodName))
		}

		http.HandleFunc(route, r.controllerHandler(ctrlType, handlerMethodName, ctrlFieldProviders))
	}

	return
}

func (r *httpImpl) controllerHandler(ctrlType reflect.Type, handlerMethodName string, ctrlFieldProviders []*typedProvider) HttpHandlerFunc {
	isPtrType := ctrlType.Kind() == reflect.Ptr

	if isPtrType {
		ctrlType = ctrlType.Elem()
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		resolvedCtrl := resolveController(ctrlType, ctrlFieldProviders, r.resolvedCtxValues(writer, request))

		if isPtrType {
			resolvedCtrl.MethodByName(handlerMethodName).Call([]reflect.Value{})
		} else {
			resolvedCtrl.Elem().MethodByName(handlerMethodName).Call([]reflect.Value{})
		}
	}
}

func (r *httpImpl) registerHandlerFunctions(handler Handler) (httpHandler HttpHandlerFunc, err error) {
	defer func() {
		e := recover()

		if injectErr, ok := e.(Error); ok {
			err = injectErr
			return
		}

		if e != nil {
			panic(e)
		}
	}()

	handlerFuncValue := funcValueOf(handler)

	handlerFuncProviders := r.registeredProviders(handler)

	httpHandler = func(responseWriter http.ResponseWriter, request *http.Request) {
		handlerFuncValue.Call(resolveProviders(
			handlerFuncProviders,
			r.resolvedCtxValues(responseWriter, request),
		))
	}

	return
}

func (r *httpImpl) resolvedCtxValues(responseWriter http.ResponseWriter, request *http.Request) resolvedValues {
	return map[reflect.Type]reflect.Value{
		r.responseWriterType: reflect.ValueOf(responseWriter),
		r.requestType:        reflect.ValueOf(request),
	}
}
