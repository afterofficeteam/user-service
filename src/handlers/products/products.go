package products

import (
	"encoding/json"
	"net/http"
	"user-service/src/util/client"
	"user-service/src/util/helper"
	"user-service/src/util/middleware"
	"user-service/src/util/repository/model/products"

	"github.com/gorilla/mux"
	"github.com/thedevsaddam/renderer"
)

var (
	responseError map[string]any
)

type Handler struct {
	render *renderer.Render
}

func NewHandler(r *renderer.Render) *Handler {
	return &Handler{render: r}
}

func (h *Handler) CreateShop(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)

	var bReq products.RequesstCreate
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

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)

	var bReq products.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	productChannel := make(chan client.Response)
	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: "http://localhost:3000/api/products",
		QueryParam: []client.QueryParams{
			{Param: "user_id", Value: usrId},
		},
	}

	netClient.Post(bReq, productChannel)
	bResp := <-productChannel
	if bResp.Err != nil {
		if err := json.Unmarshal(bResp.Res, &responseError); err != nil {
			helper.HandleResponse(w, h.render, bResp.StatusCode, bResp.Err, nil)
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

	var response products.UpsertProductResponse
	if err := json.Unmarshal(bResp.Res, &response); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, bResp.StatusCode, "Product created successfully", response)
}

func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query()
	shopId := param.Get("shop_id")
	categoryId := param.Get("category_id")
	name := param.Get("name")
	priceMinStr := param.Get("price_min")
	priceMaxStr := param.Get("price_max")
	isAvailable := param.Get("is_available")
	page := param.Get("page")
	limit := param.Get("limit")

	request := products.GetProductsRequest{
		ShopId:      shopId,
		CategoryId:  categoryId,
		Name:        name,
		PriceMinStr: priceMinStr,
		PriceMaxStr: priceMaxStr,
		IsAvailable: isAvailable,
		Page:        page,
		Limit:       limit,
	}
	shopChannel := make(chan client.Response)
	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: "http://localhost:3000/api/products",
		QueryParam: []client.QueryParams{
			{Param: "shop_id", Value: request.ShopId},
			{Param: "category_id", Value: request.CategoryId},
			{Param: "name", Value: request.Name},
			{Param: "price_min", Value: request.PriceMinStr},
			{Param: "price_max", Value: request.PriceMaxStr},
			{Param: "is_available", Value: request.IsAvailable},
			{Param: "page", Value: request.Page},
			{Param: "limit", Value: request.Limit},
		},
	}

	netClient.Get(nil, shopChannel)
	bResp := <-shopChannel
	if bResp.Err != nil {
		if err := json.Unmarshal(bResp.Res, &responseError); err != nil {
			helper.HandleResponse(w, h.render, bResp.StatusCode, bResp.Err, nil)
			return
		}

		helper.HandleResponse(w, h.render, bResp.StatusCode, responseError, nil)
		return
	}

	if bResp.StatusCode != http.StatusOK {
		if err := json.Unmarshal(bResp.Res, &responseError); err != nil {
			helper.HandleResponse(w, h.render, bResp.StatusCode, bResp.Err, nil)
			return
		}

		helper.HandleResponse(w, h.render, bResp.StatusCode, responseError, nil)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(bResp.Res, &response); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, bResp.StatusCode, "Success", response)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)

	param := mux.Vars(r)
	productId := param["product_id"]

	var bReq products.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	productChannel := make(chan client.Response)
	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: "http://localhost:3000/api/products/" + productId,
		QueryParam: []client.QueryParams{
			{Param: "user_id", Value: usrId},
		},
	}

	netClient.Patch(bReq, productChannel)
	bResp := <-productChannel
	if bResp.Err != nil {
		if err := json.Unmarshal(bResp.Res, &responseError); err != nil {
			helper.HandleResponse(w, h.render, bResp.StatusCode, bResp.Err, nil)
			return
		}

		helper.HandleResponse(w, h.render, bResp.StatusCode, responseError, nil)
		return
	}

	if bResp.StatusCode != http.StatusOK {
		if err := json.Unmarshal(bResp.Res, &responseError); err != nil {
			helper.HandleResponse(w, h.render, bResp.StatusCode, bResp.Err, nil)
			return
		}

		helper.HandleResponse(w, h.render, bResp.StatusCode, responseError, nil)
		return
	}

	var response products.UpsertProductResponse
	if err := json.Unmarshal(bResp.Res, &response); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, bResp.StatusCode, "Product updated successfully", response)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)

	param := mux.Vars(r)
	productId := param["product_id"]

	productChannel := make(chan client.Response)
	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: "http://localhost:3000/api/products/" + productId,
		QueryParam: []client.QueryParams{
			{Param: "user_id", Value: usrId},
		},
	}

	netClient.Delete(nil, productChannel)
	bResp := <-productChannel
	if bResp.Err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, bResp.Err.Error(), nil)
		return
	}

	if bResp.StatusCode != http.StatusOK {
		helper.HandleResponse(w, h.render, bResp.StatusCode, bResp.Res, nil)
		return
	}

	helper.HandleResponse(w, h.render, bResp.StatusCode, "Product deleted successfully", nil)
}
