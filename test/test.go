package test

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

const Response = "TEST-RESPONSE"
const Constant = "CONSTANT"
const HttpMethod = http.MethodGet
const CtxKey = "TEST"
const CtxVal = "TEST"
const Endpoint = "/test"

var Router *gin.Engine

// Init sets up environment to run functional and integration tests
func Init() {
	Router = gin.Default()
}

// MustUnMarshal tries to unmarshal given json into result object, panics when unsuccessful
func MustUnMarshal(rawJSON []byte, result interface{}) {
	if err := json.Unmarshal(rawJSON, result); err != nil {
		panic(err)
	}
}
