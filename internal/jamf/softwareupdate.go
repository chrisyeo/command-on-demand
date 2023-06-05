package jamf

import (
	"bytes"
	"command-on-demand/internal/errors"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"golang.org/x/mod/semver"
)

const (
	UpdateActionDownloadOnly       = "DOWNLOAD_ONLY"
	UpdateActionDownloadAndInstall = "DOWNLOAD_AND_INSTALL"
	UpdatePriorityHigh             = "HIGH"
	UpdatePriorityLow              = "LOW"
)

var (
	ForceInstallLatest = SoftwareUpdateCommandConfig{
		targetVersion:    "",
		skipVerify:       true,
		updateAction:     UpdateActionDownloadAndInstall,
		maxDeferrals:     0,
		forceRestart:     true,
		applyMajorUpdate: true,
		priority:         UpdatePriorityHigh,
	}
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

func NewSoftwareUpdateCommand(comp Computer, c SoftwareUpdateCommandConfig) SoftwareUpdateCommand {
	cmd := SoftwareUpdateCommand{
		computer: comp,
		config:   c,
	}

	return cmd
}

func NewSoftwareUpdateConfig(targetVersion string, skipVerify bool, updateAction string,
	maxDeferrals int, forceRestart bool, applyMajorUpdate bool, priority string) (SoftwareUpdateCommandConfig, error) {

	if updateAction != UpdateActionDownloadOnly && updateAction != UpdateActionDownloadAndInstall {
		return SoftwareUpdateCommandConfig{}, errors.UpdateConfigActionBad
	}

	if priority != UpdatePriorityHigh && priority != UpdatePriorityLow {
		return SoftwareUpdateCommandConfig{}, errors.UpdateConfigPriorityBad
	}

	if maxDeferrals < 0 {
		return SoftwareUpdateCommandConfig{}, errors.UpdateConfigMaxDefferalsNegative
	}

	if targetVersion != "" {
		match := regexp.MustCompile(`^\d+(\.\d+)*$`).MatchString(targetVersion)
		if !match {
			return SoftwareUpdateCommandConfig{}, errors.UpdateConfigVersionBadFormat
		}

		tv := "v" + targetVersion
		ok := semver.IsValid(tv)
		if !ok {
			return SoftwareUpdateCommandConfig{}, errors.UpdateConfigVersionNotSemver
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

// Body returns the JSON body for the SoftwareUpdateCommand
func (c SoftwareUpdateCommand) Body() ([]byte, error) {
	return c.MarshalJSON()
}

// Request builds a new http.Request for the SoftwareUpdateCommand
func (c SoftwareUpdateCommand) Request() (*http.Request, error) {
	u, err := url.JoinPath(ProAPI, "v1", "macos-managed-software-updates", "send-updates")
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

	req.Header.Add("Content-Type", "application/json")

	return req, nil
}
