package jamf

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"net/url"
)

type EraseDeviceCommand struct {
	computer Computer
	passcode string
}

// NewEraseDeviceCommand returns a new EraseDeviceCommand
func NewEraseDeviceCommand(comp Computer, pin string) EraseDeviceCommand {
	var c = EraseDeviceCommand{}

	c.passcode = pin
	c.computer = comp

	return c
}

func (c EraseDeviceCommand) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// https://developer.jamf.com/jamf-pro/reference/createcomputercommandbycommand
	var x struct {
		XMLName struct{} `xml:"computer_command"`
		General struct {
			Command  string `xml:"command"`
			Passcode string `xml:"passcode"`
		} `xml:"general"`
		Computers struct {
			Computer struct {
				Id int `xml:"id"`
			} `xml:"computer"`
		} `xml:"computers"`
	}

	x.Computers.Computer.Id = c.computer.Id
	x.General.Command = "EraseDevice"
	x.General.Passcode = c.passcode

	return e.Encode(&x)
}

// Body returns the XML body for the EraseDeviceCommand
func (c EraseDeviceCommand) Body() ([]byte, error) {
	return xml.Marshal(c)
}

// Request builds a new http.Request for the EraseDeviceCommand with its relative API path, headers and body
func (c EraseDeviceCommand) Request() (*http.Request, error) {
	u, err := url.JoinPath(ClassicAPI, "computercommands", "command", "EraseDevice")
	if err != nil {
		return nil, err
	}

	body, err := c.Body()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/xml")

	return req, nil
}
