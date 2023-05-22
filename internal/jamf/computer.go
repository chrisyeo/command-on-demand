package jamf

import (
	"fmt"
)

type ExtensionAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type General struct {
	Id           int    `json:"id"`
	Udid         string `json:"udid"`
	Name         string `json:"name"`
	SerialNumber string `json:"serial_number"`
}

type Computer struct {
	General             `json:"general"`
	ExtensionAttributes []ExtensionAttribute `json:"extension_attributes"`
}

// GetExtensionAttribute returns the value for a given extension attribute name, or an error if that EA was not found
func (c Computer) GetExtensionAttribute(name string) (string, error) {
	for _, ea := range c.ExtensionAttributes {
		if ea.Name == name {
			return ea.Value, nil
		}
	}

	return "", fmt.Errorf("no such extension attribute: %s", name)
}
