package zoom

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/google/go-querystring/query"
)

const apiVersion = "/v2"

type HTTPMethod string

const (
	// Get is GET HTTP method
	Get HTTPMethod = http.MethodGet

	// Post is POST HTTP method
	Post HTTPMethod = http.MethodPost

	// Put is PUT HTTP method
	Put HTTPMethod = http.MethodPut

	// Patch is PATCH HTTP method
	Patch HTTPMethod = http.MethodPatch

	// Delete is DELETE HTTP method
	Delete HTTPMethod = http.MethodDelete
)

type Client struct {
	Key string
	Secret string
	Transport http.RoundTripper
	Timeout time.Duration
	endpoint string
	zoomApp ZoomApp
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
		zoomApp: zoomApp,
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

// func (c *Client) tokenRequest(opts requestV2Opts, id string, secret string) ()


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

func (c *Client) executeRequest(opts requestV2Opts, token string) (*http.Response, error) {
	client := c.httpClient()
	request, err := c.httpRequest(opts)
	if err != nil {
		return nil, err
	}
	req, err := c.addRequestAuth(request, token)
	
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	return client.Do(req)
}

const GetTokenPath = "/oauth/token"

// GetTokenOptions are the options for creating and getting an access token
type GetTokenOptions struct {
	Code string `json:"code"`
	GrantType string `json:"grant_type"`
	RedirectUri string `json:"redirect_uri"`
}


type AccessTokenResult struct {
	AccessToken string `json:"access_token"`
	TokenType string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn string `json:"expires_in"`
	Scope string `json:"scope"`
}


func (c *Client) GetToken(code string) (string, error) {
	tokenOptions := GetTokenOptions{
		Code: code,
		GrantType: "authorization_code",
		RedirectUri: c.zoomApp.RedirectURL,
	}

	var buf bytes.Buffer

	// encode token option parameters
	if err := json.NewEncoder(&buf).Encode(tokenOptions); err != nil {
		return "", err
	}

	// set request URL
	requestURL := c.endpoint + GetTokenPath

	request, err := http.NewRequest(string(Post), requestURL, &buf)
	
	if err != nil {
		return "", err
	}

	// Add form type
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// Add authorization
	request.SetBasicAuth(c.zoomApp.ClientID, c.zoomApp.ClientSecret)

	response, err := c.httpClient().Do(request)

	if err != nil {
		return "", err
	}

	var ret = AccessTokenResult{}
	
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	if err := checkError(body); err != nil {
		return "", err
	}

	// Unmarshall into the result
	if err := json.Unmarshal(body, &ret); err != nil {
		return "", err
	}

	return ret.AccessToken, nil
}


// func TokenRequest(id string, secret string) (string, error) {}
