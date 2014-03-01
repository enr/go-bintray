package bintray

import (
	"io/ioutil"
	"net/http"
)

type BintrayResponse struct {
	*http.Response
}

func newResponse(r *http.Response) *BintrayResponse {
	response := &BintrayResponse{Response: r}
	return response
}
func (r *BintrayResponse) BodyAsString() (string, error) {
	body, err := r.readAndCloseResponseBody()
	return string(body), err
}
func (r *BintrayResponse) BodyAsBytes() ([]byte, error) {
	return r.readAndCloseResponseBody()
}
func (r *BintrayResponse) readAndCloseResponseBody() ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	return body, err
}
