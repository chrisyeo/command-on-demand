package server

import (
	"command-on-demand/internal/errors"
	"command-on-demand/internal/logger"
	"command-on-demand/internal/util"
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

// NewCodeStore creates a new instance of CodeStore with an empty map of codes
func NewCodeStore() *CodeStore {
	return &CodeStore{
		codes: make(map[string]Code),
	}
}

// NewCode generates a new Code object with a random value and an expiry time of 2 minutes from now.
// The Code object is associated with the given UDID and stored in the CodeStore.
// Returns the newly created Code object and any errors encountered during the process.
func (c *CodeStore) NewCode(udid string) (code *Code, err error) {
	v, err := util.RandomBytes(32, true)
	if err != nil {
		return nil, errors.CodeGenFailed.Wrap(err)
	}

	code = &Code{
		value:   v,
		expires: time.Now().Add(2 * time.Minute),
	}

	c.Lock()
	defer c.Unlock()

	c.codes[udid] = *code

	logger.Debugf("generated new code for %s. Expiry: %s", udid, code.expires)

	return code, nil
}

// ExpireCode will force the expiry of a code for the given UDID
func (c *CodeStore) ExpireCode(udid string) {
	c.Lock()
	defer c.Unlock()

	logger.Debugf("forcing expiry of code for %s", udid)

	delete(c.codes, udid)
}

// Prune is a goroutine that runs every given interval and removes expired codes from the CodeStore
func (c *CodeStore) Prune(every time.Duration) {
	for range time.Tick(every) {
		logger.Debug("pruning expired codes")
		c.Lock()
		for udid, code := range c.codes {
			if code.isExpired() {
				delete(c.codes, udid)
				logger.Debugf("pruned expired code for %s", udid)
			}
		}
		c.Unlock()
	}
}

// getCode returns the Code object for the given UDID if it exists and is not expired.
func (c *CodeStore) getCode(udid string) (*Code, error) {
	c.RLock()
	defer c.RUnlock()

	code, ok := c.codes[udid]
	if !ok {
		return nil, errors.CodeNotFound
	}

	if code.isExpired() {
		logger.Debugf("code for %s expired at: %s", udid, code.expires)
		return nil, errors.CodeExpired
	}

	return &code, nil
}

// isExpired returns true if the code has expired
func (c *Code) isExpired() bool {
	if time.Now().After(c.expires) {
		return true
	}

	return false
}
