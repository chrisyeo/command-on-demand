package jamf

import (
	"command-on-demand/internal/logger"
	"time"
)

type Token struct {
	Value   string `json:"token"`
	Expires string `json:"expires"`
}

// expired returns true if the current token has more than 0 seconds until it expires
func (t Token) expired() bool {

	if t.timeLeft() > 0 {
		return false
	}

	return true
}

// expiringSoon returns true if the current token has 2 minutes or less until expiry
func (t Token) expiringSoon() bool {
	if t.timeLeft() < 120 {
		return true
	}

	return false
}

// timeLeft returns the number of seconds until the token expires
func (t Token) timeLeft() int {
	if t.Expires == "" {
		return 0
	}

	exp, err := time.Parse(time.RFC3339, t.Expires)
	if err != nil {
		logger.Fatalf("could not parse token expiration date: %s", t.Expires)
	}

	tls := int(time.Until(exp).Seconds())

	if tls < 0 {
		return 0
	}

	return tls
}
