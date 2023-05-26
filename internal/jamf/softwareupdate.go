package jamf

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/mod/semver"
	"regexp"
	"strconv"
)

const (
	UpdateActionDownloadOnly       = "DOWNLOAD_ONLY"
	UpdateActionDownloadAndInstall = "DOWNLOAD_AND_INSTALL"
	UpdatePriorityHigh             = "HIGH"
	UpdatePriorityLow              = "LOW"
)

type SoftwareUpdateCommandConfig struct {
	targetVersion    string
	skipVerify       bool
	updateAction     string
	maxDeferrals     int
	forceRestart     bool
	applyMajorUpdate bool
	priority         string
}

type SoftwareUpdateCommand struct {
	ProCommand
	computer Computer
	config   SoftwareUpdateCommandConfig
}

type SoftwareUpdateCommandBody struct {
	// https://developer.jamf.com/jamf-pro/reference/post_v1-macos-managed-software-updates-send-updates
	DeviceIds      []string `json:"deviceIds"`
	SkipVersVerify bool     `json:"skipVersionVerification,omitempty"`
	ApplyMajor     bool     `json:"applyMajorUpdate,omitempty"`
	ForceRestart   bool     `json:"forceRestart,omitempty"`
	Priority       string   `json:"priority,omitempty"`
	UpdateAction   string   `json:"updateAction,omitempty"`
	MaxDefer       int      `json:"maxDeferrals,omitempty"`
	Version        string   `json:"version,omitempty"`
}

func (c SoftwareUpdateCommand) Path() (string, error) {
	return "/v1/macos-managed-software-updates/send-updates", nil
}

func (c SoftwareUpdateCommand) MarshalJSON() ([]byte, error) {
	b := SoftwareUpdateCommandBody{
		DeviceIds:      []string{strconv.Itoa(c.computer.Id)},
		SkipVersVerify: c.config.skipVerify,
		ApplyMajor:     c.config.applyMajorUpdate,
		ForceRestart:   c.config.forceRestart,
		Priority:       c.config.priority,
		UpdateAction:   c.config.updateAction,
		MaxDefer:       c.config.maxDeferrals,
		Version:        c.config.targetVersion,
	}

	return json.Marshal(&b)
}

func NewSoftwareUpdateCommand(comp Computer, c SoftwareUpdateCommandConfig) SoftwareUpdateCommand {
	cmd := SoftwareUpdateCommand{
		computer: comp,
		config:   c,
	}

	return cmd
}

// NewSoftwareUpdateCommandForceLatest is a convenience function for building an often desired command type;
// an ASAP update to the latest available OS version (including major releases) and forcing a restart
func NewSoftwareUpdateCommandForceLatest(comp Computer) SoftwareUpdateCommand {
	conf, _ := NewSoftwareUpdateConfig(
		"", true, UpdateActionDownloadAndInstall,
		0, true, true, UpdatePriorityHigh)

	return NewSoftwareUpdateCommand(comp, conf)
}

func NewSoftwareUpdateConfig(targetVersion string, skipVerify bool, updateAction string,
	maxDeferrals int, forceRestart bool, applyMajorUpdate bool, priority string) (SoftwareUpdateCommandConfig, error) {

	if updateAction != UpdateActionDownloadOnly && updateAction != UpdateActionDownloadAndInstall {
		return SoftwareUpdateCommandConfig{}, fmt.Errorf("updateAction not recognised")
	}

	if priority != UpdatePriorityHigh && priority != UpdatePriorityLow {
		return SoftwareUpdateCommandConfig{}, fmt.Errorf("value for priority not recognised")
	}

	if maxDeferrals < 0 {
		return SoftwareUpdateCommandConfig{}, fmt.Errorf("maxDeferrals cannot be negative")
	}

	if targetVersion != "" {
		match := regexp.MustCompile(`^\d+(\.\d+)*$`).MatchString(targetVersion)
		if !match {
			return SoftwareUpdateCommandConfig{},
				errors.New("version may only contain numbers separated by dots and must start and end with a number")
		}

		tv := "v" + targetVersion
		ok := semver.IsValid(tv)
		if !ok {
			return SoftwareUpdateCommandConfig{}, errors.New("version is invalid: does not conform to semver")
		}
	}

	c := SoftwareUpdateCommandConfig{
		targetVersion:    targetVersion,
		skipVerify:       skipVerify,
		updateAction:     updateAction,
		maxDeferrals:     maxDeferrals,
		forceRestart:     forceRestart,
		applyMajorUpdate: applyMajorUpdate,
		priority:         priority,
	}

	return c, nil
}
