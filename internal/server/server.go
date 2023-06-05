package server

import (
	"command-on-demand/internal/errors"
	"command-on-demand/internal/jamf"
	"command-on-demand/internal/logger"
	"encoding/json"
	e "errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
)

const (
	eraseDevicePin = "000000"
)

type Server struct {
	env       Environment
	jamf      *jamf.Client
	CodeStore *CodeStore
}

// ServiceResponse represents the return body for request responses
type ServiceResponse struct {
	Status      *int   `json:"status,omitempty"`
	Message     string `json:"message,omitempty"`
	IsError     bool   `json:"error"`
	ErrorOrigin string `json:"errorOrigin,omitempty"`
}

func (e ServiceResponse) Error() string {
	if !e.IsError {
		return ""
	}
	return fmt.Sprintf("%s. status: %d", e.Message, e.Status)
}

func NewServer() Server {
	env, err := NewEnvironment("CMDOD_")
	if err != nil {
		logger.Fatal(err)
	}

	auth := jamf.BasicAuth{
		Username: env[EnvJamfAPIUser],
		Password: env[EnvJamfAPIPassword],
	}

	client, err := jamf.NewClient(env[EnvJamfFQDN], auth)
	if err != nil {
		logger.Fatal(err)
	}

	store := NewCodeStore()

	svc := Server{jamf: client, env: env, CodeStore: store}

	return svc
}

func (s Server) token() string {
	return s.env[EnvServerBearerToken]
}

func (s Server) ListenInterface() string {
	i, ok := s.env[EnvServiceListenInterface]
	if !ok {
		return "0.0.0.0"
	}

	if net.ParseIP(i) == nil {
		return "0.0.0.0"
	}

	return i
}

func (s Server) ListenPort() string {
	p, ok := s.env[EnvServiceListenPort]
	if !ok {
		return "8080"
	}

	_, err := strconv.Atoi(p)
	if err != nil {
		return "8080"
	}

	return p
}

// writeResponse writes a ServiceResponse to the response body and sets the response status
func writeResponse(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)

	r := ServiceResponse{
		Status:  &status,
		Message: msg,
		IsError: false,
	}
	json.NewEncoder(w).Encode(&r)
}

// writeErrorResponse writes an error to the response body and sets the response status
func writeErrorResponse(w http.ResponseWriter, err error) {
	status, msg, origin := classifyError(err)
	w.WriteHeader(status)

	r := ServiceResponse{
		Status:      &status,
		Message:     msg,
		IsError:     true,
		ErrorOrigin: origin,
	}
	json.NewEncoder(w).Encode(&r)
}

// classifyError determines the status code and message for a given error
func classifyError(err error) (status int, msg string, origin string) {
	var sErr errors.Service
	if e.As(err, &sErr) {
		status = http.StatusInternalServerError
		msg = sErr.Message
		origin = "service"
		wErr := e.Unwrap(sErr)
		if wErr != nil {
			logger.Debugf("service error: %s: %s", msg, wErr)
		}
		return
	}

	var jErr errors.Jamf
	if e.As(err, &jErr) {
		status = jErr.Status
		msg = jErr.Message
		origin = "jamf"
		return
	}

	var rErr errors.Request
	if e.As(err, &rErr) {
		status = rErr.Status
		msg = rErr.Message
		origin = "request"
		return
	}

	status = http.StatusInternalServerError
	msg = "unknown or unhandled error"
	origin = "unknown"
	return
}
