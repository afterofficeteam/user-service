package order

import (
	"encoding/json"
	"net/http"
	"user-service/src/util/client"
	"user-service/src/util/helper"
	"user-service/src/util/middleware"
	"user-service/src/util/repository/model/order"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/thedevsaddam/renderer"
)

type Handler struct {
	render    *renderer.Render
	validator *validator.Validate
}

const (
	createOrderUrl      = "http://localhost:9993/order/create"
	createOrderItemsUrl = "http://localhost:9993/order/create/items"
	createOrderItemLogs = "http://localhost:9993/order/create/items/logs"
)

func NewHandler(r *renderer.Render, validator *validator.Validate) *Handler {
	return &Handler{render: r, validator: validator}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the request context which send from token
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)
	uid := uuid.MustParse(usrId)

	var bReq order.Order
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, err.Error(), nil)
		return
	}
	bReq.UserID = uid

	if err := h.validator.Struct(bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, err.Error(), nil)
		return
	}

	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: createOrderUrl,
	}
	createChan := make(chan client.Response)
	go netClient.Post(bReq, createChan)
	bResp := <-createChan
	if bResp.Err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, bResp.Err.Error(), nil)
		return
	}

	var response string
	if err := json.Unmarshal(bResp.Res, &response); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, bResp.StatusCode, helper.SUCCESS_MESSSAGE, response)
}
