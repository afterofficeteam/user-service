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

// SHOPS SECTION
func (h *Handler) CreateShop(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)
	response := make(map[string]any)

	var bReq products.CreateShopRequest
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	shopChannel := make(chan client.Response)
	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: h.baseURL + "/shops",
		QueryParam: []client.QueryParams{
			{Param: "user_id", Value: usrId},
		},
	}

	netClient.Post(bReq, shopChannel)
	bResp := <-shopChannel
	if bResp.Err != nil {
		if err := json.Unmarshal(bResp.Res, &response); err != nil {
			helper.HandleResponse(w, h.render, bResp.StatusCode, err.Error(), nil)
			return
		}

		helper.HandleResponse(w, h.render, bResp.StatusCode, response, nil)
		return
	}

	if bResp.StatusCode != http.StatusCreated {
		if err := json.Unmarshal(bResp.Res, &response); err != nil {
			helper.HandleResponse(w, h.render, bResp.StatusCode, bResp.Err, nil)
			return
		}

		helper.HandleResponse(w, h.render, bResp.StatusCode, response["message"], nil)
		return
	}

	if err := json.Unmarshal(bResp.Res, &response); err != nil {
		helper.HandleResponse(w, h.render, bResp.StatusCode, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, bResp.StatusCode, "Shop created successfully", response["data"])
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

// PRODUCTS SECTION
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)
	response := make(map[string]any)

	var bReq products.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	productChannel := make(chan client.Response)
	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: h.baseURL + "/products",
		QueryParam: []client.QueryParams{
			{Param: "user_id", Value: usrId},
		},
	}

	netClient.Post(bReq, productChannel)
	bResp := <-productChannel
	if bResp.Err != nil {
		if err := json.Unmarshal(bResp.Res, &response); err != nil {
			helper.HandleResponse(w, h.render, bResp.StatusCode, bResp.Err, nil)
			return
		}

		helper.HandleResponse(w, h.render, bResp.StatusCode, response, nil)
		return
	}

	if bResp.StatusCode != http.StatusCreated {
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

	helper.HandleResponse(w, h.render, bResp.StatusCode, "Product created successfully", response["data"])
}

func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query()
	request := products.GetProductsRequest{
		ShopId:      param.Get("shop_id"),
		CategoryId:  param.Get("category_id"),
		Name:        param.Get("name"),
		PriceMinStr: param.Get("price_min"),
		PriceMaxStr: param.Get("price_max"),
		IsAvailable: param.Get("is_available"),
		Page:        param.Get("page"),
		Limit:       param.Get("limit"),
	}
	response := make(map[string]any)

	shopChannel := make(chan client.Response)
	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: h.baseURL + "/products",
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

		helper.HandleResponse(w, h.render, bResp.StatusCode, response["message"], nil)
		return
	}

	if err := json.Unmarshal(bResp.Res, &response); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, bResp.StatusCode, "Success", response["data"])
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
	response := make(map[string]any)

	productChannel := make(chan client.Response)
	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: h.baseURL + "/products/" + productId,
		QueryParam: []client.QueryParams{
			{Param: "user_id", Value: usrId},
		},
	}

	netClient.Patch(bReq, productChannel)
	bResp := <-productChannel
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

		helper.HandleResponse(w, h.render, bResp.StatusCode, response["message"], nil)
		return
	}

	if err := json.Unmarshal(bResp.Res, &response); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, bResp.StatusCode, "Product updated successfully", response["data"])
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)

	param := mux.Vars(r)
	productId := param["product_id"]

	productChannel := make(chan client.Response)
	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: h.baseURL + "/products/" + productId,
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
