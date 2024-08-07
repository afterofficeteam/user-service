package shop

import (
	"encoding/json"
	"net/http"
	"user-service/src/util/client"
	"user-service/src/util/helper"
	"user-service/src/util/middleware"
	"user-service/src/util/repository/model/products"

	"github.com/thedevsaddam/renderer"
)

var (
	responseError map[string]any
)

type Handler struct {
	render  *renderer.Render
	baseURL string
}

func NewHandler(r *renderer.Render) *Handler {
	return &Handler{
		render:  r,
		baseURL: "http://localhost:3000/api",
	}
}

func (h *Handler) CreateShop(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)

	var bReq products.CreateShopRequest
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	shopChannel := make(chan client.Response)
	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: "http://localhost:3000/api/shops",
		QueryParam: []client.QueryParams{
			{Param: "user_id", Value: usrId},
		},
	}

	netClient.Post(bReq, shopChannel)
	bResp := <-shopChannel
	if bResp.Err != nil {
		if err := json.Unmarshal(bResp.Res, &responseError); err != nil {
			helper.HandleResponse(w, h.render, bResp.StatusCode, err.Error(), nil)
			return
		}

		helper.HandleResponse(w, h.render, bResp.StatusCode, responseError, nil)
		return
	}

	if bResp.StatusCode != http.StatusCreated {
		if err := json.Unmarshal(bResp.Res, &responseError); err != nil {
			helper.HandleResponse(w, h.render, bResp.StatusCode, bResp.Err, nil)
			return
		}

		helper.HandleResponse(w, h.render, bResp.StatusCode, responseError, nil)
		return
	}

	var resp products.Response
	if err := json.Unmarshal(bResp.Res, &resp); err != nil {
		helper.HandleResponse(w, h.render, bResp.StatusCode, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, bResp.StatusCode, "Shop created successfully", resp)
}

func (h *Handler) GetShops(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query()
	request := products.GetShopsRequest{
		UserId:   param.Get("user_id"),
		ShopName: param.Get("shop_name"),
		Page:     param.Get("page"),
		Limit:    param.Get("limit"),
	}
	response := make(map[string]any)

	shopChannel := make(chan client.Response)
	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: h.baseURL + "/shops",
		QueryParam: []client.QueryParams{
			{Param: "user_id", Value: request.UserId},
			{Param: "shop_name", Value: request.ShopName},
			{Param: "page", Value: request.Page},
			{Param: "limit", Value: request.Limit},
		},
	}

	netClient.Get(nil, shopChannel)
	bResp := <-shopChannel
	if bResp.Err != nil {
		if err := json.Unmarshal(bResp.Res, &response); err != nil {
			helper.HandleResponse(w, h.render, bResp.StatusCode, bResp.Err, nil)
			return
		}
		helper.HandleResponse(w, h.render, bResp.StatusCode, response, nil)
		return
	}

	if bResp.StatusCode != http.StatusOK {
		if err := json.Unmarshal(bResp.Res, &response); err != nil {
			helper.HandleResponse(w, h.render, bResp.StatusCode, bResp.Err, nil)
			return
		}
		helper.HandleResponse(w, h.render, bResp.StatusCode, response, nil)
		return
	}

	if err := json.Unmarshal(bResp.Res, &response); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, bResp.StatusCode, "Success", response["data"])
}
