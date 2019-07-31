package injection

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/surmus/injection/test"
	"net/http"
	"reflect"
	"testing"
)

var middlewareFnExecuted bool

type testRoutes struct {
	t *testing.T
}

func (r *testRoutes) Use(handlerFnValues ...reflect.Value) Routes {
	ctx := context.WithValue(context.Background(), test.CtxKey, test.CtxVal)
	ctxValue := reflect.ValueOf(ctx)

	for _, handlerValue := range handlerFnValues {
		handlerValue.Call([]reflect.Value{ctxValue})
	}

	return r
}

func (r *testRoutes) Handle(httpMethod string, endPoint string, handlerFnValues ...reflect.Value) Routes {
	assert.Equal(r.t, test.Endpoint, endPoint)
	assert.Equal(r.t, test.HttpMethod, httpMethod)

	ctx := context.WithValue(context.Background(), test.CtxKey, test.CtxVal)
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

type PointerController struct {
	BaseController

	primitiveZeroValue int

	PrimitiveValConstant string

	Context context.Context

	valueWithInnerDependency test.DependencyInterface

	valueRequiringContext *test.DependencyStruct

	t *testing.T
}

func NewPointerController(t *testing.T) *PointerController {
	return &PointerController{
		PrimitiveValConstant: test.Constant,
		t:                    t,
	}
}

func (c *PointerController) Routes() map[string][]string {
	return map[string][]string{test.Endpoint: {"GetTest"}}
}

func (c *PointerController) GetTest(context context.Context) {
	assert.NotNil(c.t, c.Context)
	assert.NotNil(c.t, c.valueRequiringContext)
	assert.Equal(c.t, test.Constant, c.PrimitiveValConstant)
	assert.NotNil(c.t, c.valueWithInnerDependency)
	assert.Equal(c.t, 0, c.primitiveZeroValue)

	assert.Equal(c.t, c.Context, c.valueRequiringContext.Ctx)
	assert.Equal(c.t, c.Context, context)

	assert.True(
		c.t,
		c.valueWithInnerDependency.(*test.DependencyStruct) == c.valueRequiringContext,
		"valueWithInnerDependency interface should be created from valueRequiringContext",
	)
}

type ValueController struct {
	primitiveValConstant string

	ValueRequiringContext *test.DependencyStruct

	T *testing.T
}

func NewValueController(t *testing.T) ValueController {
	return ValueController{
		primitiveValConstant: test.Constant,
		T:                    t,
	}
}

func (c ValueController) Routes() map[string][]string {
	return map[string][]string{test.Endpoint: {"HandleRequest"}}
}

func (c ValueController) Middleware() map[string][]Handler {
	middlewareFnExecuted = false

	return map[string][]Handler{"HandleRequest": {
		func(valueRequiringContext *test.DependencyStruct) {
			middlewareFnExecuted = true
		},
	}}
}

func (c ValueController) HandleRequest(context context.Context, valueWithInnerDependency test.DependencyInterface) {
	assert.NotNil(c.T, c.ValueRequiringContext)
	assert.Equal(c.T, test.Constant, c.primitiveValConstant)
	assert.NotNil(c.T, valueWithInnerDependency)

	assert.True(
		c.T,
		valueWithInnerDependency.(*test.DependencyStruct) == c.ValueRequiringContext,
		"valueWithInnerDependency interface should be created from valueRequiringContext",
	)
}

type InvalidRoutesMapController struct {
	BaseController
}

func (c *InvalidRoutesMapController) Routes() map[string][]string {
	return map[string][]string{test.Endpoint: {"PostTest"}}
}

type InvalidMiddlewareController struct {
	ValueController
}

func (c InvalidMiddlewareController) Middleware() map[string][]Handler {
	return map[string][]Handler{"HandleRequest": {"INVALID-VALUE"}}
}

func testHandlerFn(ctx context.Context) {}

func setupTestHandlerFn(t *testing.T) interface{} {
	return func(
		ctx context.Context,
		valueWithInnerDependency test.DependencyInterface,
		valueRequiringContext *test.DependencyStruct,
		providedConstant string,
	) {
		assert.NotNil(t, ctx)
		assert.NotNil(t, valueRequiringContext)
		assert.Equal(t, test.Constant, providedConstant)
		assert.NotNil(t, valueWithInnerDependency)

		assert.Equal(t, ctx, valueRequiringContext.Ctx)

		assert.True(
			t,
			valueWithInnerDependency.(*test.DependencyStruct) == valueRequiringContext,
			"valueWithInnerDependency interface should be created from valueRequiringContext",
		)

		assert.Equal(t, test.CtxVal, ctx.Value(test.CtxKey))
	}
}

func setupInjector(t *testing.T) *Injector {
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
	injector.RegisterProviders(valueWithInnerDependencyProvider, valueRequiringContextProvider, constantProvider)

	return injector
}

func TestInjector_Handle(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"successfully register handler and handle request": func(t *testing.T) {
			injector := setupInjector(t)

			handleRegisterErr := injector.Handle(http.MethodGet, test.Endpoint, setupTestHandlerFn(t))

			assert.Nil(t, handleRegisterErr)
		},
		"resolve singletonProvider only once": func(t *testing.T) {
			var resolvedValues []*test.DependencyStruct

			provider := func() *test.DependencyStruct { return &test.DependencyStruct{} }
			injector, _ := NewInjector(&testRoutes{t: t})
			err := injector.RegisterProviders(NewSingletonProvider(provider))

			for i := 0; i < 2; i++ {
				injector.Handle(http.MethodGet, test.Endpoint, func(singletonVal *test.DependencyStruct) {
					resolvedValues = append(resolvedValues, singletonVal)
				})
			}

			assert.Nil(t, err)
			assert.Len(t, resolvedValues, 2)
			assert.True(t, resolvedValues[0] == resolvedValues[1], "both values in resolvedValues should be same instance")
		},
		"resolve singletonProvider only once when singleton is dependency to non singleton provider": func(t *testing.T) {
			var providerExecutedTimes int

			provider := func(ctx context.Context) *test.DependencyStruct {
				providerExecutedTimes++
				return &test.DependencyStruct{}
			}
			provider2 := func(dep *test.DependencyStruct) test.DependencyInterface {
				return dep
			}

			injector, _ := NewInjector(&testRoutes{t: t})
			err := injector.RegisterProviders(NewSingletonProvider(provider), provider2)

			for i := 0; i < 2; i++ {
				injector.Handle(http.MethodGet, test.Endpoint, func(s test.DependencyInterface) {})
			}

			assert.Nil(t, err)
			assert.Equal(t, 1, providerExecutedTimes)
		},
		"resolve new instance with non singletonProvider": func(t *testing.T) {
			var resolvedValues []*test.DependencyStruct

			provider := func() *test.DependencyStruct { return &test.DependencyStruct{} }
			injector, _ := NewInjector(&testRoutes{t: t})
			err := injector.RegisterProviders(provider)

			for i := 0; i < 2; i++ {
				injector.Handle(http.MethodGet, test.Endpoint, func(singletonVal *test.DependencyStruct) {
					resolvedValues = append(resolvedValues, singletonVal)
				})
			}

			assert.Nil(t, err)
			assert.Len(t, resolvedValues, 2)
			assert.True(t, resolvedValues[0] != resolvedValues[1], "both values should be different instances")
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			injector, _ := NewInjector(&testRoutes{t: t})
			testHandlerFn := setupTestHandlerFn(t)

			registrationError := injector.Handle(http.MethodGet, test.Endpoint, testHandlerFn)

			assert.IsType(t, Error{}, registrationError)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}

func TestInjector_Use(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"successfully register handler and handle request": func(t *testing.T) {
			injector := setupInjector(t)

			handleRegisterErr := injector.Use(setupTestHandlerFn(t))

			assert.Nil(t, handleRegisterErr)
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			injector, _ := NewInjector(&testRoutes{t: t})
			testHandlerFn := setupTestHandlerFn(t)

			registrationError := injector.Use(testHandlerFn)

			assert.IsType(t, Error{}, registrationError)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}

func TestInjector_RegisterController(t *testing.T) {
	tests := map[string]func(t *testing.T){
		"successfully for Controller pointer receiver request handler method": func(t *testing.T) {
			r := setupInjector(t)

			registrationError := r.RegisterController(NewPointerController(t))

			assert.Nil(t, registrationError)
		},
		"successfully for Controller value receiver request handler method": func(t *testing.T) {
			r := setupInjector(t)

			registrationError := r.RegisterController(NewValueController(t))

			assert.Nil(t, registrationError)
		},
		"should execute controller method middleware": func(t *testing.T) {
			r := setupInjector(t)

			registrationError := r.RegisterController(NewValueController(t))

			assert.Nil(t, registrationError)
			assert.True(t, middlewareFnExecuted)
		},
		"fail registering Controller with invalid method middleware": func(t *testing.T) {
			r := setupInjector(t)

			registrationError := r.RegisterController(InvalidMiddlewareController{ValueController{T: t}})

			assert.IsType(t, Error{}, registrationError)
		},
		"fail registering Controller with unregistered dependencies": func(t *testing.T) {
			injector, _ := NewInjector(&testRoutes{t: t})

			registrationError := injector.RegisterController(NewPointerController(t))

			assert.IsType(t, Error{}, registrationError)
		},
		"fail registering Controller with incorrect routes mapping": func(t *testing.T) {
			injector, _ := NewInjector(&testRoutes{t: t})

			registrationError := injector.RegisterController(new(InvalidRoutesMapController))

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

			err := injector.RegisterProviders(func(dependency *test.DependencyStruct) test.DependencyInterface {
				return dependency
			})

			assert.NotNil(t, err)
			assert.IsType(t, Error{}, err)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}

func TestFrom(t *testing.T) {
	dependencyValueProvider := func(ctx context.Context) *test.DependencyStruct {
		return &test.DependencyStruct{Ctx: ctx}
	}
	valueWithDependencyProvider := func(dependency *test.DependencyStruct) test.DependencyInterface {
		return dependency
	}

	injectorFrom, _ := NewInjector(&testRoutes{t: t})
	injectorFrom.RegisterProviders(dependencyValueProvider)

	injectorCpy, cpyCreationErr := From(injectorFrom, &testRoutes{t: t})
	regError := injectorCpy.RegisterProviders(valueWithDependencyProvider)

	assert.Nil(t, cpyCreationErr)
	assert.Nil(t, regError)

	// New injector does have previously registered test.DependencyInterface
	assert.Nil(t, injectorCpy.RegisterProviders(func(test.DependencyInterface) int { return 1 }))

	// Old injector does not have previously registered test.DependencyInterface
	assert.NotNil(t, injectorFrom.RegisterProviders(func(test.DependencyInterface) int { return 1 }))
}
