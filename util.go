package injection

import (
	"context"
	"fmt"
	"github.com/fatih/camelcase"
	"net/http"
	"reflect"
	"strings"
	"unsafe"
)

func resolveProviders(providers []*typedProvider, resolvedValues resolvedValues) []reflect.Value {
	var provided []reflect.Value

	for _, provider := range providers {
		if resolvedValue, alreadyResolved := resolvedValues[provider.kind]; alreadyResolved {
			provided = append(provided, resolvedValue)
			continue
		}

		providedVal := reflect.ValueOf(provider.value(resolvedValues))
		provided = append(provided, providedVal)

		resolvedValues[provider.kind] = providedVal
	}

	return provided
}

func resolveController(ctrlType reflect.Type, ctrlFieldProviders []*typedProvider, resolvedValues resolvedValues) *reflect.Value {
	ctrlPtrVal := reflect.New(ctrlType)
	ctrlVal := ctrlPtrVal.Elem()

	for i := 0; i < ctrlType.NumField(); i++ {
		provider := ctrlFieldProviders[i]
		var providedVal reflect.Value

		if resolvedValue, alreadyResolved := resolvedValues[provider.kind]; alreadyResolved {
			providedVal = resolvedValue
		} else {
			providedVal = reflect.ValueOf(provider.value(resolvedValues))

			resolvedValues[provider.kind] = providedVal
		}

		unsafeFieldElem(ctrlVal, i).Set(providedVal)
	}

	return &ctrlPtrVal
}

func providerString(p Provider) string {
	providerType := reflect.TypeOf(p)
	inputParamTypes := make([]string, 0)

	for i := 0; i < providerType.NumIn(); i++ {
		inputParamTypes = append(inputParamTypes, providerType.In(i).String())
	}

	return fmt.Sprintf(
		"(%s) -> %s",
		strings.Join(inputParamTypes, ", "),
		providerType.Out(0),
	)
}

func handlerHttpMethod(handlerMethodName string) string {
	nameParts := camelcase.Split(handlerMethodName)

	for _, httpMethod := range httpMethods {
		if strings.ToUpper(nameParts[0]) == httpMethod {
			return httpMethod
		}
	}

	return http.MethodGet
}

func funcValueOf(fn interface{}) reflect.Value {
	val := reflect.ValueOf(fn)

	if val.Kind() != reflect.Func {
		panic(newInvalidHandlerError(fn))
	}

	return val
}

func unsafeFieldElem(structVal reflect.Value, fieldIndex int) reflect.Value {
	field := structVal.Field(fieldIndex)

	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
}

func addressableCpy(structVal reflect.Value) reflect.Value {
	ctrlValCpyPtr := reflect.New(structVal.Type()).Elem()
	ctrlValCpyPtr.Set(structVal)

	return ctrlValCpyPtr
}

func isNilValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return value.IsNil()
	}

	return false
}

func isFnType(valType reflect.Type) bool {
	return valType.Kind() == reflect.Func
}

func validateRoutes(routes Routes) error {

	handlerFnType := routes.HandlerFnType()
	contextType := reflect.TypeOf(new(context.Context)).Elem()

	if !isFnType(routes.HandlerFnType()) {
		return newInvalidHandlerError(routes.HandlerFnType().Kind())
	}

	if handlerFnType.NumIn() != 1 {
		return newInvalidHandlerFnParamCountError()
	}

	routesContextType := routes.HandlerFnType().In(0)

	if !routesContextType.Implements(contextType) {
		return newInvalidContextTypeError(routesContextType)
	}

	return nil
}

func validateControllerMethod(methodName string, ctrlVal reflect.Value, routesCtxType reflect.Type) error {
	method := ctrlVal.MethodByName(methodName)

	if !method.IsValid() {
		panic(newUnknownHttpHandlerMethodName(ctrlVal.Type(), methodName))
	}

	methodType := method.Type()

	if methodType.NumIn() != 1 {
		return newMethodParamCountError(methodName)
	}

	if methodType.In(0) != routesCtxType {
		return newInvalidControllerMethod(methodName, routesCtxType)
	}

	return nil
}
