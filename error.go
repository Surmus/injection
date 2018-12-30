package injection

import (
	"fmt"
	"reflect"
	"strings"
)

type Error struct {
	error string
}

func (e Error) Error() string {
	return e.error
}

func newProviderInvalidReturnCountError(providerType reflect.Type) Error {
	return Error{fmt.Sprintf("cannot register value(%s) with multiple return values", providerType)}
}

func newUnknownProviderRequestError(providerType reflect.Type) Error {
	return Error{fmt.Sprintf("cannot inject value for unregistered type %s", providerType)}
}

func newUnknownHttpHandlerMethodName(ctrlType reflect.Type, missingMethod string) Error {
	return Error{fmt.Sprintf(
		"cannot register unknown request handler method %s for controller %s",
		missingMethod,
		ctrlType,
	)}
}

func newCannotRegisterProvidersError(providers []Provider) Error {
	providerMessages := make([]string, 0)

	for _, provider := range providers {
		providerMessages = append(providerMessages, providerString(provider))
	}

	return Error{fmt.Sprintf(
		"cannot register providers with signatures: \n %s",
		strings.Join(providerMessages, "\n"),
	)}
}

func newInvalidHandlerError(handler interface{}) Error {
	return Error{fmt.Sprintf("value passed as handler is not function: %T", handler)}
}

func newInvalidHandlerFnParamCountError() Error {
	return Error{"function passed as routes request handler should have one context parameter"}
}

func newInvalidContextTypeError(contextType reflect.Type) Error {
	return Error{fmt.Sprintf(
		"routes request handler context parameter(%s) does not implement context.Context interface",
		contextType,
	)}
}

func newMethodParamCountError(methodName string) Error {
	return Error{fmt.Sprintf(
		"controller method %s signature should contain one parameter with context",
		methodName,
	)}
}

func newInvalidControllerMethod(methodName string, routesCtxType reflect.Type) Error {
	return Error{fmt.Sprintf(
		"controller method %s context parameter should be same type as routes request handler context type %s",
		methodName,
		routesCtxType,
	)}
}
