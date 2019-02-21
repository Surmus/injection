package injection

import (
	"reflect"
)

// Routes is used as integration layer between http library and Injector for usage see gin package
type Routes interface {
	Use(handlerFnValues ...reflect.Value) Routes

	Handle(httpMethod string, endPoint string, handlerFnValues ...reflect.Value) Routes

	HandlerFnType() reflect.Type
}

// Injector acts as DI container, resolver and register for underlying Routes implementation
type Injector struct {
	routes      Routes
	contextType reflect.Type
	providers   map[reflect.Type]registeredProvider
}

// NewInjector crates new Injector instance,
// returns error when given Routes implementation http request handler function(HandlerFnType):
// - has more than one input parameter or the single parameter does not implement context.Context interface
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

// From creates new Injector from existing by copying over all registered value providers from given Injector
func From(from *Injector, routes Routes) (*Injector, error) {
	if err := validateRoutes(routes); err != nil {
		return nil, err
	}

	injector := &Injector{
		routes:    routes,
		providers: map[reflect.Type]registeredProvider{},
	}
	injector.registerContextProvider()

	for providerType, provider := range from.providers {
		injector.providers[providerType] = provider
	}

	return injector, nil
}

func (r *Injector) registerContextProvider() {
	r.contextType = r.routes.HandlerFnType().In(0)
	r.providers[r.contextType] = nil
}

func (r *Injector) registeredProviders(handler Provider) (handlerFuncProviders []*typedProvider) {
	var handlerType reflect.Type
	var firstFnParamIndex int

	if handlerMethod, ok := handler.(reflect.Method); ok {
		handlerType = handlerMethod.Func.Type()
		firstFnParamIndex = 1 // first param for method type is receiver, ignore it
	} else {
		handlerType = reflect.TypeOf(handler)
	}

	for ; firstFnParamIndex < handlerType.NumIn(); firstFnParamIndex++ {
		dependencyType := handlerType.In(firstFnParamIndex)

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
	var providerValue reflect.Value
	var providerType reflect.Type

	if singletonProvider, ok := provider.(*singletonProvider); ok {
		providerValue = funcValueOf(singletonProvider.provider)
		providerType = providerValue.Type()
	} else {
		providerValue = funcValueOf(provider)
		providerType = providerValue.Type()
	}

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

	providedValueType := providerType.Out(0)
	r.providers[providedValueType] = func(resolvedValues resolvedValues) interface{} {
		callResults := providerValue.Call(resolveProviders(dependencyProviders, resolvedValues))
		resolvedValue := callResults[0].Interface()

		if _, ok := provider.(*singletonProvider); ok {
			r.providers[providedValueType] = registeredSingletonProvider(resolvedValue)
		}

		return resolvedValue
	}

	return true
}

// RegisterProviders registers value provider functions into DI container.
// Providable values are saved as type/value map, one type can only have one value, providing another will overwrite old value
// returns error when:
// - provider is not a function
// - value provider function returns more than one value
// - value provider function call signature contains type which is registered or is not present as provider in providers slice
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

// RegisterController enables given Controller implementation to have field values and http request handler function input values
// injected from registered value providers,
// returns error when given Controller Routes method result contains unknown Controller method
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

	for _, controllerRoute := range routesList(controller) {
		if validationErr := validateControllerMethod(controllerRoute.methodName, ctrlVal); validationErr != nil {
			panic(validationErr)
		}

		httpMethod := handlerHTTPMethod(controllerRoute.methodName)

		routeHandlers := r.routeMiddlewareHandlers(controller.Middleware()[controllerRoute.methodName])
		routeHandlers = append(routeHandlers, r.controllerHandler(ctrlType, controllerRoute.methodName, ctrlFieldProviders))

		r.routes = r.routes.Handle(
			httpMethod,
			controllerRoute.route,
			routeHandlers...,
		)
	}

	return err
}

func (r *Injector) routeMiddlewareHandlers(handlers []Handler) []reflect.Value {
	registeredHandlers, err := r.registerHandlerFunctions(handlers)

	if err != nil {
		panic(err)
	}

	return registeredHandlers
}

func (r *Injector) controllerHandler(ctrlType reflect.Type, handlerMethodName string, ctrlFieldProviders []*typedProvider) reflect.Value {
	handlerMethodType, _ := ctrlType.MethodByName(handlerMethodName)
	handlerMethodProviders := r.registeredProviders(handlerMethodType)

	return reflect.MakeFunc(r.routes.HandlerFnType(), func(args []reflect.Value) (results []reflect.Value) {
		resolvedValues := r.resolvedCtxValues(args[0])

		// resolveProviders fn adds any resolved value into resolvedValues variable
		methodProvidersValues := resolveProviders(
			handlerMethodProviders,
			resolvedValues,
		)

		if isPtrType(ctrlType) {
			resolvedCtrl := resolveController(ctrlType.Elem(), ctrlFieldProviders, resolvedValues)
			resolvedCtrl.MethodByName(handlerMethodName).Call(methodProvidersValues)
		} else {
			resolvedCtrl := resolveController(ctrlType, ctrlFieldProviders, resolvedValues)
			resolvedCtrl.Elem().MethodByName(handlerMethodName).Call(methodProvidersValues)
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

// Use registers http middleware handlers, returns error when handler function signature contains unregistered values
func (r *Injector) Use(handlers ...Handler) error {
	registeredHandlers, err := r.registerHandlerFunctions(handlers)

	if err == nil {
		r.routes = r.routes.Use(registeredHandlers...)
	}

	return err
}

// Handle registers a new request handle and middleware with the given path and method.
// The last handler should be the real handler, the other ones should be middleware that can and should be shared among different routes.
// Returns error when handler function signature contains unregistered values
func (r *Injector) Handle(httpMethod string, endPoint string, handlers ...Handler) error {
	registeredHandlers, err := r.registerHandlerFunctions(handlers)

	if err == nil {
		r.routes = r.routes.Handle(httpMethod, endPoint, registeredHandlers...)
	}

	return err
}
