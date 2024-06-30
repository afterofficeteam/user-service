package helper

import (
	"net/http"
	"user-service/src/util/repository/model"

	"github.com/thedevsaddam/renderer"
)

const (
	SUCCESS_MESSSAGE string = "Success"
)

func HandleResponse(w http.ResponseWriter, render *renderer.Render, statusCode int, message interface{}, data interface{}) {
	response := model.BaseResponse{
		Message: message,
		Data:    data,
	}

	render.JSON(w, statusCode, response)
}
