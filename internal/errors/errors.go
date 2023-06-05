package errors

import (
	"fmt"
	"net/http"
)

var (
	JamfErrNotFound      = Jamf{Message: "not found", Status: http.StatusNotFound}
	JamfErrNotAuthorized = Jamf{Message: "not authorized", Status: http.StatusUnauthorized}
	JamfErrForbidden     = Jamf{Message: "forbidden", Status: http.StatusForbidden}
	JamfErrBadRequest    = Jamf{Message: "bad request", Status: http.StatusBadRequest}
	JamfErrUnhandled     = Jamf{Message: "unhandled Jamf error", Status: http.StatusInternalServerError}
)

var (
	CodeNotFound     = Request{Message: "code not found", Status: http.StatusNotFound}
	CodeExpired      = Request{Message: "code expired", Status: http.StatusGone}
	CodeMismatch     = Request{Message: "code mismatch", Status: http.StatusBadRequest}
	UdidNotSpecified = Request{Message: "UDID not specified", Status: http.StatusBadRequest}
	UdidInvalid      = Request{Message: "UDID invalid", Status: http.StatusBadRequest}
	ExtAttrNotFound  = Request{Message: "extension attribute not found", Status: http.StatusNotFound}
	BadToken         = Request{Message: "malformed or missing token", Status: http.StatusBadRequest}
	InvalidToken     = Request{Message: "invalid token", Status: http.StatusUnauthorized}
)

var (
	UpdateConfigVersionNotSemver     = Request{Message: "version does not conform to semver", Status: http.StatusBadRequest}
	UpdateConfigVersionBadFormat     = Request{Message: "not a valid macOS version format", Status: http.StatusBadRequest}
	UpdateConfigMaxDefferalsNegative = Request{Message: "maxDeferrals cannot be negative", Status: http.StatusBadRequest}
	UpdateConfigPriorityBad          = Request{Message: "updatePriority value not recognised", Status: http.StatusBadRequest}
	UpdateConfigActionBad            = Request{Message: "updateAction value not recognised", Status: http.StatusBadRequest}
)

var (
	RequestSendFailed   = Service{Message: "failed to send request"}
	RequestCreateFailed = Service{Message: "failed to create request"}
	BodyDecodeFailed    = Service{Message: "failed to decode response body"}
	CodeGenFailed       = Service{Message: "failed to generate code"}
)

// Jamf is an error type for errors returned by Jamf
type Jamf struct {
	Message string
	Status  int
}

// Request is an error type for errors in inbound requests
type Request struct {
	Message string
	Status  int
}

// Service is an error type for errors returned by internal service functions
type Service struct {
	Message string
	Err     error
}

func (e Service) Wrap(err error) Service {
	e.Err = err
	return e
}

func (e Service) Error() string {
	return e.Message
}

func (e Service) Unwrap() error {
	return e.Err
}

func (e Jamf) Error() string {
	return fmt.Sprintf("%s (%d)", e.Message, e.Status)
}

func (e Request) Error() string {
	return e.Message
}
