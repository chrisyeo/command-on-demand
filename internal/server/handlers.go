package server

import (
	"command-on-demand/internal/jamf"
	"command-on-demand/internal/logger"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
)

func (s Server) CodeHandler(w http.ResponseWriter, r *http.Request) {
	udid, err := s.checkUDID(r)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, err.Error(), true, "")
	}

	code, err := s.CodeStore.GenerateCode(udid)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, "could not generate code", true, "")
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		json.NewEncoder(w).Encode(struct {
			Code string `json:"code"`
		}{Code: code})
		return
	}

	fmt.Fprint(w, code)
	return
}

func (s Server) checkUDID(r *http.Request) (string, error) {
	vars := mux.Vars(r)

	udid, exists := vars["udid"]
	if !exists {
		return udid, errors.New("UDID not specified")
	}

	_, err := uuid.Parse(udid)
	if err != nil {
		return udid, errors.New("not a valid UDID")
	}

	return udid, nil
}

func (s Server) EraseHandler(w http.ResponseWriter, r *http.Request) {
	udid, err := s.checkUDID(r)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, err.Error(), true, "")
	}

	comp, err := s.jamf.GetComputer(udid)
	if err != nil {
		logger.Error(err)
		cErr, ok := err.(*jamf.ClientError)

		if !ok {
			writeResponse(w, http.StatusInternalServerError, "Error getting computer record", true, "")
		} else {
			var errStatus int
			if cErr.Status != nil {
				errStatus = *cErr.Status
			} else {
				errStatus = http.StatusInternalServerError
			}
			writeResponse(w, errStatus, cErr.Message, true, "jamf")
		}
		return
	}

	eaVal, err := comp.GetExtensionAttribute(s.env[EnvCodeProofExtAttName])
	if err != nil {
		logger.Error("no such extension attribute: ", s.env[EnvCodeProofExtAttName])
		writeResponse(w, http.StatusBadRequest, "Unable to find required extension attribute", true, "jamf")
		return
	}

	code, err := s.CodeStore.GetCodeValue(udid)
	defer s.CodeStore.ExpireCode(udid)
	if err != nil {
		logger.Error(err)
		writeResponse(w, http.StatusUnauthorized, "No valid code found for this UDID", true, "")
		return
	}

	if code != eaVal {
		logger.Error("code match failed.")
		logger.Debugf("code match failed. want: %s, got: %s", code, eaVal)
		writeResponse(w, http.StatusBadRequest, "Code did not match value in Jamf", true, "")
		return
	}

	logger.Info("code match succeeded")
	logger.Debug("sending EraseDevice command with PIN:", s.eraseDevicePin())

	cmd := jamf.NewEraseDeviceCommand(comp, s.eraseDevicePin())
	err = s.jamf.SendClassicCommand(cmd)
	if err != nil {
		cErr, ok := err.(*jamf.ClientError)
		if !ok {
			logger.Error("unhandled error when sending EraseDevice command", err)
			writeResponse(w, http.StatusInternalServerError, "error when sending EraseDevice command", true, "")
		} else {
			var errStatus int
			if cErr.Status != nil {
				errStatus = *cErr.Status
			} else {
				errStatus = http.StatusInternalServerError
			}
			logger.Error("client error when sending EraseDevice command", cErr)
			writeResponse(w, errStatus, cErr.Message, true, "jamf")
		}
		return
	}

	logger.Info("EraseDevice command sent successfully")
	writeResponse(w, http.StatusCreated, "EraseDevice command sent. Prepare thyself!", false, "")
	return
}

func (s Server) SoftwareUpdateHandler(w http.ResponseWriter, r *http.Request) {
	udid, err := s.checkUDID(r)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, err.Error(), true, "")
	}

	comp, err := s.jamf.GetComputer(udid)
	if err != nil {
		logger.Error(err)
		cErr, ok := err.(*jamf.ClientError)

		if !ok {
			writeResponse(w, http.StatusInternalServerError, "Error getting computer record", true, "")
		} else {
			var errStatus int
			if cErr.Status != nil {
				errStatus = *cErr.Status
			} else {
				errStatus = http.StatusInternalServerError
			}
			writeResponse(w, errStatus, cErr.Message, true, "jamf")
		}
		return
	}

	eaVal, err := comp.GetExtensionAttribute(s.env[EnvCodeProofExtAttName])
	if err != nil {
		logger.Error("no such extension attribute: ", s.env[EnvCodeProofExtAttName])
		writeResponse(w, http.StatusBadRequest, "Unable to find required extension attribute", true, "jamf")
		return
	}

	code, err := s.CodeStore.GetCodeValue(udid)
	defer s.CodeStore.ExpireCode(udid)
	if err != nil {
		logger.Error(err)
		writeResponse(w, http.StatusUnauthorized, "No valid code found for this UDID", true, "")
		return
	}

	if code != eaVal {
		logger.Error("code match failed.")
		logger.Debugf("code match failed. want: %s, got: %s", code, eaVal)
		writeResponse(w, http.StatusBadRequest, "Code did not match value in Jamf", true, "")
		return
	}

	logger.Info("code match succeeded")
	logger.Debug("sending Software Update command")

	cmd := jamf.NewSoftwareUpdateCommand(comp, jamf.ForceInstallLatest)
	err = s.jamf.SendProCommand(cmd)
	if err != nil {
		cErr, ok := err.(*jamf.ClientError)
		if !ok {
			logger.Error("unhandled error when sending Software Update command", err)
			writeResponse(w, http.StatusInternalServerError, "error when sending Software Update command", true, "")
		} else {
			var errStatus int
			if cErr.Status != nil {
				errStatus = *cErr.Status
			} else {
				errStatus = http.StatusInternalServerError
			}
			logger.Error("client error when sending Software Update command", cErr)
			writeResponse(w, errStatus, cErr.Message, true, "jamf")
		}
		return
	}

	logger.Info("Software Update command sent successfully")
	writeResponse(w, http.StatusCreated, "Software Update command sent", false, "")
	return
}
