package jamf

import (
	"bytes"
	"command-on-demand/internal/logger"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	BaseURL    string
	Auth       BasicAuth `json:"auth"`
	token      *Token
	HTTPClient *http.Client
}

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ClientError struct {
	Status  *int   `json:"status,omitempty"`
	Message string `json:"error,omitempty"`
}

// apiError represents the error body that Jamf *sometimes* returns
type apiError struct {
	HttpStatus int      `json:"httpStatus"`
	Errors     []string `json:"errors"`
}

func (e apiError) Error() string {
	return fmt.Sprintf("HTTP Status: %d, Errors: %s", e.HttpStatus, e.Errors)
}

func (e ClientError) Error() string {
	return e.Message
}

func NewClient(baseUrl string, auth BasicAuth) (*Client, error) {
	c := &Client{
		BaseURL: baseUrl,
		Auth:    auth,
		HTTPClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}

	if err := c.handleToken(); err != nil {
		return nil, err
	}

	return c, nil
}

// classicApi returns a URL string with the Jamf Classic API suffix
func (c *Client) classicApi() string {
	u := url.URL{
		Scheme: "https",
		Host:   c.BaseURL,
		Path:   "JSSResource",
	}

	return u.String()
}

// proApi returns a URL string with the Jamf Pro API suffix
func (c *Client) proApi() string {
	u := url.URL{
		Scheme: "https",
		Host:   c.BaseURL,
		Path:   "api",
	}

	return u.String()
}

// handleToken wraps all functionality for checking token validity and subsequent API calls to claim and renew tokens
func (c *Client) handleToken() error {
	if c.token == nil {
		c.token = &Token{}
	}

	if !c.token.expired() && !c.token.expiringSoon() {
		return nil
	}

	req, _ := http.NewRequest("POST", c.proApi(), nil)
	req.Header.Set("Accept", "application/json")
	u, _ := url.JoinPath(c.proApi(), "/v1/auth")
	if c.token.expired() {
		u, _ = url.JoinPath(u, "/token")
		pu, _ := url.Parse(u)
		req.URL = pu
		req.SetBasicAuth(c.Auth.Username, c.Auth.Password)
		logger.Debug("Jamf token expired, using Basic Auth to renew")
	} else {
		u, _ = url.JoinPath(u, "/keep-alive")
		pu, _ := url.Parse(u)
		req.URL = pu
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.Value))
		logger.Debug("Jamf token nearing expiry, refreshing using current token")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d when getting Jamf API token", resp.StatusCode)
	}

	var t Token

	if err = json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return fmt.Errorf("error decoding response")
	}

	c.token = &t
	logger.Info("successfully acquired new Jamf API token")

	return nil
}

// sendRequest is a helper function for dispatching requests
// centralises logic for handling tokens, headers and request/response body parsing
func (c *Client) sendRequest(req *http.Request, v interface{}) error {
	err := c.handleToken()
	if err != nil {
		logger.Fatalf("unable to fetch or refresh token: %s", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.Value))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		logger.Error("sendRequest: " + err.Error())
		return err
	}

	defer resp.Body.Close()

	var ok = func(status int) bool {
		okStatus := []int{http.StatusOK, http.StatusCreated}
		for _, ok := range okStatus {
			if status == ok {
				return true
			}
		}
		return false
	}(resp.StatusCode)

	if !ok {
		var errResp apiError

		if err = json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			logger.Error("sendRequest: " + errResp.Error())
			return &ClientError{
				Status:  &errResp.HttpStatus,
				Message: fmt.Sprintf("errors: %s", errResp.Errors),
			}
		}

		if resp.StatusCode == http.StatusNotFound {
			logger.Errorf("sendRequest: not found: %s", req.URL)
			return &ClientError{
				Status:  &resp.StatusCode,
				Message: "not found",
			}
		}

		logger.Errorf("sendRequest: unhandled Jamf API error, status code: %d", resp.StatusCode)
		return &ClientError{
			Status:  &resp.StatusCode,
			Message: "unexpected error",
		}
	}

	if v == nil {
		return nil
	}

	if err = json.NewDecoder(resp.Body).Decode(&v); err != nil {
		logger.Error("sendRequest: " + "failed to decode response body")
		return &ClientError{
			Status:  nil,
			Message: "failed to decode response body",
		}
	}

	return nil
}

// GetComputer retrieves a Computer record from Jamf API, or returns an error
func (c *Client) GetComputer(udid string) (Computer, error) {
	u, _ := url.JoinPath(c.classicApi(), "computers", "udid", udid)
	s := struct {
		Computer Computer `json:"computer"`
	}{}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return s.Computer, err
	}

	if err = c.sendRequest(req, &s); err != nil {
		return s.Computer, err
	}

	return s.Computer, nil
}

// SendSimpleCommand sends an API command with an empty body; all arguments being in the URL path.
// can send to either the Jamf Classic or Jamf Pro APIs
func (c *Client) SendSimpleCommand(cmd Commander) error {
	ep, err := cmd.Path()
	if err != nil {
		return err
	}

	var u string
	if cmd.Classic() {
		u, _ = url.JoinPath(c.classicApi(), "computercommands", "command", ep)
	} else {
		u, _ = url.JoinPath(c.proApi(), ep)
	}

	req, err := http.NewRequest("POST", u, nil)

	if err = c.sendRequest(req, nil); err != nil {
		return err
	}

	return nil
}

// SendClassicCommand sends a command with an XML encoded body via the Jamf Classic API
func (c *Client) SendClassicCommand(cmd ClassicCommander) error {
	ep, err := cmd.Path()
	if err != nil {
		return err
	}

	u, _ := url.JoinPath(c.classicApi(), "computercommands", "command", ep)

	body, err := xml.Marshal(&cmd)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/xml")

	if err = c.sendRequest(req, nil); err != nil {
		return err
	}

	return nil
}

// SendProCommand sends a command with a JSON encoded body via the Jamf Pro API
func (c *Client) SendProCommand(cmd ProCommander) error {
	ep, err := cmd.Path()
	if err != nil {
		return err
	}

	u, _ := url.JoinPath(c.classicApi(), ep)

	body, err := json.Marshal(&cmd)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	if err = c.sendRequest(req, nil); err != nil {
		return err
	}

	return nil
}
