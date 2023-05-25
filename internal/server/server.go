package server

import (
	"command-on-demand/internal/jamf"
	"command-on-demand/internal/logger"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
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

func NewService() Server {
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

// eraseDevicePin returns the 6 digit pin used in EraseDevice commands.
// This is currently hard-coded to six zeroes
func (s Server) eraseDevicePin() string {
	return "000000"
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

func writeResponse(w http.ResponseWriter, status int, msg string, isError bool, origin string) {
	w.WriteHeader(status)
	var ogn string
	if isError {
		if origin == "" {
			ogn = "internal"
		} else {
			ogn = origin
		}
	}
	e := ServiceResponse{
		Status:      &status,
		Message:     msg,
		IsError:     isError,
		ErrorOrigin: ogn,
	}
	json.NewEncoder(w).Encode(&e)
}
