package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/surmus/injection"
	"github.com/surmus/injection/test"
	"net/http"
	"testing"
)

const testEndpoint = "/test-path"
const ctrlPostEndpoint = "/test-path-post"
const ctrlGetEndpoint = "/test-path-get"

type PointerController struct {
	primitiveZeroValue int

	PrimitiveValConstant string

	Context *gin.Context

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

func (c *PointerController) Routes() map[string]string {
	return map[string]string{ctrlPostEndpoint: "PostTest"}
}

func (c *PointerController) PostTest(context *gin.Context) {
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

	c.Context.String(http.StatusTeapot, "%s", test.Response)
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

func (c ValueController) Routes() map[string]string {
	return map[string]string{ctrlGetEndpoint: "GetTest"}
}

func (c ValueController) GetTest(context *gin.Context, valueWithInnerDependency test.DependencyInterface) {
	assert.NotNil(c.T, c.ValueRequiringContext)
	assert.Equal(c.T, test.Constant, c.primitiveValConstant)
	assert.NotNil(c.T, valueWithInnerDependency)

	assert.True(
		c.T,
		valueWithInnerDependency.(*test.DependencyStruct) == c.ValueRequiringContext,
		"valueWithInnerDependency interface should be created from valueRequiringContext",
	)

	context.String(http.StatusTeapot, "%s", test.Response)
}

type InvalidRoutesMapController struct{}

func (c *InvalidRoutesMapController) Routes() map[string]string {
	return map[string]string{ctrlPostEndpoint: "PostTest"}
}

func setupRouterWithProviders() *injection.Injector {
	valueRequiringContextProvider := func(ctx *gin.Context) *test.DependencyStruct {
		return &test.DependencyStruct{Ctx: ctx}
	}

	constantProvider := func() string {
		return test.Constant
	}

	valueWithInnerDependencyProvider := func(dependency *test.DependencyStruct) test.DependencyInterface {
		return dependency
	}

	test.Init()
	r := Adapt(test.Router)

	r.RegisterProviders(valueWithInnerDependencyProvider, valueRequiringContextProvider, constantProvider)

	return r
}

func setupTestHandlerFn(t *testing.T) interface{} {
	return func(
		context *gin.Context,
		valueWithInnerDependency test.DependencyInterface,
		valueRequiringContext *test.DependencyStruct,
		providedConstant string,
	) {
		assert.NotNil(t, context)
		assert.NotNil(t, valueRequiringContext)
		assert.Equal(t, test.Constant, providedConstant)
		assert.NotNil(t, valueWithInnerDependency)

		assert.Equal(t, context, valueRequiringContext.Ctx)

		assert.True(
			t,
			valueWithInnerDependency.(*test.DependencyStruct) == valueRequiringContext,
			"valueWithInnerDependency interface should be created from valueRequiringContext",
		)

		context.String(http.StatusTeapot, "%s", test.Response)
	}
}

func TestIRoutesImpl_GET(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"successfully register handler and handle request": func(t *testing.T) {
			r := setupRouterWithProviders()
			testHandlerFn := setupTestHandlerFn(t)

			registrationError := r.Handle(http.MethodGet, testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodGet).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, test.Response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := Adapt(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			registrationError := r.Handle(http.MethodGet, testEndpoint, testHandlerFn)

			assert.IsType(t, injection.Error{}, registrationError)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}

func TestIRoutesImpl_POST(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"successfully register handler and handle request": func(t *testing.T) {
			r := setupRouterWithProviders()
			testHandlerFn := setupTestHandlerFn(t)

			registrationError := r.Handle(http.MethodPost, testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodPost).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, test.Response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := Adapt(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			registrationError := r.Handle(http.MethodPost, testEndpoint, testHandlerFn)

			assert.IsType(t, injection.Error{}, registrationError)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}

func TestIRoutesImpl_PUT(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"successfully register handler and handle request": func(t *testing.T) {
			r := setupRouterWithProviders()
			testHandlerFn := setupTestHandlerFn(t)

			registrationError := r.Handle(http.MethodPut, testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodPut).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, test.Response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := Adapt(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			registrationError := r.Handle(http.MethodPut, testEndpoint, testHandlerFn)

			assert.IsType(t, injection.Error{}, registrationError)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}

func TestIRoutesImpl_OPTIONS(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"successfully register handler and handle request": func(t *testing.T) {
			r := setupRouterWithProviders()
			testHandlerFn := setupTestHandlerFn(t)

			registrationError := r.Handle(http.MethodOptions, testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodOptions).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, test.Response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := Adapt(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			registrationError := r.Handle(http.MethodOptions, testEndpoint, testHandlerFn)

			assert.IsType(t, injection.Error{}, registrationError)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}

func TestIRoutesImpl_PATCH(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"successfully register handler and handle request": func(t *testing.T) {
			r := setupRouterWithProviders()
			testHandlerFn := setupTestHandlerFn(t)

			registrationError := r.Handle(http.MethodPatch, testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodPatch).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, test.Response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := Adapt(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			registrationError := r.Handle(http.MethodPatch, testEndpoint, testHandlerFn)

			assert.IsType(t, injection.Error{}, registrationError)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}

func TestIRoutesImpl_DELETE(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"successfully register handler and handle request": func(t *testing.T) {
			r := setupRouterWithProviders()
			testHandlerFn := setupTestHandlerFn(t)

			registrationError := r.Handle(http.MethodDelete, testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodDelete).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, test.Response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := Adapt(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			registrationError := r.Handle(http.MethodDelete, testEndpoint, testHandlerFn)

			assert.IsType(t, injection.Error{}, registrationError)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}

func TestIRoutesImpl_HEAD(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"successfully register handler and handle request": func(t *testing.T) {
			r := setupRouterWithProviders()
			testHandlerFn := setupTestHandlerFn(t)

			registrationError := r.Handle(http.MethodHead, testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodHead).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, test.Response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := Adapt(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			registrationError := r.Handle(http.MethodHead, testEndpoint, testHandlerFn)

			assert.IsType(t, injection.Error{}, registrationError)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}

func TestIRoutesImpl_Use(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"successfully register middleware and intercept request after handle": func(t *testing.T) {
			r := setupRouterWithProviders()

			registrationError := r.Use(func(c *gin.Context) {
				c.Next()

				c.String(http.StatusTeapot, "%s", test.Response)
			})

			req := test.NewRequest(test.Endpoint, http.MethodHead).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, test.Response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := Adapt(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			registrationError := r.Use(testHandlerFn)

			assert.IsType(t, injection.Error{}, registrationError)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}

func TestIRoutesImpl_RegisterController(t *testing.T) {
	tests := map[string]func(t *testing.T){
		"successfully for Controller pointer receiver request handler method": func(t *testing.T) {
			r := setupRouterWithProviders()

			registrationError := r.RegisterController(NewPointerController(t))

			req := test.NewRequest(ctrlPostEndpoint, http.MethodPost).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, test.Response, string(req.Response.Body.Bytes()))
		},
		"successfully for Controller value receiver request handler method": func(t *testing.T) {
			r := setupRouterWithProviders()

			registrationError := r.RegisterController(NewValueController(t))

			req := test.NewRequest(ctrlGetEndpoint, http.MethodGet).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, test.Response, string(req.Response.Body.Bytes()))
		},
		"fail registering Controller with unregistered dependencies": func(t *testing.T) {
			r := Adapt(gin.New())

			registrationError := r.RegisterController(NewPointerController(t))

			assert.IsType(t, injection.Error{}, registrationError)
		},
		"fail registering Controller with incorrect routes mapping": func(t *testing.T) {
			r := Adapt(gin.New())

			registrationError := r.RegisterController(new(InvalidRoutesMapController))

			assert.IsType(t, injection.Error{}, registrationError)
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, testCase)
	}
}
