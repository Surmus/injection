package injection

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"injection/test"
	"net/http"
	"testing"
)

const response = "TEST-RESPONSE"
const testEndpoint = "/test-path"
const ctrlPostEndpoint = "/test-path-post"
const ctrlGetEndpoint = "/test-path-get"
const constant = "CONSTANT"

type TestDependencyStruct struct {
	Ctx context.Context
}

func (t *TestDependencyStruct) InnerDependency() context.Context {
	return t.Ctx
}

type TestDependecyInterface interface {
	InnerDependency() context.Context
}

type TestPointerController struct {
	primitiveZeroValue int

	PrimitiveValConstant string

	Context *gin.Context

	valueWithInnerDependency TestDependecyInterface

	valueRequiringContext *TestDependencyStruct

	t *testing.T
}

func (c *TestPointerController) Routes() map[string]string {
	return map[string]string{ctrlPostEndpoint: "PostTest"}
}

func (c *TestPointerController) PostTest() {
	assert.NotNil(c.t, c.Context)
	assert.NotNil(c.t, c.valueRequiringContext)
	assert.Equal(c.t, constant, c.PrimitiveValConstant)
	assert.NotNil(c.t, c.valueWithInnerDependency)
	assert.Equal(c.t, 0, c.primitiveZeroValue)

	assert.Equal(c.t, c.Context, c.valueRequiringContext.Ctx)

	assert.True(
		c.t,
		c.valueWithInnerDependency.(*TestDependencyStruct) == c.valueRequiringContext,
		"valueWithInnerDependency interface should be created from valueRequiringContext",
	)

	c.Context.String(http.StatusTeapot, "%s", response)
}

type TestValueController struct {
	primitiveValConstant string

	context *gin.Context

	ValueWithInnerDependency TestDependecyInterface

	ValueRequiringContext *TestDependencyStruct

	T *testing.T
}

func (c TestValueController) Routes() map[string]string {
	return map[string]string{ctrlGetEndpoint: "GetTest"}
}

func (c TestValueController) GetTest() {
	assert.NotNil(c.T, c.context)
	assert.NotNil(c.T, c.ValueRequiringContext)
	assert.Equal(c.T, constant, c.primitiveValConstant)
	assert.NotNil(c.T, c.ValueWithInnerDependency)

	assert.Equal(c.T, c.context, c.ValueRequiringContext.Ctx)

	assert.True(
		c.T,
		c.ValueWithInnerDependency.(*TestDependencyStruct) == c.ValueRequiringContext,
		"valueWithInnerDependency interface should be created from valueRequiringContext",
	)

	c.context.String(http.StatusTeapot, "%s", response)
}

type TestInvalidRoutesMapController struct{}

func (c *TestInvalidRoutesMapController) Routes() map[string]string {
	return map[string]string{ctrlPostEndpoint: "PostTest"}
}

type ginRoutesMock struct {
	mock.Mock
	*gin.Engine
}

func (r *ginRoutesMock) StaticFile(relativePath, filepath string) gin.IRoutes {
	r.Called(relativePath, filepath)

	return r
}

func (r *ginRoutesMock) Static(relativePath, root string) gin.IRoutes {
	r.Called(relativePath, root)

	return r
}

func (r *ginRoutesMock) StaticFS(relativePath string, fs http.FileSystem) gin.IRoutes {
	r.Called(relativePath, fs)

	return r
}

func setupRouterWithProviders() Routes {
	valueRequiringContextProvider := func(ctx *gin.Context) *TestDependencyStruct {
		return &TestDependencyStruct{Ctx: ctx}
	}

	constantProvider := func() string {
		return constant
	}

	valueWithInnerDependencyProvider := func(dependency *TestDependencyStruct) TestDependecyInterface {
		return dependency
	}

	test.Init()
	r := NewRouter(test.Router)

	r.RegisterProviders(valueWithInnerDependencyProvider, valueRequiringContextProvider)
	r.RegisterProvider(constantProvider)

	return r
}

func setupTestHandlerFn(t *testing.T) interface{} {
	return func(
		context *gin.Context,
		valueWithInnerDependency TestDependecyInterface,
		valueRequiringContext *TestDependencyStruct,
		providedConstant string,
	) {
		assert.NotNil(t, context)
		assert.NotNil(t, valueRequiringContext)
		assert.Equal(t, constant, providedConstant)
		assert.NotNil(t, valueWithInnerDependency)

		assert.Equal(t, context, valueRequiringContext.Ctx)

		assert.True(
			t,
			valueWithInnerDependency.(*TestDependencyStruct) == valueRequiringContext,
			"valueWithInnerDependency interface should be created from valueRequiringContext",
		)

		context.String(http.StatusTeapot, "%s", response)
	}
}

