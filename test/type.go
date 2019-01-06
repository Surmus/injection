package test

import (
	"context"
)

type DependencyStruct struct {
	Ctx context.Context
}

func (t *DependencyStruct) InnerDependency() context.Context {
	return t.Ctx
}

type DependencyInterface interface {
	InnerDependency() context.Context
}
