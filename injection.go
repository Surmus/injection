package injection

import "net/http"

var httpMethods = []string{
	http.MethodPost,
	http.MethodGet,
	http.MethodDelete,
	http.MethodPut,
	http.MethodConnect,
	http.MethodHead,
	http.MethodOptions,
	http.MethodPatch,
	http.MethodTrace,
}
