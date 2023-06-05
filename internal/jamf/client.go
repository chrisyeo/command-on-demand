package jamf

import (
	"command-on-demand/internal/errors"
	"command-on-demand/internal/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	ClassicAPI        = "JSSResource"
	ProAPI            = "api"
	endpointToken     = "/v1/auth/token"
	endpointKeepAlive = "/v1/auth/keep-alive"
)

type Client struct {
	fqdn       string
	auth       BasicAuth
	token      *Token
	httpClient *http.Client
}

type BasicAuth struct {
	Username string
	Password string
}

func NewClient(fqdn string, auth BasicAuth) (*Client, error) {
	c := &Client{
		fqdn: fqdn,
		auth: auth,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}

	if err := c.handleToken(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) baseUrl() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   c.fqdn,
	}
}

// apiBaseUrl returns the Jamf base URL with the given API suffix.
// This is used to switch between the Jamf Classic and Pro APIs
func (c *Client) apiBaseUrl(suffix string) string {
	u := c.baseUrl()
	u.Path = suffix

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

	req, _ := http.NewRequest(http.MethodPost, "", nil)
	req.Header.Set("Accept", "application/json")
	var endpoint string

	if c.token.expired() {
		endpoint = endpointToken
		req.SetBasicAuth(c.auth.Username, c.auth.Password)
		logger.Debug("Jamf token expired, using Basic Auth to renew")
	} else {
		endpoint = endpointKeepAlive
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.Value))
		logger.Debug("Jamf token nearing expiry, refreshing using current token")
	}

	req.URL, _ = url.Parse(c.apiBaseUrl(ProAPI) + endpoint)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.RequestSendFailed.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Jamf{
			Message: "error when requesting API token",
			Status:  resp.StatusCode,
		}
	}

	var t Token

	if err = json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return errors.BodyDecodeFailed.Wrap(err)
	}

	c.token = &t
	logger.Debug("successfully acquired new Jamf API token")

	return nil
}

// sendRequest is a helper function for dispatching requests
// centralises logic for handling tokens, headers and request/response body parsing
func (c *Client) sendRequest(req *http.Request, v interface{}) error {
	err := c.handleToken()
	if err != nil {
		logger.Fatalf("failed to get token: %s", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.Value))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.RequestSendFailed.Wrap(err)
	}

	defer resp.Body.Close()

	// jamf API does not always return a usable error response body, so we need to map common status codes to errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return errors.JamfErrNotFound
		case http.StatusUnauthorized:
			return errors.JamfErrNotAuthorized
		case http.StatusForbidden:
			return errors.JamfErrForbidden
		case http.StatusBadRequest:
			return errors.JamfErrBadRequest
		default:
			return errors.Jamf{
				Message: errors.JamfErrUnhandled.Message,
				Status:  resp.StatusCode,
			}
		}
	}

	if v == nil {
		return nil
	}

	if err = json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return errors.BodyDecodeFailed.Wrap(err)
	}

	return nil
}

// GetComputer retrieves a Computer record from Jamf API, or returns an error
func (c *Client) GetComputer(udid string) (Computer, error) {
	u, _ := url.JoinPath(c.apiBaseUrl(ClassicAPI), "computers", "udid", udid)
	s := struct {
		Computer Computer `json:"computer"`
	}{}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return s.Computer, errors.RequestCreateFailed.Wrap(err)
	}

	if err = c.sendRequest(req, &s); err != nil {
		return s.Computer, err
	}

	return s.Computer, nil
}

func (c *Client) SendCommand(cmd Commander) error {
	req, err := cmd.Request()
	if err != nil {
		return errors.RequestCreateFailed.Wrap(err)
	}

	req.URL = c.baseUrl().ResolveReference(req.URL)
	if err != nil {
		return errors.RequestCreateFailed.Wrap(err)
	}

	if err = c.sendRequest(req, nil); err != nil {
		return err
	}

	return nil
}
