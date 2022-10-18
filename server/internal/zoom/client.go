package zoom

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

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
	Transport http.RoundTripper
	Timeout time.Duration
	endpoint string
	logger zerolog.Logger
}


// NewClient returns a new API client
func NewClient() *Client {
	
	logger := log.With().Str("module", "zoom").Logger()
	
	var uri = url.URL{
		Scheme: "https",
		Host: "api.zoom.us",
		Path: apiVersion,
	}

	return &Client{
		endpoint: uri.String(),
		logger: logger,
	}
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

func (c *Client) addRequestAuth(req *http.Request, token string) (*http.Request, error) {

	// set JWT Authorization header
	req.Header.Add("Authorization", "Bearer "+token)

	return req, nil
}


const GetTokenPath = "/oauth/token"

// GetTokenOptions are the options for creating and getting an access token
type GetTokenOptions struct {
	Code string `url:"code"`
	GrantType string `url:"grant_type"`
	RedirectUri string `url:"redirect_uri"`
}


type AccessTokenResult struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int `json:"expires_in"`
	Scope string `json:"scope"`
}


func (c *Client) GetToken(code string) (string, error) {

	c.logger.Info().Msg("Retrieving access token")

	var buf bytes.Buffer

	tokenOptions := GetTokenOptions{
		Code: code,
		GrantType: "authorization_code",
		RedirectUri: "https://rides-centres-profit-disclaimer.trycloudflare.com/auth",
	}

	values, err := query.Values(tokenOptions)
	if err != nil {
		return "", err
	}

	// set request URL
	requestURL := "https://zoom.us/oauth/token"
	if len(values) > 0 {
		requestURL += "?" + values.Encode()
	}

	c.logger.Info().Msgf("Request URL: %s", requestURL)

	request, err := http.NewRequest(string(Post), requestURL, &buf)
	
	if err != nil {
		return "", err
	}

	// Add form type
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// Add authorization
	request.SetBasicAuth("3HFvhLMeRZyz7K6Cjr62Q", "Y6edLK2mnAbDC78bOgoxgQt7DdQay03i")

	response, err := c.httpClient().Do(request)

	if err != nil {
		return "", err
	}

	var ret AccessTokenResult
	
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	c.logger.Info().Msgf("Response Body: %s", string(body))

	if err := checkError(body); err != nil {
		c.logger.Err(err).Msg("Error response returned from zoom api")
		return "", err
	}

	// Unmarshall into the result
	if err := json.Unmarshal(body, &ret); err != nil {
		c.logger.Err(err).Msg("Unmarshall returned error")
		return "", err
	}
	
	return ret.AccessToken, nil
}

const GetDeepLinkPath = "/zoomapp/deeplink"

type GetDeepLinkOptions struct {
	Action string `json:"action"`
}

type Action struct {
	URL string `json:"url"`
	RoleName string `json:"role_name"`
	Verified int `json:"verified"`
	RoleId int `json:"role_id"`
}

type DeepLinkResult struct {
	DeepLink string `json:"deeplink"`
}

func (c *Client) GetDeepLink(token string) (string, error) {
	ac := Action{
		URL: "/",
		RoleName: "Owner",
		Verified: 1,
		RoleId: 0,
	}

	acStr, err := json.Marshal(ac)
	if err != nil {
		return "", err
	}

	getDeepLinkOptions := GetDeepLinkOptions{
		Action: string(acStr),
	}

	response, err := c.executeRequest(requestV2Opts{
										Method: Post,
										Path: GetDeepLinkPath,
										DataParameters: getDeepLinkOptions}, token)

	if err != nil {
		return "", err
	}

	var ret DeepLinkResult

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	log.Printf("Response Body: %s", string(body))

	if err := checkError(body); err != nil {
		return "", err
	}

	// Unmarshall into the result
	if err := json.Unmarshal(body, &ret); err != nil {
		return "", err
	}

	return ret.DeepLink, nil
}
