package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/surmus/injection"
	"reflect"
)

type adapter struct {
	ginRoutesValue reflect.Value
	handlerFnType  reflect.Type
	contextType    reflect.Type
}

func newAdapter(ginRoutes gin.IRoutes) *adapter {
	return &adapter{
		ginRoutesValue: reflect.ValueOf(ginRoutes),
		handlerFnType:  reflect.TypeOf(new(gin.HandlerFunc)).Elem(),
	}
}

func Adapt(ginRoutes gin.IRoutes) *injection.Injector {
	injector, err := injection.NewInjector(newAdapter(ginRoutes))

	if err != nil {
		panic(err)
	}

	return injector
}

func (r *adapter) Use(handlerFnValues ...reflect.Value) injection.Routes {
	fnCallResult := r.ginRoutesValue.
		MethodByName("Use").
		Call(handlerFnValues)[0]

	return newAdapter(fnCallResult.Interface().(gin.IRoutes))
}

func (r *adapter) Handle(httpMethod string, endPoint string, handlerFnValues ...reflect.Value) injection.Routes {
	handleFnInputValues := []reflect.Value{reflect.ValueOf(httpMethod), reflect.ValueOf(endPoint)}

	for _, value := range handlerFnValues {
		handleFnInputValues = append(handleFnInputValues, value)
	}

	fnCallResult := r.ginRoutesValue.
		MethodByName("Handle").
		Call(handleFnInputValues)[0]

	return newAdapter(fnCallResult.Interface().(gin.IRoutes))
}

func (r *adapter) HandlerFnType() reflect.Type {
	return r.handlerFnType
}
