package order

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"user-service/src/util/client"
	"user-service/src/util/helper"
	"user-service/src/util/middleware"
	"user-service/src/util/repository/model/order"
	"user-service/src/util/repository/model/products"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/thedevsaddam/renderer"
)

var (
	responseError = map[string]interface{}{}
)

type Handler struct {
	render    *renderer.Render
	validator *validator.Validate
	mutex     *sync.Mutex
	clientKey string
	serverKey string
}

const (
	createOrderUrl   = "http://localhost:9993/order/create"
	getproductUrl    = "http://localhost:3000/api/products"
	updateProductUrl = "http://localhost:3000/api/product-stocks"
	paymentUrl       = "http://localhost:4000/api/payments"
)

func NewHandler(r *renderer.Render, validator *validator.Validate, mutex *sync.Mutex, clientKey, serverKey string) *Handler {
	return &Handler{render: r, validator: validator, mutex: mutex, clientKey: clientKey, serverKey: serverKey}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Get user id from context, obtained from token
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)
	uid := uuid.MustParse(usrId)

	var bReq order.CreateOrderRequest

	// Get limit from query param
	param := r.URL.Query()
	bReq.Limit = param.Get("limit")

	// Decode from body request to struct
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, err.Error(), nil)
		return
	}
	bReq.UserID = uid

	// Channel for each request
	getProductChannel := make(chan client.Response)
	createOrderChannel := make(chan client.Response)
	updateStockChannel := make(chan client.Response)
	paymentChannel := make(chan client.Response)

	// Get data product from product service
	var productIDs []string
	for _, product := range bReq.ProductOrder {
		productIDs = append(productIDs, product.ProductID)
	}
	productId := strings.Join(productIDs, ",")

	// Lock the mutex before accessing the critical section
	h.mutex.Lock()
	defer h.mutex.Unlock()

	netClientGetProducts := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: getproductUrl,
		QueryParam: []client.QueryParams{
			{Param: "product_ids", Value: productId},
			{Param: "limit", Value: bReq.Limit},
		},
	}

	netClientGetProducts.Get(nil, getProductChannel)
	responseGetProducts := <-getProductChannel
	if responseGetProducts.Err != nil || responseGetProducts.StatusCode != http.StatusOK {
		if err := json.Unmarshal(responseGetProducts.Res, &responseError); err != nil {
			helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
			return
		}

		helper.HandleResponse(w, h.render, responseGetProducts.StatusCode, responseError, nil)
		return
	}

	var dataProducts products.DataProduct
	if err := json.Unmarshal(responseGetProducts.Res, &dataProducts); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Handle check if product < order product
	for _, prod := range dataProducts.Data.Items {
		for _, orderProd := range bReq.ProductOrder {
			if prod.Id == orderProd.ProductID && prod.Stock < orderProd.Qty {
				helper.HandleResponse(w, h.render, http.StatusBadRequest, "Product out of stock", nil)
				return
			}
		}
	}

	// Calucate subtotal based on price (data products) * qty (input user)
	for i, orderProd := range bReq.ProductOrder {
		for _, prod := range dataProducts.Data.Items {
			if prod.Id == orderProd.ProductID {
				bReq.ProductOrder[i].Price = prod.Price
				bReq.ProductOrder[i].SubtotalPrice = prod.Price * float64(orderProd.Qty)
			}
		}
	}

	// Calculate total prices based on subtotal price
	var total float64
	for _, orderProd := range bReq.ProductOrder {
		total += orderProd.SubtotalPrice
	}
	bReq.TotalPrice = total

	// Create order to order services
	netClientCreateOrder := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: createOrderUrl,
	}

	netClientCreateOrder.Post(bReq, createOrderChannel)
	responseCreateOrder := <-createOrderChannel
	if responseCreateOrder.Err != nil || responseCreateOrder.StatusCode != http.StatusCreated {
		helper.HandleResponse(w, h.render, responseCreateOrder.StatusCode, responseCreateOrder.Res, nil)
		return
	}

	var orderID string
	if err := json.Unmarshal(responseCreateOrder.Res, &orderID); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Calculate stock product based on data products - qty order
	for _, orderProd := range bReq.ProductOrder {
		for _, prod := range dataProducts.Data.Items {
			if prod.Id == orderProd.ProductID {
				var requestUpdate order.UpdateQtyRequest
				requestUpdate.Stock = prod.Stock - orderProd.Qty
				requestUpdate.ProductId = prod.Id
				bReq.UpdateQty = append(bReq.UpdateQty, requestUpdate)
			}
		}
	}

	// Update stock product to product service
	netClientUpdateStock := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: updateProductUrl,
	}

	netClientUpdateStock.Patch(bReq.UpdateQty, updateStockChannel)
	responseUpdateStock := <-updateStockChannel
	if responseUpdateStock.Err != nil || responseUpdateStock.StatusCode != http.StatusOK {
		if err := json.Unmarshal(responseUpdateStock.Res, &responseError); err != nil {
			helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
			return
		}

		helper.HandleResponse(w, h.render, responseUpdateStock.StatusCode, responseError, nil)
		return
	}

	// Set auth header, payment type, and transaction details for request payment services
	bReq.BasicAuthHeader = "Basic " + base64.StdEncoding.EncodeToString([]byte(h.serverKey+":"))
	bReq.PaymentType = "bank_transfer"
	bReq.TransactionDetails.OrderID = orderID
	bReq.TransactionDetails.GrossAmount = bReq.TotalPrice

	// Create payment to payment services
	netClientPayment := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: paymentUrl,
	}

	netClientPayment.Post(bReq, paymentChannel)
	responsePayment := <-paymentChannel

	var paymentResponse map[string]interface{}
	if responsePayment.Err != nil || responsePayment.StatusCode != http.StatusCreated {
		if err := json.Unmarshal(responsePayment.Res, &paymentResponse); err != nil {
			helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
			return
		}

		helper.HandleResponse(w, h.render, responsePayment.StatusCode, paymentResponse, nil)
		return
	}

	if err := json.Unmarshal(responsePayment.Res, &paymentResponse); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, responsePayment.StatusCode, helper.SUCCESS_MESSSAGE, paymentResponse)
}
