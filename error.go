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

func newInvalidHandlerError(handler interface{}) error {
	return Error{fmt.Sprintf("Value passed as handler is not function: %T", handler)}
}
