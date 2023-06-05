package jamf

import (
	"net/http"
)

// Commander is the interface for all Jamf commands
// Body returns the unmarshalled XML or JSON command body, or empty []byte for an empty body
// Request returns a new http.Request with the appropriate path, headers and body
type Commander interface {
	Body() ([]byte, error)
	Request() (*http.Request, error)
}
