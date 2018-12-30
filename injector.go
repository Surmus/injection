package injection

import (
	"reflect"
)

type Routes interface {
	Use(handlerFnValues ...reflect.Value) Routes

	Handle(httpMethod string, endPoint string, handlerFnValues ...reflect.Value) Routes

	HandlerFnType() reflect.Type
}

type Injector struct {
	routes      Routes
	contextType reflect.Type
	providers   map[reflect.Type]registeredProvider
}

func NewInjector(routes Routes) (*Injector, error) {
	if err := validateRoutes(routes); err != nil {
		return nil, err
	}

	injector := &Injector{
		routes:    routes,
		providers: map[reflect.Type]registeredProvider{},
	}
	injector.registerContextProvider()

	return injector, nil
}

func (r *Injector) registerContextProvider() {
	r.contextType = r.routes.HandlerFnType().In(0)
	r.providers[r.contextType] = nil
}

func (r *Injector) registeredProviders(handler Provider) (handlerFuncProviders []*typedProvider) {
	handlerType := reflect.TypeOf(handler)

	for i := 0; i < handlerType.NumIn(); i++ {
		dependencyType := handlerType.In(i)

		handlerFuncProviders = append(
			handlerFuncProviders,
			newTypedProvider(dependencyType, r.registeredProvider(dependencyType)),
		)
	}

	return
}

func (r *Injector) registeredProvider(providerType reflect.Type) registeredProvider {
	if provider, exists := r.providers[providerType]; exists {
		return provider
	}

	panic(newUnknownProviderRequestError(providerType))
}

func (r *Injector) registerProvider(provider Provider) bool {
	var dependencyProviders []*typedProvider

	providerValue := funcValueOf(provider)
	providerType := providerValue.Type()

	if providerType.NumOut() != 1 {
		panic(newProviderInvalidReturnCountError(providerType))
	}

	for i := 0; i < providerType.NumIn(); i++ {
		providerType := providerType.In(i)

		if provider, exists := r.providers[providerType]; exists {
			dependencyProviders = append(dependencyProviders, newTypedProvider(providerType, provider))

			continue
		}

		return false
	}

	r.providers[providerType.Out(0)] = func(resolvedValues resolvedValues) interface{} {
		resolvedValue := providerValue.Call(resolveProviders(dependencyProviders, resolvedValues))

		return resolvedValue[0].Interface()
	}

	return true
}

func (r *Injector) RegisterProviders(providers ...Provider) (err error) {
	var unRegistered []Provider

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

	for _, provider := range providers {
		providerRegistered := r.registerProvider(provider)

		if !providerRegistered {
			unRegistered = append(unRegistered, provider)
		}
	}

	if err != nil {
		return err
	}

	if len(unRegistered) == 0 {
		return nil
	}

	if len(providers) != len(unRegistered) {
		return r.RegisterProviders(unRegistered...)
	}

	return newCannotRegisterProvidersError(unRegistered)
}

func (r *Injector) controllerProviders(ctrlVal reflect.Value) (ctrlFieldProviders []*typedProvider) {

	if ctrlVal.Kind() == reflect.Ptr {
		ctrlVal = ctrlVal.Elem()
	} else {
		ctrlVal = addressableCpy(ctrlVal)
	}

	for i := 0; i < ctrlVal.NumField(); i++ {
		fieldVal := unsafeFieldElem(ctrlVal, i)
		fieldType := fieldVal.Type()

		if isNilValue(fieldVal) {
			ctrlFieldProviders = append(
				ctrlFieldProviders,
				newTypedProvider(fieldType, r.registeredProvider(fieldType)),
			)

			continue
		}

		ctrlFieldProviders = append(
			ctrlFieldProviders,
			newTypedProvider(fieldType, staticValueRegisterProvider(fieldVal)),
		)
	}

	return
}

func (r *Injector) clone() *Injector {
	copiedProviders := map[reflect.Type]registeredProvider{}

	for providerType, provider := range r.providers {
		copiedProviders[providerType] = provider
	}

	return &Injector{providers: copiedProviders}
}

func (r *Injector) RegisterController(controller Controller) (err error) {
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
		if validationErr := validateControllerMethod(handlerMethodName, ctrlVal, r.contextType); validationErr != nil {
			panic(validationErr)
		}

		httpMethod := handlerHttpMethod(handlerMethodName)

		r.routes = r.routes.Handle(
			httpMethod,
			route,
			r.controllerHandler(ctrlType, handlerMethodName, ctrlFieldProviders),
		)
	}

	return err
}

func (r *Injector) controllerHandler(ctrlType reflect.Type, handlerMethodName string, ctrlFieldProviders []*typedProvider) reflect.Value {
	isPtrType := ctrlType.Kind() == reflect.Ptr

	if isPtrType {
		ctrlType = ctrlType.Elem()
	}

	return reflect.MakeFunc(r.routes.HandlerFnType(), func(args []reflect.Value) (results []reflect.Value) {
		resolvedCtrl := resolveController(ctrlType, ctrlFieldProviders, r.resolvedCtxValues(args[0]))
		ctxVal := args[0]

		if isPtrType {
			resolvedCtrl.MethodByName(handlerMethodName).Call([]reflect.Value{ctxVal})
		} else {
			resolvedCtrl.Elem().MethodByName(handlerMethodName).Call([]reflect.Value{ctxVal})
		}

		return
	})
}

func (r *Injector) registerHandlerFunctions(handlers []Handler) (registeredHandlers []reflect.Value, err error) {
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
		registeredHandlers = append(registeredHandlers, r.routeHandler(handlerFunc))
	}

	return registeredHandlers, err
}

func (r *Injector) routeHandler(handlerFunc Handler) reflect.Value {
	handlerFuncValue := funcValueOf(handlerFunc)
	handlerFuncProviders := r.registeredProviders(handlerFunc)

	return reflect.MakeFunc(r.routes.HandlerFnType(), func(args []reflect.Value) (results []reflect.Value) {
		providersValues := resolveProviders(
			handlerFuncProviders,
			r.resolvedCtxValues(args[0]),
		)

		handlerFuncValue.Call(providersValues)

		return
	})
}

func (r *Injector) resolvedCtxValues(ctxVal reflect.Value) resolvedValues {
	return map[reflect.Type]reflect.Value{r.contextType: ctxVal}
}

func (r *Injector) Use(handlers ...Handler) error {
	registeredHandlers, err := r.registerHandlerFunctions(handlers)

	if err == nil {
		r.routes = r.routes.Use(registeredHandlers...)
	}

	return err
}

func (r *Injector) Handle(httpMethod string, endPoint string, handlers ...Handler) error {
	registeredHandlers, err := r.registerHandlerFunctions(handlers)

	if err == nil {
		r.routes = r.routes.Handle(httpMethod, endPoint, registeredHandlers...)
	}

	return err
}
