package bintray

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/enr/go-commons/lang"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// A Client manages communication with the Bintray API.
type BintrayClient struct {
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

// NewClient returns a new BintrayClient API client. If a nil httpClient is
// provided, http.DefaultClient will be used.
func NewClient(httpClient *http.Client, subject, apikey string) *BintrayClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(defaultBaseURL)
	c := &BintrayClient{client: httpClient, BaseURL: baseURL, UserAgent: userAgent, subject: subject, apikey: apikey}
	return c
}

// Find if package is present
// GET /packages/:subject/:repo/:package
func (c *BintrayClient) PackageExists(subject, repository, pkg string) (bool, error) {
	if subject == "" || repository == "" || pkg == "" {
		return false, errors.New("PackageExists: subject, repository and package name shouldn't be empty!")
	}
	url := "/packages/" + subject + "/" + repository + "/" + pkg
	req, err := c.newRequestWithReader("GET", url, nil, 0)
	if err != nil {
		return false, err
	}
	resp, err := c.execute(req)
	if err != nil {
		return false, err
	}
	return (resp.StatusCode == 200), nil
}

func (c *BintrayClient) GetVersions(subject, repository, pkg string) ([]string, error) {
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

//POST /packages/:subject/:repo/:package/versions
func (c *BintrayClient) CreateVersionWithMeta(subject, repository, pkg, version string, reqJson map[string]interface{}) error {
	return c.executeCreateVersion(subject, repository, pkg, version, reqJson)
}

//POST /packages/:subject/:repo/:package/versions
func (c *BintrayClient) CreateVersion(subject, repository, pkg, version string) error {
	reqJson := map[string]interface{}{"name": version}
	return c.executeCreateVersion(subject, repository, pkg, version, reqJson)
}

func (c *BintrayClient) executeCreateVersion(subject, repository, pkg, version string, reqJson map[string]interface{}) error {
	if subject == "" || repository == "" || pkg == "" || version == "" {
		return errors.New("create version: subject, repository, package name and version shouldn't be empty!")
	}
	if _, vnameExists := reqJson["name"]; !vnameExists {
		return errors.New("create version: metadata must contain the name key")
	}
	requestData, err := json.Marshal(reqJson)
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

//PUT /content/:subject/:repo/:package/:version/:path
func (c *BintrayClient) UploadFile(subject, repository, pkg, version, projectGroupId, projectName, filePath string, mavenRepo bool) error {
	fullPath, _ := filepath.Abs(filePath)
	var entityPath string
	var uploadUrl string
	fileName := filepath.Base(fullPath)
	if mavenRepo {
		entityPath = strings.Replace(projectGroupId, ".", "/", -1) + "/" + projectName + "/" + version + "/" + fileName
		uploadUrl = "content/" + subject + "/" + repository + "/" + pkg + "/" + version + "/" + entityPath
	} else {
		entityPath = version + "/" + fileName
		uploadUrl = "content/" + subject + "/" + repository + "/" + pkg + "/" + entityPath
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

	req, err := c.newRequestWithReader("PUT", uploadUrl, file, fi.Size())
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

// cambiare a return boolean, error
// true se body.files > 0
func (c *BintrayClient) Publish(subject, repository, pkg, version string) error {
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
func (c *BintrayClient) execute(req *http.Request) (*BintrayResponse, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return newResponse(resp), err
}

// newRequestWithBody creates an API request using the given string as the body.
// A relative URL can be provided in urlStr, in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash.
func (c *BintrayClient) newRequestWithBody(method, urlStr, body string) (*http.Request, error) {
	requestData := []byte(body)
	requestLength := int64(len(requestData))
	requestReader := bytes.NewReader(requestData)

	return c.newRequestWithReader(method, urlStr, requestReader, requestLength)
}

// newRequestWithReader creates an API request.
// A relative URL can be provided in urlStr, in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash.
func (c *BintrayClient) newRequestWithReader(method, urlStr string, requestReader io.Reader, requestLength int64) (*http.Request, error) {
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
