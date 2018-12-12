package injection

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
)

type IRoutes interface {
	Use(...Handler) (IRoutes, error)

	Handle(string, string, ...Handler) (IRoutes, error)
	Any(string, ...Handler) (IRoutes, error)
	GET(string, ...Handler) (IRoutes, error)
	POST(string, ...Handler) (IRoutes, error)
	DELETE(string, ...Handler) (IRoutes, error)
	PATCH(string, ...Handler) (IRoutes, error)
	PUT(string, ...Handler) (IRoutes, error)
	OPTIONS(string, ...Handler) (IRoutes, error)
	HEAD(string, ...Handler) (IRoutes, error)

	StaticFile(string, string) IRoutes
	Static(string, string) IRoutes
	StaticFS(string, http.FileSystem) IRoutes

	RegisterProvider(Provider) bool
	RegisterProviders(...Provider) error

	RegisterController(Controller) (IRoutes, error)
}

type RouterGroup struct {
	*Injector
	contextType reflect.Type
	ginRoutes   gin.IRoutes
}

func NewRouter(ginRoutes gin.IRoutes) *RouterGroup {
	routesImpl := &RouterGroup{ginRoutes: ginRoutes, Injector: newInjector()}
	routesImpl.registerContextProvider()

	return routesImpl
}

// Lahendus tage ja string nimetusega contekstist otse küsida, koos skoobi toega vajab providerStruct lahendust

func (r *RouterGroup) registerContextProvider() {
	var context *gin.Context
	contextType := reflect.TypeOf(context)

	r.contextType = contextType
	r.providers[contextType] = nil
}

func (r *RouterGroup) Use(handlers ...Handler) (IRoutes, error) {
	ginHandlers, err := r.registerGinHandlerFunctions(handlers)

	if err == nil {
		r.ginRoutes = r.ginRoutes.Use(ginHandlers...)
	}

	return r, err
}

func (r *RouterGroup) Handle(requestType string, endPoint string, handlers ...Handler) (IRoutes, error) {
	ginHandlers, err := r.registerGinHandlerFunctions(handlers)

	if err == nil {
		r.ginRoutes = r.ginRoutes.Handle(requestType, endPoint, ginHandlers...)
	}

	return r, err
}

func (r *RouterGroup) Any(endPoint string, handlers ...Handler) (IRoutes, error) {
	ginHandlers, err := r.registerGinHandlerFunctions(handlers)

	if err == nil {
		r.ginRoutes = r.ginRoutes.Any(endPoint, ginHandlers...)
	}

	return r, err
}

func (r *RouterGroup) GET(endPoint string, handlers ...Handler) (IRoutes, error) {
	ginHandlers, err := r.registerGinHandlerFunctions(handlers)

	if err == nil {
		r.ginRoutes = r.ginRoutes.GET(endPoint, ginHandlers...)
	}

	return r, err
}

func (r *RouterGroup) POST(endPoint string, handlers ...Handler) (IRoutes, error) {
	ginHandlers, err := r.registerGinHandlerFunctions(handlers)

	if err == nil {
		r.ginRoutes = r.ginRoutes.POST(endPoint, ginHandlers...)
	}

	return r, err
}

func (r *RouterGroup) DELETE(endPoint string, handlers ...Handler) (IRoutes, error) {
	ginHandlers, err := r.registerGinHandlerFunctions(handlers)

	if err == nil {
		r.ginRoutes = r.ginRoutes.DELETE(endPoint, ginHandlers...)
	}

	return r, err
}

func (r *RouterGroup) PATCH(endPoint string, handlers ...Handler) (IRoutes, error) {
	ginHandlers, err := r.registerGinHandlerFunctions(handlers)

	if err == nil {
		r.ginRoutes = r.ginRoutes.PATCH(endPoint, ginHandlers...)
	}

	return r, err
}

func (r *RouterGroup) PUT(endPoint string, handlers ...Handler) (IRoutes, error) {
	ginHandlers, err := r.registerGinHandlerFunctions(handlers)

	if err == nil {
		r.ginRoutes = r.ginRoutes.PUT(endPoint, ginHandlers...)
	}

	return r, err
}

func (r *RouterGroup) OPTIONS(endPoint string, handlers ...Handler) (IRoutes, error) {
	ginHandlers, err := r.registerGinHandlerFunctions(handlers)

	if err == nil {
		r.ginRoutes = r.ginRoutes.OPTIONS(endPoint, ginHandlers...)
	}

	return r, err
}

func (r *RouterGroup) HEAD(endPoint string, handlers ...Handler) (IRoutes, error) {
	ginHandlers, err := r.registerGinHandlerFunctions(handlers)

	if err == nil {
		r.ginRoutes = r.ginRoutes.HEAD(endPoint, ginHandlers...)
	}

	return r, err
}

func (r *RouterGroup) StaticFile(relativePath, filepath string) IRoutes {
	r.ginRoutes = r.ginRoutes.StaticFile(relativePath, filepath)

	return r
}

func (r *RouterGroup) Static(relativePath, root string) IRoutes {
	r.ginRoutes = r.ginRoutes.Static(relativePath, root)

	return r
}

func (r *RouterGroup) StaticFS(relativePath string, fs http.FileSystem) IRoutes {
	r.ginRoutes = r.ginRoutes.StaticFS(relativePath, fs)

	return r
}

func (r *RouterGroup) RegisterController(controller Controller) (routes IRoutes, err error) {
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

		httpMethod := handlerHttpMethod(handlerMethodName)

		r.ginRoutes = r.ginRoutes.Handle(
			httpMethod,
			route,
			r.controllerHandler(ctrlType, handlerMethodName, ctrlFieldProviders),
		)
	}

	return routes, err
}

func (r *RouterGroup) controllerHandler(ctrlType reflect.Type, handlerMethodName string, ctrlFieldProviders []*typedProvider) gin.HandlerFunc {
	isPtrType := ctrlType.Kind() == reflect.Ptr

	if isPtrType {
		ctrlType = ctrlType.Elem()
	}

	return func(context *gin.Context) {
		resolvedCtrl := resolveController(ctrlType, ctrlFieldProviders, r.resolvedCtxValues(context))

		if isPtrType {
			resolvedCtrl.MethodByName(handlerMethodName).Call([]reflect.Value{})
		} else {
			resolvedCtrl.Elem().MethodByName(handlerMethodName).Call([]reflect.Value{})
		}
	}
}

func (r *RouterGroup) registerGinHandlerFunctions(handlers []Handler) (ginHandlers []gin.HandlerFunc, err error) {
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

	for _, handlerFunc := range handlers {
		ginHandlers = append(ginHandlers, r.routeHandler(handlerFunc))
	}

	return ginHandlers, err
}

func (r *RouterGroup) routeHandler(handlerFunc Handler) gin.HandlerFunc {
	handlerFuncValue := funcValueOf(handlerFunc)
	handlerFuncProviders := r.registeredProviders(handlerFunc)

	return func(context *gin.Context) {
		providersValues := resolveProviders(
			handlerFuncProviders,
			r.resolvedCtxValues(context),
		)

		handlerFuncValue.Call(providersValues)
	}
}

func (r *RouterGroup) resolvedCtxValues(ginContext *gin.Context) resolvedValues {
	return map[reflect.Type]reflect.Value{r.contextType: reflect.ValueOf(ginContext)}
}
