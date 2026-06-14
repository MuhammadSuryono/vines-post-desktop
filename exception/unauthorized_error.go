package exception

import (
	"net/http"

	"github.com/MuhammadSuryono/module-golang-server/http/response"
	"github.com/MuhammadSuryono/module-golang-server/server"
)

type UnAuthorizeError struct {
	Error     string `json:"error"`
	ProcessID string `json:"process_id"`
}

func NewUnAuthorizeError(error string, processID string) UnAuthorizeError {
	return UnAuthorizeError{Error: error, ProcessID: processID}
}

func unAuthorizeErrorHandler(err interface{}) bool {
	exception, ok := err.(UnAuthorizeError)
	if ok {
		server.Context.JSON(http.StatusUnauthorized, response.FailureResponse(
			response.UNATHORIZED_STATUS,
			response.UNAUTHORIZED_MESSAGE,
			exception,
		))
		return true
	} else {
		return false
	}
}
