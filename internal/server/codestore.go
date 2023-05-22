package server

import (
	"command-on-demand/internal/logger"
	"command-on-demand/internal/util"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

// Code contains the random value and expiry time
type Code struct {
	value   string
	expires time.Time
}

// CodeStore stores Code objects and provides a mutex for safe concurrent access
type CodeStore struct {
	sync.RWMutex
	codes map[string]Code
}

func NewCodeStore() *CodeStore {
	var c CodeStore
	c.codes = make(map[string]Code)
	return &c
}

// GetCodeValue is a convenience wrapper for retrieving a code by UDID
func (c *CodeStore) GetCodeValue(udid string) (string, error) {
	code, err := c.getCode(udid)
	if err != nil {
		return "", err
	}

	return code.value, nil
}

// GenerateCode creates a random code value and associates it with the given UDID in the CodeStore
// An expiration time of 2 minutes is hard-coded
func (c *CodeStore) GenerateCode(udid string) (string, error) {
	v, err := util.RandomBytes(32, true)
	if err != nil {
		logger.Errorf("error getting random string for code: %s", err)
		return "", fmt.Errorf("error when generating code")
	}

	nc := Code{
		value:   v,
		expires: time.Now().Add(2 * time.Minute),
	}

	err = c.setCode(udid, nc)
	if err != nil {
		return "", err
	}

	return nc.value, nil
}

// ExpireCode expires the code associated with a given UDID.
// If the code doesn't exist or is already expired, it's a no-op and errors are not returned
func (c *CodeStore) ExpireCode(udid string) {
	code, err := c.getCode(udid)
	if err != nil {
		return
	}

	if !code.isExpired() {
		code.expireNow()
		_ = c.setCode(udid, *code)
	}
}

// Prune is responsible for removing expired Code objects from CodeStore
// it should be run as a goroutine as part of app initialisation
func (c *CodeStore) Prune(every time.Duration) {
	for range time.Tick(every) {
		logger.Debug("sweeping expired codes")
		c.Lock()
		for udid, code := range c.codes {
			if code.isExpired() {
				delete(c.codes, udid)
				logger.Infof("deleted expired code for %s", udid)
			}
		}
		c.Unlock()
	}
}

// setCode is responsible for adding or replacing an already existing Code in CodeStore
func (c *CodeStore) setCode(udid string, code Code) error {
	_, err := uuid.Parse(udid)
	if err != nil {
		return errors.New("invalid UDID")
	}

	c.Lock()
	defer c.Unlock()
	c.codes[udid] = code

	return nil
}

// getCode will return a Code for the given UDID
// returns errors in cases where the code was not found or was expired
func (c *CodeStore) getCode(udid string) (*Code, error) {
	c.RLock()
	defer c.RUnlock()

	code, ok := c.codes[udid]
	if !ok {
		return nil, errors.New("code not found")
	}

	if code.isExpired() {
		return nil, errors.New("code expired")
	}

	return &code, nil
}

// expireNow is a helper function to force the expiry of a code
func (c *Code) expireNow() {
	c.expires = time.Now()
}

// isExpired is a helper function for checking if a Code has expired
func (c *Code) isExpired() bool {
	if time.Now().After(c.expires) {
		return true
	}

	return false
}
