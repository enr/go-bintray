package bintray

import (
	"fmt"
	"github.com/enr/go-commons/lang"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const (
	// response body for /subject/repository/pkg
	RESP_BODY_PKG = `
{"name":"optools","repo":"prova","owner":"enrico","desc":"this entity does...","labels":[],"attribute_names":[],"followers":0,
"created":"2013-03-04T09:50:00.742Z","versions":["0.1","0.1.1","0.4","0.9"],"latest_version":"0.9",
"updated":"2013-03-04T09:50:00.742Z","rating_count":0}`
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the BintrayClient client being tested.
	client *BintrayClient

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

// setup sets up a test HTTP server along with a bintray.BintrayClient that is
// configured to talk to that test server.  Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// bintray client configured to use test server
	client = NewClient(nil, "sub", "api")
	client.BaseURL, _ = url.Parse(server.URL)
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

// smoke test for basic functionality of client.execute
func TestExecute(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := "GET"; m != r.Method {
			t.Errorf("Request method = %v, want %v", r.Method, m)
		}
		fmt.Fprint(w, `{"A":"a"}`)
	})
	req, err := client.newRequestWithBody("GET", "/", "request_data")
	if err != nil {
		t.Errorf("Expected nil error; got %#v.", err)
	}
	response, err := client.execute(req)
	if err != nil {
		t.Errorf("Expected nil error; got %#v.", err)
	}
	testResponse(t, response, `{"A":"a"}`, 200)
}

// test for error response management
func TestExecute_httpError(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", 400)
	})
	req, err := client.newRequestWithBody("GET", "/", "request_data")
	if err != nil {
		t.Errorf("Expected nil error; got %#v.", err)
	}
	response, err := client.execute(req)
	if err != nil {
		t.Errorf("Expected nil error; got %#v.", err)
	}
	testResponse(t, response, "Bad Request", 400)
}

func TestPackageExists_false(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", 404)
	})
	pe, err := client.PackageExists("subject", "repository", "pkg")
	if err != nil {
		t.Errorf("unexpected error thrown %s", err)
	}
	if pe {
		t.Errorf("expected no package")
	}
}
func TestPackageExists_true(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, RESP_BODY_PKG, 200)
	})
	pe, err := client.PackageExists("subject", "repository", "pkg")
	if err != nil {
		t.Errorf("unexpected error thrown %s", err)
	}
	if !pe {
		t.Errorf("expected package")
	}
}

func TestGetVersions(t *testing.T) {
	setup()
	defer teardown()
	expectedVersions := []string{"0.1", "0.1.1", "0.4", "0.9"}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, RESP_BODY_PKG, 200)
	})
	versions, err := client.GetVersions("subject", "repository", "pkg")
	if err != nil {
		t.Errorf("unexpected error thrown %s", err)
	}
	if len(versions) != len(expectedVersions) {
		t.Errorf("expected versions %d but got %d", len(expectedVersions), len(versions))
	}
	for _, v := range versions {
		if !lang.SliceContainsString(expectedVersions, v) {
			t.Errorf("versions %s not expected", v)
		}
	}
}

func TestCreateVersion(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", 200)
	})
	err := client.CreateVersion("subject", "repository", "pkg", "0.1.2")
	if err != nil {
		t.Errorf("unexpected error thrown %s", err)
	}
}

func TestUploadFile(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Printf("Req %s\n", r.RequestURI)
		fmt.Fprint(w, `{"A":"a"}`)
	})
	err := client.UploadFile("subject", "repository", "pkg", "1.2", "", "", "testdata/01.txt", false)
	if err != nil {
		t.Errorf("unexpected error thrown %s", err)
	}
}

