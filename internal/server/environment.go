package server

import (
	"fmt"
	"os"
	"strings"
)

type Environment map[string]string

// Environment variable keys/names
const (
	EnvJamfFQDN               = "JAMF_FQDN"
	EnvJamfAPIUser            = "JAMF_API_USER"
	EnvJamfAPIPassword        = "JAMF_API_PASSWORD"
	EnvServerBearerToken      = "SERVER_BEARER_TOKEN"
	EnvCodeProofExtAttName    = "CODE_PROOF_EA_NAME"
	EnvServiceListenInterface = "SERVER_LISTEN_INTERFACE"
	EnvServiceListenPort      = "SERVER_LISTEN_PORT"
)

// NewEnvironment reads all environment variables prefixed with a given namespace and returns an Environment
// whose keys are set without the namespace prefix.
// namespace is used to prevent collision with other OS env vars, but is then removed when set on Environment
// An error is returned if not all required keys are present and set with a non-empty value
func NewEnvironment(namespace string) (e Environment, err error) {
	e = make(Environment)
	for _, env := range os.Environ() {
		k, v, ok := strings.Cut(env, "=")
		if ok && strings.HasPrefix(k, namespace) {
			e[strings.TrimPrefix(k, namespace)] = v
		}
	}

	for _, ek := range required() {
		v, exists := e[ek]
		if !exists {
			err = fmt.Errorf("missing required env variable: %s%s", namespace, ek)
			return
		} else if v == "" {
			err = fmt.Errorf("cannot have empty value for required variable: %s%s", namespace, ek)
		}
	}

	return
}

// required returns a list of environment variable names/keys which must be present
func required() []string {
	req := []string{
		EnvJamfFQDN,
		EnvJamfAPIUser,
		EnvJamfAPIPassword,
		EnvServerBearerToken,
		EnvCodeProofExtAttName,
	}

	return req
}
