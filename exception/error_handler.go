package exception

import (
	"net/http"

	"github.com/MuhammadSuryono/module-golang-server/http/response"
	"github.com/MuhammadSuryono/module-golang-server/server"
)

func ErrorHandler(err interface{}) {
	if internalServerErrorHandler(err) {
		return
	}

	if unAuthorizeErrorHandler(err) {
		return
	}

	server.Context.JSON(http.StatusInternalServerError, response.FailureResponse(
		response.INTERNAL_SERVER_ERROR_STATUS,
		response.INTERNAL_SERVER_ERROR_MESSAGE,
		err,
	))
}
