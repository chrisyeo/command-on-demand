package server

import (
	"command-on-demand/internal/errors"
	"command-on-demand/internal/jamf"
	"command-on-demand/internal/logger"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// CodeHandler returns a new code for the given udid
func (s Server) CodeHandler(w http.ResponseWriter, r *http.Request) {
	udid, err := s.checkUDID(r)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	code, err := s.CodeStore.NewCode(udid)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		json.NewEncoder(w).Encode(struct {
			Code string `json:"code"`
		}{Code: code.value})
		return
	}

	fmt.Fprint(w, code.value)
	return
}

// validateRequest returns a populated computer object if the udid is valid and a valid code is present in Jamf
func (s Server) validateRequest(r *http.Request) (comp jamf.Computer, err error) {
	udid, err := s.checkUDID(r)
	if err != nil {
		return
	}

	comp, err = s.jamf.GetComputer(udid)
	if err != nil {
		logger.Error("could not get computer from Jamf: ", err)
		return
	}

	err = s.checkCode(comp)
	if err != nil {
		logger.Error("code match failed: ", err)
		return
	}

	return
}

// checkUDID returns the udid from the request if it is valid
func (s Server) checkUDID(r *http.Request) (string, error) {
	vars := mux.Vars(r)

	udid, exists := vars["udid"]
	if !exists {
		return udid, errors.UdidNotSpecified
	}

	_, err := uuid.Parse(udid)
	if err != nil {
		return udid, errors.UdidInvalid
	}

	return udid, nil
}

// checkCode returns an error if the code in Jamf does not match the code in the request or if the code has expired
// an error is also returned if the extension attribute is not present on the computer record
func (s Server) checkCode(comp jamf.Computer) (err error) {
	code, err := s.CodeStore.getCode(comp.Udid)
	defer s.CodeStore.ExpireCode(comp.Udid)
	if err != nil {
		return
	}

	eaName := s.env[EnvCodeProofExtAttName]
	eaVal, err := comp.GetExtensionAttribute(eaName)
	if err != nil {
		logger.Debugf("could not get extension attribute: %s", eaName)
		return
	}

	if code.value != eaVal {
		err = errors.CodeMismatch
		logger.Debugf("code mismatch: extension attribute '%s': want: %s, got: %s", eaName, code.value, eaVal)
		return
	}

	return
}

// EraseHandler sends an EraseDevice command to the computer specified in the request
func (s Server) EraseHandler(w http.ResponseWriter, r *http.Request) {
	comp, err := s.validateRequest(r)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	logger.Debug("sending EraseDevice command with PIN: ", eraseDevicePin)

	cmd := jamf.NewEraseDeviceCommand(comp, eraseDevicePin)
	err = s.jamf.SendCommand(cmd)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	logger.Info("EraseDevice command sent successfully")
	writeResponse(w, http.StatusCreated, "EraseDevice command sent. Prepare thyself!")
	return
}

// SoftwareUpdateHandler sends a Software Update command to the computer specified in the request
func (s Server) SoftwareUpdateHandler(w http.ResponseWriter, r *http.Request) {
	comp, err := s.validateRequest(r)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	logger.Debug("sending Software Update command with forceInstallLatest preset")

	cmd := jamf.NewSoftwareUpdateCommand(comp, jamf.ForceInstallLatest)
	err = s.jamf.SendCommand(cmd)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	logger.Info("Software Update command sent successfully")
	writeResponse(w, http.StatusCreated, "Software Update command sent")
	return
}
