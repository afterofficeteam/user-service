package cart

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/thedevsaddam/renderer"
	"net/http"
	"user-service/src/util/client"
	"user-service/src/util/helper"
	"user-service/src/util/middleware"
	"user-service/src/util/repository/model/cart"
)

type Handler struct {
	render *renderer.Render
}

const (
	getCartUrl    = "http://localhost:9993/cart/"
	updateCartUrl = "http://localhost:9993/cart/update/"
	addCartUrl    = "http://localhost:9993/cart/add"
	deleteCartUrl = "http://localhost:9993/cart/delete/"
)

func NewHandler(r *renderer.Render) *Handler {
	return &Handler{render: r}
}

func (h *Handler) GetCartByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)

	cartChannel := make(chan client.Response)
	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: getCartUrl + usrId,
	}

	var bReq cart.GetCartRequest
	json.NewDecoder(r.Body).Decode(&bReq)

	go netClient.Get(bReq, cartChannel)
	bResp := <-cartChannel
	if bResp.Err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, bResp.Err.Error(), nil)
		return
	}

	var response []cart.Cart
	if err := json.Unmarshal(bResp.Res, &response); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, bResp.StatusCode, helper.SUCCESS_MESSSAGE, response)
}

func (h *Handler) UpdateCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)

	var bReq cart.Cart
	json.NewDecoder(r.Body).Decode(&bReq)
	cartChannel := make(chan client.Response)
	url := updateCartUrl + usrId

	go client.Put(client.NetClient, url, bReq, cartChannel)
	bResp := <-cartChannel
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

func (h *Handler) AddCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)
	uid := uuid.MustParse(usrId)

	var bReq cart.Cart
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, err.Error(), nil)
		return
	}
	bReq.UserID = uid

	if bReq.Qty == 0 {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "Qty must be greater than 0", nil)
		return
	}

	cartChannel := make(chan client.Response)
	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: addCartUrl,
	}

	go netClient.Post(bReq, cartChannel)
	bResp := <-cartChannel
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

func (h *Handler) DeleteCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)

	var bReq cart.DeleteCartRequest
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, err.Error(), nil)
		return
	}

	url := deleteCartUrl + usrId
	deleteChan := make(chan client.Response)
	httpClient := client.NetClient
	go client.Delete(httpClient, url, bReq, deleteChan)
	bResp := <-deleteChan
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
