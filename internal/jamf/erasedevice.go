package jamf

import (
	"encoding/xml"
)

type EraseDeviceCommand struct {
	ClassicCommand
	name     string
	computer Computer
	passcode string
}

// MarshalXML implements xml.Marshaller to build the XML structure for an EraseDevice command
// https://developer.jamf.com/jamf-pro/reference/createcomputercommandbycommand
func (c EraseDeviceCommand) MarshalXML(e *xml.Encoder, start xml.StartElement) error {

	var tmp struct {
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

	tmp.Computers.Computer.Id = c.computer.Id
	tmp.General.Command = c.name
	tmp.General.Passcode = c.passcode

	return e.Encode(&tmp)
}

func (c EraseDeviceCommand) Path() (string, error) {
	return c.name, nil
}

// NewEraseDeviceCommand returns an EraseDeviceCommand constructed by using a given Computer and pin (passcode)
func NewEraseDeviceCommand(comp Computer, pin string) EraseDeviceCommand {
	var c = EraseDeviceCommand{}

	c.name = "EraseDevice"
	c.passcode = pin
	c.computer = comp

	return c
}
