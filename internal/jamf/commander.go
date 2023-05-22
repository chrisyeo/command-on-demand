package jamf

import (
	"encoding/json"
	"encoding/xml"
)

// Commander is implemented by types that can represent a Jamf device command.
// At the least providing a relative URL path and indication of API version used
type Commander interface {
	Path() (string, error)
	Classic() bool
}

// ClassicCommander is implemented by command types which need to deliver XML based
// commands to Jamf Classic API command endpoints
type ClassicCommander interface {
	Commander
	xml.Marshaler
}

// ProCommander is implemented by command types which need to deliver JSON based
// commands to Jamf Pro API command endpoints.
type ProCommander interface {
	Commander
	json.Marshaler
}

// ClassicCommand is a convenience type which can provide default implementations to
// types implementing the ClassicCommander interface when included in them
type ClassicCommand struct {
	ClassicCommander
	name string
}

func (c ClassicCommand) Classic() bool {
	return true
}

func (c ClassicCommand) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return nil
}

// ProCommand is a convenience type which can provide default implementations to
// types implementing the ProCommander interface when included in them
type ProCommand struct {
	ProCommander
}

func (c ProCommand) Classic() bool {
	return false
}

func (c ProCommand) MarshalJSON() ([]byte, error) {
	return nil, nil
}