func TestIRoutesImpl_GET(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"successfully register handler and handle request": func(t *testing.T) {
			r := setupRouterWithProviders()
			testHandlerFn := setupTestHandlerFn(t)

			_, registrationError := r.GET(testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodGet).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := NewRouter(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			_, registrationError := r.GET(testEndpoint, testHandlerFn)

			assert.IsType(t, Error{}, registrationError)
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

			_, registrationError := r.POST(testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodPost).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := NewRouter(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			_, registrationError := r.POST(testEndpoint, testHandlerFn)

			assert.IsType(t, Error{}, registrationError)
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

			_, registrationError := r.PUT(testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodPut).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := NewRouter(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			_, registrationError := r.PUT(testEndpoint, testHandlerFn)

			assert.IsType(t, Error{}, registrationError)
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

			_, registrationError := r.OPTIONS(testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodOptions).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := NewRouter(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			_, registrationError := r.OPTIONS(testEndpoint, testHandlerFn)

			assert.IsType(t, Error{}, registrationError)
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

			_, registrationError := r.PATCH(testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodPatch).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := NewRouter(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			_, registrationError := r.PATCH(testEndpoint, testHandlerFn)

			assert.IsType(t, Error{}, registrationError)
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

			_, registrationError := r.DELETE(testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodDelete).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := NewRouter(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			_, registrationError := r.DELETE(testEndpoint, testHandlerFn)

			assert.IsType(t, Error{}, registrationError)
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

			_, registrationError := r.HEAD(testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodHead).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := NewRouter(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			_, registrationError := r.HEAD(testEndpoint, testHandlerFn)

			assert.IsType(t, Error{}, registrationError)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}

func TestIRoutesImpl_Any(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"successfully register handler and handle request": func(t *testing.T) {
			r := setupRouterWithProviders()
			testHandlerFn := setupTestHandlerFn(t)

			_, registrationError := r.Any(testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodGet).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := NewRouter(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			_, registrationError := r.Any(testEndpoint, testHandlerFn)

			assert.IsType(t, Error{}, registrationError)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}

func TestIRoutesImpl_Handle(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"successfully register handler and handle request": func(t *testing.T) {
			r := setupRouterWithProviders()
			testHandlerFn := setupTestHandlerFn(t)

			_, registrationError := r.Handle(http.MethodGet, testEndpoint, testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodGet).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := NewRouter(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			_, registrationError := r.Handle(http.MethodGet, testEndpoint, testHandlerFn)

			assert.IsType(t, Error{}, registrationError)
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, testCase)
	}
}

func TestIRoutesImpl_Use(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"successfully register handler and handle request": func(t *testing.T) {
			r := setupRouterWithProviders()
			testHandlerFn := setupTestHandlerFn(t)

			_, registrationError := r.Use(testHandlerFn)

			req := test.NewRequest(testEndpoint, http.MethodHead).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, response, string(req.Response.Body.Bytes()))
		},
		"fail to register handler with unregistered dependencies": func(t *testing.T) {
			r := NewRouter(gin.New())

			testHandlerFn := setupTestHandlerFn(t)

			_, registrationError := r.Use(testHandlerFn)

			assert.IsType(t, Error{}, registrationError)
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

			_, registrationError := r.RegisterController(&TestPointerController{
				PrimitiveValConstant: constant,
				t:                    t,
			})

			req := test.NewRequest(ctrlPostEndpoint, http.MethodPost).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, response, string(req.Response.Body.Bytes()))
		},
		"successfully for Controller value receiver request handler method": func(t *testing.T) {
			r := setupRouterWithProviders()

			_, registrationError := r.RegisterController(TestValueController{
				primitiveValConstant: constant,
				T:                    t,
			})

			req := test.NewRequest(ctrlGetEndpoint, http.MethodGet).
				MustBuild().Do(test.Router)

			assert.Nil(t, registrationError)
			assert.Equal(t, http.StatusTeapot, req.Response.Code)
			assert.Equal(t, response, string(req.Response.Body.Bytes()))
		},
		"fail registering Controller with unregistered dependencies": func(t *testing.T) {
			r := NewRouter(gin.New())

			_, registrationError := r.RegisterController(&TestPointerController{
				PrimitiveValConstant: constant,
				t:                    t,
			})

			assert.IsType(t, Error{}, registrationError)
		},
		"fail registering Controller with incorrect routes mapping": func(t *testing.T) {
			r := NewRouter(gin.New())

			_, registrationError := r.RegisterController(new(TestInvalidRoutesMapController))

			assert.IsType(t, Error{}, registrationError)
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, testCase)
	}
}

func TestIRoutesImpl_StaticFile(t *testing.T) {
	const testFilePath = "test-path"

	testObj := new(ginRoutesMock)
	testObj.On("StaticFile", testEndpoint, testFilePath).Return(testObj)

	r := NewRouter(testObj)

	r.StaticFile(testEndpoint, testFilePath)

	testObj.AssertExpectations(t)
}

func TestIRoutesImpl_Static(t *testing.T) {
	const testFile = "test-file.ext"

	testObj := new(ginRoutesMock)
	testObj.On("Static", testEndpoint, testFile).Return(testObj)

	r := NewRouter(testObj)

	r.Static(testEndpoint, testFile)

	testObj.AssertExpectations(t)
}

func TestIRoutesImpl_StaticFS(t *testing.T) {
	var testFs = new(http.Dir)

	testObj := new(ginRoutesMock)
	testObj.On("StaticFS", testEndpoint, testFs).Return(testObj)

	r := NewRouter(testObj)

	r.StaticFS(testEndpoint, testFs)

	testObj.AssertExpectations(t)
}
