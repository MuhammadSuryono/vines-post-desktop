package exception

import (
	"net/http"

	"github.com/MuhammadSuryono/module-golang-server/http/response"
	"github.com/MuhammadSuryono/module-golang-server/server"
)

type InternalServerError struct {
	Error     string `json:"error"`
	ProcessID string `json:"process_id"`
}

func NewInternalServerError(error string, processID string) InternalServerError {
	return InternalServerError{Error: error, ProcessID: processID}
}

func internalServerErrorHandler(err interface{}) bool {
	exception, ok := err.(InternalServerError)
	if ok {
		server.Context.JSON(http.StatusInternalServerError, response.FailureResponse(
			response.INTERNAL_SERVER_ERROR_STATUS,
			response.INTERNAL_SERVER_ERROR_MESSAGE,
			exception,
		))
		return true
	} else {
		return false
	}
}
