package bintray

import (
	"io/ioutil"
	"net/http"
)

// Response wraps the Bintray API response.
type Response struct {
	*http.Response
}

func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	return response
}

// BodyAsString returns the response body as string.
func (r *Response) BodyAsString() (string, error) {
	body, err := r.readAndCloseResponseBody()
	return string(body), err
}

// BodyAsBytes returns the response body as bytes slice.
func (r *Response) BodyAsBytes() ([]byte, error) {
	return r.readAndCloseResponseBody()
}

// BodyAsString returns the response body as string.
func (r *Response) readAndCloseResponseBody() ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	return body, err
}
