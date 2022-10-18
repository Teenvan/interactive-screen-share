package zoom

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/google/go-querystring/query"
)

const apiVersion = "/v2"

type HTTPMethod string

type Client struct {
	Key string
	Secret string
	Transport http.RoundTripper
	Timeout time.Duration
	endpoint string
}


// NewClient returns a new API client
func NewClient() (*Client, error) {
	
	zoomApp, err := GetZoomApp()
	
	if err != nil {
		return nil, err
	}

	var uri = url.URL{
		Scheme: "https",
		Host: zoomApp.Host,
		Path: apiVersion,
	}

	return &Client{
		Key: zoomApp.ClientID,
		Secret: zoomApp.ClientSecret,
		endpoint: uri.String(),
	}, nil
}

type requestV2Opts struct {
	Client         *Client
	Method         HTTPMethod
	URLParameters  interface{}
	Path           string
	DataParameters interface{}
	Ret            interface{}
	// HeadResponse represents responses that don't have a body
	HeadResponse bool
}


func (c *Client) httpClient() *http.Client {
	client := &http.Client{Transport: c.Transport}
	if c.Timeout > 0 {
		client.Timeout = c.Timeout
	}

	return client
}


func (c *Client) httpRequest(opts requestV2Opts) (*http.Request, error) {
	var buf bytes.Buffer

	// encode body parameters if any
	if err := json.NewEncoder(&buf).Encode(&opts.DataParameters); err != nil {
		return nil, err
	}

	// set URL parameters
	values, err := query.Values(opts.URLParameters)
	if err != nil {
		return nil, err
	}

	// set request URL
	requestURL := c.endpoint + opts.Path
	if len(values) > 0 {
		requestURL += "?" + values.Encode()
	}

	// create HTTP request
	return http.NewRequest(string(opts.Method), requestURL, &buf)
}

func (c *Client) executeRequest(opts requestV2Opts) (*http.Response, error) {
	client := c.httpClient()
	req, err := c.addRequestAuth(c.httpRequest(opts))
	
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	return client.Do(req)
}

func (c *Client) requestV2(opts requestV2Opts) error {
	// execute HTTP request
	resp, err := c.executeRequest(opts)
	if err != nil {
		return err
	}

	// If there is no body in response
	if opts.HeadResponse {
		return c.requestV2HeadOnly(resp)
	}

	return c.requestV2WithBody(opts, resp)
}

func (c *Client) requestV2WithBody(opts requestV2Opts, resp *http.Response) error {
	// read HTTP response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// check for Zoom errors in the response
	if err := checkError(body); err != nil {
		return err
	}

	// unmarshal the response body into the return object
	return json.Unmarshal(body, &opts.Ret)
}

func (c *Client) requestV2HeadOnly(resp *http.Response) error {
	if resp.StatusCode != 204 {
		return errors.New(resp.Status)
	}

	// there were no errors, just return
	return nil
}


// func TokenRequest(id string, secret string) (string, error) {}


func GetToken(code string, verifier string) {}
