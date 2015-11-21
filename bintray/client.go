package bintray

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/enr/go-commons/lang"
)

// A Client manages communication with the Bintray API.
type Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client

	subject string
	apikey  string

	// Base URL for API requests. Defaults to the public Bintray API, but can be
	// set to a domain endpoint to use with Bintray Enterprise. BaseURL should
	// always be specified with a trailing slash.
	BaseURL *url.URL

	// User agent used when communicating with the Bintray API.
	UserAgent string

	downloadsHost string
}

// NewClient returns a new Client API client. If a nil httpClient is
// provided, http.DefaultClient will be used.
func NewClient(httpClient *http.Client, subject, apikey string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(defaultBaseURL)
	c := &Client{client: httpClient, BaseURL: baseURL, UserAgent: userAgent, subject: subject, apikey: apikey}
	return c
}

// PackageExists returns if a given package is present in the repository.
// GET /packages/:subject/:repo/:package
func (c *Client) PackageExists(subject, repository, pkg string) (bool, error) {
	if subject == "" || repository == "" || pkg == "" {
		return false, errors.New("PackageExists: subject, repository and package name shouldn't be empty!")
	}
	url := "/packages/" + subject + "/" + repository + "/" + pkg
	req, err := c.newRequestWithReader("GET", url, nil, 0)
	if err != nil {
		return false, err
	}
	resp, err := c.execute(req)
	// we consider 404 an acceptable error in this case.
	if err, ok := err.(*ErrorResponse); ok && err.Response.StatusCode != http.StatusNotFound {
		return false, err
	}
	return (resp.StatusCode == 200), nil
}

// GetVersions returns all versions for the given package.
func (c *Client) GetVersions(subject, repository, pkg string) ([]string, error) {
	if subject == "" || repository == "" || pkg == "" {
		return nil, errors.New("GetVersions: subject, repository and package name shouldn't be empty!")
	}
	url := "/packages/" + subject + "/" + repository + "/" + pkg
	req, err := c.newRequestWithReader("GET", url, nil, 0)
	if err != nil {
		return nil, err
	}
	resp, err := c.execute(req)

	if err != nil {
		fmt.Printf("Error calling %s - %v", url, err)
		return nil, err
	}
	body, err := resp.BodyAsBytes()
	if err != nil {
		return nil, err
	}
	versions, err := lang.ExtractJSONFieldValue(body, "versions")
	if err != nil {
		return nil, err
	}
	return lang.JSONArrayToStringSlice(versions, "versions")
}

// CreateVersionWithMeta creates a new version adding metadata.
//POST /packages/:subject/:repo/:package/versions
func (c *Client) CreateVersionWithMeta(subject, repository, pkg, version string, reqJSON map[string]interface{}) error {
	return c.executeCreateVersion(subject, repository, pkg, version, reqJSON)
}

// CreateVersion creates new version for a package.
//POST /packages/:subject/:repo/:package/versions
func (c *Client) CreateVersion(subject, repository, pkg, version string) error {
	reqJSON := map[string]interface{}{"name": version}
	return c.executeCreateVersion(subject, repository, pkg, version, reqJSON)
}

func (c *Client) executeCreateVersion(subject, repository, pkg, version string, reqJSON map[string]interface{}) error {
	if subject == "" || repository == "" || pkg == "" || version == "" {
		return errors.New("create version: subject, repository, package name and version shouldn't be empty")
	}
	if _, vnameExists := reqJSON["name"]; !vnameExists {
		return errors.New("create version: metadata must contain the name key")
	}
	requestData, err := json.Marshal(reqJSON)
	if err != nil {
		return err
	}
	url := "/packages/" + subject + "/" + repository + "/" + pkg + "/versions"
	req, err := c.newRequestWithBody("POST", url, string(requestData))
	if err != nil {
		return err
	}
	_, err = c.execute(req)
	return err
}

