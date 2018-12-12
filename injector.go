package injection

import (
	"reflect"
)

type Injector struct {
	providers map[reflect.Type]registeredProvider
}

func newInjector() *Injector {
	injector := &Injector{providers: map[reflect.Type]registeredProvider{}}

	return injector
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

func (r *Injector) RegisterProvider(provider Provider) bool {
	var dependencyProviders []*typedProvider

	providerValue := funcValueOf(provider)
	providerType := providerValue.Type()

	if providerType.NumOut() > 1 {
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
		providerRegistered := r.RegisterProvider(provider)

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
