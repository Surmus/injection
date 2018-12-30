package injection

import (
	"context"
	"github.com/stretchr/testify/assert"
	"injection/test"
	"net/http"
	"reflect"
	"testing"
)

const testHttpMethod = http.MethodGet
const testCtxKey = "TEST"
const testCtxVal = "TEST"
const testEndpoint = "/test"

func testHandlerFn(ctx context.Context) {}

type testRoutes struct {
	t *testing.T
}

func (r *testRoutes) Use(handlerFnValues ...reflect.Value) Routes {
	ctx := context.WithValue(context.Background(), testCtxKey, testCtxVal)
	ctxValue := reflect.ValueOf(ctx)

	for _, handlerValue := range handlerFnValues {
		handlerValue.Call([]reflect.Value{ctxValue})
	}

	return r
}

func (r *testRoutes) Handle(httpMethod string, endPoint string, handlerFnValues ...reflect.Value) Routes {
	assert.Equal(r.t, testEndpoint, endPoint)
	assert.Equal(r.t, testHttpMethod, httpMethod)

	ctx := context.WithValue(context.Background(), testCtxKey, testCtxVal)
	ctxValue := reflect.ValueOf(ctx)

	for _, handlerValue := range handlerFnValues {
		handlerValue.Call([]reflect.Value{ctxValue})
	}

	return r
}

func (*testRoutes) HandlerFnType() reflect.Type {
	return reflect.TypeOf(testHandlerFn)
}

type invalidHandlerTypeRoutes struct {
	testRoutes
}

func (*invalidHandlerTypeRoutes) HandlerFnType() reflect.Type {
	return reflect.TypeOf("string")
}

type invalidSignatureHandlerRoutes struct {
	testRoutes
}

func (*invalidSignatureHandlerRoutes) HandlerFnType() reflect.Type {
	return reflect.TypeOf(func() {})
}

type nonContextHandlerRoutes struct {
	testRoutes
}

func (*nonContextHandlerRoutes) HandlerFnType() reflect.Type {
	return reflect.TypeOf(func(interface{}) {})
}

type TestController struct {
	t *testing.T
}

func (c *TestController) Routes() map[string]string {
	return map[string]string{testEndpoint: "GetTest"}
}

func (c *TestController) GetTest(context context.Context) {
	assert.Equal(c.t, testCtxVal, context.Value(testCtxKey).(string))
}

type InvalidMethodParamCountCtrl struct {
	TestController
}

func (c *InvalidMethodParamCountCtrl) GetTest() {}

type InvalidMethodCtxValueTypeCtrl struct {
	TestController
}

func (c *InvalidMethodCtxValueTypeCtrl) GetTest(invalidParam string) {}

func setupInjector(t *testing.T) *Injector {
	injector, _ := NewInjector(&testRoutes{t: t})

	return injector
}

func TestInjector_Handle(t *testing.T) {
	var ctxVal string

	injector := setupInjector(t)

	handleRegisterErr := injector.Handle(http.MethodGet, testEndpoint, func(ctx context.Context) {
		ctxVal = ctx.Value(testCtxKey).(string)
	})

	assert.Nil(t, handleRegisterErr)
	assert.Equal(t, testCtxVal, ctxVal)
}

func TestInjector_Use(t *testing.T) {
	var ctxVal string

	injector := setupInjector(t)

	handleRegisterErr := injector.Use(func(ctx context.Context) {
		ctxVal = ctx.Value(testCtxKey).(string)
	})

	assert.Nil(t, handleRegisterErr)
	assert.Equal(t, testCtxVal, ctxVal)
}

func TestInjector_RegisterController(t *testing.T) {
	tests := map[string]func(t *testing.T){
		"should successfully register controller with injector": func(t *testing.T) {
			injector := setupInjector(t)

			registrationError := injector.RegisterController(&TestController{t: t})

			assert.Nil(t, registrationError)
		},
		"should fail to register controller with invalid controller route handler signature param count": func(t *testing.T) {
			injector := setupInjector(t)

			registrationError := injector.RegisterController(&InvalidMethodParamCountCtrl{TestController{t: t}})

			assert.NotNil(t, registrationError)
			assert.IsType(t, Error{}, registrationError)
		},
		"should fail to register controller with invalid controller route handler context parameter": func(t *testing.T) {
			injector := setupInjector(t)

			registrationError := injector.RegisterController(&InvalidMethodCtxValueTypeCtrl{TestController{t: t}})

			assert.NotNil(t, registrationError)
			assert.IsType(t, Error{}, registrationError)
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, testCase)
	}
}

func TestNewInjector(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"should register new injector": func(t *testing.T) {
			injector, setupErr := NewInjector(&testRoutes{t: t})

			assert.Nil(t, setupErr)
			assert.NotNil(t, injector)
		},
		"should fail to register injector with non function routes request handler": func(t *testing.T) {
			_, setupErr := NewInjector(&invalidHandlerTypeRoutes{testRoutes{t: t}})

			assert.NotNil(t, setupErr)
		},
		"should fail to register injector with invalid signature routes request handler function": func(t *testing.T) {
			_, setupErr := NewInjector(&invalidSignatureHandlerRoutes{testRoutes{t: t}})

			assert.NotNil(t, setupErr)
		},
		"should fail to register injector with request handler when context param not implementing context": func(t *testing.T) {
			_, setupErr := NewInjector(&nonContextHandlerRoutes{testRoutes{t: t}})

			assert.NotNil(t, setupErr)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}

func TestInjector_RegisterProviders(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"should register providers with injector": func(t *testing.T) {
			valueRequiringContextProvider := func(ctx context.Context) *test.DependencyStruct {
				return &test.DependencyStruct{Ctx: ctx}
			}

			constantProvider := func() string {
				return test.Constant
			}

			valueWithInnerDependencyProvider := func(dependency *test.DependencyStruct) test.DependencyInterface {
				return dependency
			}

			injector, _ := NewInjector(&testRoutes{t: t})

			err := injector.RegisterProviders(
				valueWithInnerDependencyProvider,
				valueRequiringContextProvider,
				constantProvider,
			)

			assert.Nil(t, err)
		},
		"should fail to register invalid type provider": func(t *testing.T) {
			injector, _ := NewInjector(&testRoutes{t: t})

			err := injector.RegisterProviders("INVALID-PROVIDER")

			assert.NotNil(t, err)
			assert.IsType(t, Error{}, err)
		},
		"should fail to register provider with invalid return values count": func(t *testing.T) {
			injector, _ := NewInjector(&testRoutes{t: t})

			err := injector.RegisterProviders(func() {})

			assert.NotNil(t, err)
			assert.IsType(t, Error{}, err)
		},
		"should fail to register provider with unregistered dependency": func(t *testing.T) {
			injector, _ := NewInjector(&testRoutes{t: t})

			err := injector.RegisterProviders(func(dependency *test.DependencyStruct) {})

			assert.NotNil(t, err)
			assert.IsType(t, Error{}, err)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}