// UploadFile uploads a file into `/content/:subject/:repo/:package/:version/:path`.
func (c *Client) UploadFile(subject, repository, pkg, version, projectGroupID, projectName, filePath, extraArgs string, mavenRepo bool) error {
	fullPath, _ := filepath.Abs(filePath)
	var entityPath string
	var uploadURL string
	fileName := filepath.Base(fullPath)
	if mavenRepo {
		entityPath = strings.Replace(projectGroupID, ".", "/", -1) + "/" + projectName + "/" + version + "/" + fileName
		uploadURL = "content/" + subject + "/" + repository + "/" + pkg + "/" + version + "/" + entityPath
	} else {
		entityPath = version + "/" + fileName
		uploadURL = "content/" + subject + "/" + repository + "/" + pkg + "/" + entityPath + extraArgs
	}
	file, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		return err
	}

	req, err := c.newRequestWithReader("PUT", uploadURL, file, fi.Size())
	if err != nil {
		return err
	}
	_, err = c.execute(req)
	/*
		if serr, ok := err.(httpError); ok {
			if serr.statusCode == 409 {
				//conflict. skip
				//continue but dont publish.
				//TODO - provide an option to replace existing artifact
				//TODO - ?check exists before attempting upload?
			} else {

			}
		} else if err != nil {

		}
	*/
	return err
}

// Publish an uploaded file.
func (c *Client) Publish(subject, repository, pkg, version string) error {
	if subject == "" || repository == "" || pkg == "" || version == "" {
		return errors.New("Publish: subject, repository, package name and version shouldn't be empty!")
	}
	url := "/content/" + subject + "/" + repository + "/" + pkg + "/" + version + "/publish"
	req, err := c.newRequestWithReader("POST", url, nil, 0)
	if err != nil {
		return err
	}
	resp, err := c.execute(req)
	var objmap map[string]*json.RawMessage
	jsonBlob, _ := resp.BodyAsBytes()
	err = json.Unmarshal(jsonBlob, &objmap)
	if err != nil {
		fmt.Println("error:", err)
	}
	filesNum := 0
	if _, ok := objmap["files"]; ok {
		//var filesNum int
		err = json.Unmarshal(*objmap["files"], &filesNum)
		//fmt.Printf("files filesNum=%d\n", filesNum)
	}
	// return (filesNum > 0), err
	return err
}

// execute sends an API request and returns the API response and error if any.
func (c *Client) execute(req *http.Request) (*Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	response := newResponse(resp)

	err = CheckResponse(resp)
	// even though there was an error, we still return the response
	// in case the caller wants to inspect it further
	return response, err
}

// newRequestWithBody creates an API request using the given string as the body.
// A relative URL can be provided in urlStr, in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash.
func (c *Client) newRequestWithBody(method, urlStr, body string) (*http.Request, error) {
	requestData := []byte(body)
	requestLength := int64(len(requestData))
	requestReader := bytes.NewReader(requestData)

	return c.newRequestWithReader(method, urlStr, requestReader, requestLength)
}

// newRequestWithReader creates an API request.
// A relative URL can be provided in urlStr, in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash.
func (c *Client) newRequestWithReader(method, urlStr string, requestReader io.Reader, requestLength int64) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	u := c.BaseURL.ResolveReference(rel)
	req, err := http.NewRequest(method, u.String(), requestReader)
	if err != nil {
		return nil, err
	}
	if requestLength > 0 {
		req.ContentLength = int64(requestLength)
	}
	req.Header.Add("User-Agent", c.UserAgent)
	if c.subject != "" {
		req.SetBasicAuth(c.subject, c.apikey)
	}
	return req, nil
}

// CheckResponse checks the API response for errors, and returns them if
// present.  A response is considered an error if it has a status code outside
// the 200 range.
func CheckResponse(r *http.Response) error {
	c := r.StatusCode
	if 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	return errorResponse
}
