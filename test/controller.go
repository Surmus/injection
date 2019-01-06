package test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

type PointerController struct {
	primitiveZeroValue int

	PrimitiveValConstant string

	Context context.Context

	valueWithInnerDependency DependencyInterface

	valueRequiringContext *DependencyStruct

	t *testing.T
}

func NewPointerController(t *testing.T) *PointerController {
	return &PointerController{
		PrimitiveValConstant: Constant,
		t:                    t,
	}
}

func (c *PointerController) Routes() map[string]string {
	return map[string]string{Endpoint: "GetTest"}
}

func (c *PointerController) GetTest(context context.Context) {
	assert.NotNil(c.t, c.Context)
	assert.NotNil(c.t, c.valueRequiringContext)
	assert.Equal(c.t, Constant, c.PrimitiveValConstant)
	assert.NotNil(c.t, c.valueWithInnerDependency)
	assert.Equal(c.t, 0, c.primitiveZeroValue)

	assert.Equal(c.t, c.Context, c.valueRequiringContext.Ctx)
	assert.Equal(c.t, c.Context, context)

	assert.True(
		c.t,
		c.valueWithInnerDependency.(*DependencyStruct) == c.valueRequiringContext,
		"valueWithInnerDependency interface should be created from valueRequiringContext",
	)
}

type ValueController struct {
	primitiveValConstant string

	ValueRequiringContext *DependencyStruct

	T *testing.T
}

func NewValueController(t *testing.T) ValueController {
	return ValueController{
		primitiveValConstant: Constant,
		T:                    t,
	}
}

func (c ValueController) Routes() map[string]string {
	return map[string]string{Endpoint: "HandleRequest"}
}

func (c ValueController) HandleRequest(context context.Context, valueWithInnerDependency DependencyInterface) {
	assert.NotNil(c.T, c.ValueRequiringContext)
	assert.Equal(c.T, Constant, c.primitiveValConstant)
	assert.NotNil(c.T, valueWithInnerDependency)

	assert.True(
		c.T,
		valueWithInnerDependency.(*DependencyStruct) == c.ValueRequiringContext,
		"valueWithInnerDependency interface should be created from valueRequiringContext",
	)
}

type InvalidRoutesMapController struct{}

func (c *InvalidRoutesMapController) Routes() map[string]string {
	return map[string]string{Endpoint: "PostTest"}
}