func TestPublish(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Printf("Req %s\n", r.RequestURI)
		fmt.Fprint(w, `{"A":"a"}`)
	})
	err := client.Publish("subject", "repository", "pkg", "1.2")
	if err != nil {
		t.Errorf("unexpected error thrown %s", err)
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient(nil, "", "")
	if c.BaseURL.String() != defaultBaseURL {
		t.Errorf("NewClient BaseURL = %v, want %v", c.BaseURL.String(), defaultBaseURL)
	}
	if c.UserAgent != userAgent {
		t.Errorf("NewClient UserAgent = %v, want %v", c.UserAgent, userAgent)
	}
}
func TestAuthenticationClient(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		testHeader(t, r, "Authorization", "Basic dGVzdHN1Yjp0ZXN0YXBp")
		testHeader(t, r, "User-Agent", userAgent)
		fmt.Fprint(w, "ok")
	})
	c := NewClient(nil, "testsub", "testapi")
	c.BaseURL, _ = url.Parse(server.URL)
	req, err := c.newRequestWithBody("GET", "/", "request_data")
	if err != nil {
		t.Errorf("Expected nil error; got %#v.", err)
	}
	response, err := c.execute(req)
	if err != nil {
		t.Errorf("Expected nil error; got %#v.", err)
	}
	testResponse(t, response, "ok", 200)
}

func TestNewRequestWithBody(t *testing.T) {
	c := NewClient(nil, "", "")
	inURL, outURL := "/foo", defaultBaseURL+"foo"
	req, _ := c.newRequestWithBody("GET", inURL, "request_body")
	// test that relative URL was expanded
	if req.URL.String() != outURL {
		t.Errorf("NewRequestWithBody(%v) URL = %v, want %v", inURL, req.URL, outURL)
	}
	// test that default user-agent is attached to the request
	userAgent := req.Header.Get("User-Agent")
	if c.UserAgent != userAgent {
		t.Errorf("NewRequestWithBody() User-Agent = %v, want %v", userAgent, c.UserAgent)
	}
}

// Test handling of an error caused by the internal http client's execute()
// function.  A redirect loop is pretty unlikely to occur within the Bintray
// API, but does allow us to exercise the right code path.
func TestExecute_redirectLoop(t *testing.T) {
	setup()
	defer teardown()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusFound)
	})
	req, _ := client.newRequestWithBody("GET", "/", "")
	_, err := client.execute(req)
	if err == nil {
		t.Error("Expected error to be returned.")
	}
	if err, ok := err.(*url.Error); !ok {
		t.Errorf("Expected a URL error; got %#v.", err)
	}
}

func testResponse(t *testing.T, response *BintrayResponse, expectedBody string, expectedStatusCode int) {
	body, err := response.BodyAsString()
	if err != nil {
		t.Errorf("Error getting response body: %#v.", err)
	}
	actualBody := strings.TrimSpace(body)
	if actualBody != expectedBody {
		t.Errorf("Response body = %v, want %v", actualBody, expectedBody)
	}
	if response.StatusCode != expectedStatusCode {
		t.Errorf("Response status = %v, want %v", response.StatusCode, expectedStatusCode)
	}
}

func testHeader(t *testing.T, r *http.Request, header string, want string) {
	if value := r.Header.Get(header); want != value {
		t.Errorf("Header %s = %s, want: %s", header, value, want)
	}
}

func handler(r *http.Request) {
	reader, err := r.MultipartReader()
	if err != nil {
		//fmt.Println(err)
		//http.Error(w, "not a form", http.StatusBadRequest)
	}
	defer r.Body.Close()

	// read part by part
	for {
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				// end of parts
				break
			}
			//fmt.Println(err)
			//http.Error(w, "bad form part", http.StatusBadRequest)
		}

		/**************************************************
		        //if part.FileName() is empty, skip this iteration.
				if part.FileName() == "" {
					continue
				}
		        fmt.Println("filename: ", part.FileName())
		        fmt.Println("formname: ", part.FormName())
		        // use part.Read to read content
				dst, err := os.Create("/home/sanat/" + part.FileName())
				defer dst.Close()

				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if _, err := io.Copy(dst, part); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				//display success message.
				display(w, "upload", "Upload successful.")
				***************************************************/
		// then
		part.Close()
	}
}
