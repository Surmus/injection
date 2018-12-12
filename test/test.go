package test

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
)

var Router *gin.Engine

// Init sets up environment to run functional and integration tests
func Init() {
	Router = gin.Default()
}

// MustUnMarshal tries to unmarshal given json into result object, panics when unsuccessful
func MustUnMarshal(rawJson []byte, result interface{}) {
	if err := json.Unmarshal(rawJson, result); err != nil {
		panic(err)
	}
}
