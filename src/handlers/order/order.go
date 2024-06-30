package order

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/thedevsaddam/renderer"
	"net/http"
	"user-service/src/util/client"
	"user-service/src/util/helper"
	"user-service/src/util/middleware"
	"user-service/src/util/repository/model/order"
)

type Handler struct {
	render *renderer.Render
}

const (
	createOrderUrl      = "http://localhost:9993/order/create"
	createOrderItemsUrl = "http://localhost:9993/order/create/items"
	createOrderItemLogs = "http://localhost:9993/order/create/items/logs"
)

func NewHandler(r *renderer.Render) *Handler {
	return &Handler{render: r}
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

	if bReq.UserID == uuid.Nil || bReq.PaymentTypeID == uuid.Nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "user id or payment type id is required", nil)
		return
	}

	if bReq.OrderNumber == "" || bReq.Status == "" {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "order number or status is required", nil)
		return
	}

	if bReq.TotalPrice == 0 {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "total price is required", nil)
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

	var response *uuid.UUID
	if err := json.Unmarshal(bResp.Res, &response); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, bResp.StatusCode, helper.SUCCESS_MESSSAGE, response)
}

func (h *Handler) CreateOrderItems(w http.ResponseWriter, r *http.Request) {
	var bReq order.OrderItems
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if bReq.OrderID == uuid.Nil || bReq.ProductID == uuid.Nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "order id or product id is required", nil)
		return
	}

	if bReq.Qty == 0 {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "qty is required", nil)
		return
	}

	if bReq.Price == 0 || bReq.SubtotalPrice == 0 {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "price or subtotal price is required", nil)
		return
	}

	if bReq.ProductName == "" {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "product name is required", nil)
		return
	}

	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: createOrderItemsUrl,
	}
	createChan := make(chan client.Response)
	go netClient.Post(bReq, createChan)
	bResp := <-createChan
	if bResp.Err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, bResp.Err.Error(), nil)
		return
	}

	var response *uuid.UUID
	if err := json.Unmarshal(bResp.Res, &response); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, bResp.StatusCode, helper.SUCCESS_MESSSAGE, response)
}

func (h *Handler) CreateOrderItemlogs(w http.ResponseWriter, r *http.Request) {
	var bReq order.OrderItemsLogs
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if bReq.OrderID == uuid.Nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "order id is required", nil)
		return
	}

	if bReq.FromStatus == "" || bReq.ToStatus == "" {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "from status or to status is required", nil)
		return
	}

	if bReq.Notes == "" {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "notes is required", nil)
		return
	}

	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: createOrderItemLogs,
	}
	createChan := make(chan client.Response)
	go netClient.Post(bReq, createChan)
	bResp := <-createChan
	if bResp.Err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, bResp.Err.Error(), nil)
		return
	}

	var response *uuid.UUID
	if err := json.Unmarshal(bResp.Res, &response); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, bResp.StatusCode, helper.SUCCESS_MESSSAGE, response)
}
