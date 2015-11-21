package bintray

import (
	"fmt"
	"net/http"
)

// An ErrorResponse reports one or more errors caused by an API request.
type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d", r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode)
}
