

package zoom

import (
	"net/http"
)

func (c *Client) addRequestAuth(req *http.Request, token string) (*http.Request, error) {

	// set JWT Authorization header
	req.Header.Add("Authorization", "Bearer "+token)

	return req, nil
}
