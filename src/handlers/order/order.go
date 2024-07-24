package order

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
	"user-service/src/util/client"
	"user-service/src/util/helper"
	"user-service/src/util/middleware"
	"user-service/src/util/repository/model/order"
	"user-service/src/util/repository/model/payment"
	"user-service/src/util/repository/model/products"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
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
	// Service Order
	createOrderUrl   = "https://8c5c-182-253-51-145.ngrok-free.app/order/create"
	callbackUrl      = "https://8c5c-182-253-51-145.ngrok-free.app/order/callback"
	checkStatusUrl   = "https://8c5c-182-253-51-145.ngrok-free.app/order/status/"
	updateStatusUrl  = "https://localhost:9993/order/status/update"
	updateShppingUrl = "https://localhost:9993/order/shipping/update"

	// Service Product
	getproductUrl    = "https://362c-182-253-51-145.ngrok-free.app/api/products"
	updateProductUrl = "https://362c-182-253-51-145.ngrok-free.app/api/product-stocks"

	// Service Payment
	paymentUrl = "https://51e6-182-253-51-145.ngrok-free.app/api/payments"
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

	var paymentResponse payment.CreatePaymentResponse
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

func (h *Handler) CallbackPayment(w http.ResponseWriter, r *http.Request) {
	var midtrans order.RequestFromMidtrans
	if err := json.NewDecoder(r.Body).Decode(&midtrans); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if strings.Contains(midtrans.StatusMessage, "notification") {
		helper.HandleResponse(w, h.render, http.StatusNotFound, "not a notification path", nil)
		return
	}

	timeNow := time.Now()
	var bReq order.RequestCallback
	bReq.OrderId = midtrans.OrderID
	bReq.Status = "Payment"
	bReq.IsPaid = true
	bReq.UpdatedAt = &timeNow

	callbackChannel := make(chan client.Response)
	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: callbackUrl,
	}
	netClient.Post(bReq, callbackChannel)
	responseCallback := <-callbackChannel
	if responseCallback.Err != nil || responseCallback.StatusCode != http.StatusOK {
		helper.HandleResponse(w, h.render, responseCallback.StatusCode, responseCallback.Res, nil)
		return
	}

	var bResp string
	if err := json.Unmarshal(responseCallback.Res, &bResp); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, responseCallback.StatusCode, helper.SUCCESS_MESSSAGE, bResp)
}

func (h *Handler) CheckStatusPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)

	param := mux.Vars(r)
	orderId := param["order_id"]

	checkStatusChannel := make(chan client.Response)
	netClient := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: checkStatusUrl + usrId,
		QueryParam: []client.QueryParams{
			{Param: "order_id", Value: orderId},
		},
	}

	netClient.Get(nil, checkStatusChannel)
	responseCheckStatus := <-checkStatusChannel
	if responseCheckStatus.Err != nil || responseCheckStatus.StatusCode != http.StatusOK {
		helper.HandleResponse(w, h.render, responseCheckStatus.StatusCode, responseCheckStatus.Res, nil)
		return
	}

	var bResp order.Order
	if err := json.Unmarshal(responseCheckStatus.Res, &bResp); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, responseCheckStatus.StatusCode, helper.SUCCESS_MESSSAGE, bResp)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usrID := middleware.GetUserID(ctx)
	uid, err := uuid.Parse(usrID)
	if err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "Error parse uuid", nil)
		return
	}

	orderID := mux.Vars(r)["order_id"]
	oid, err := uuid.Parse(orderID)
	if err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "Error parse uuid", nil)
		return
	}

	var bReq order.UpdateStatus
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, err.Error(), nil)
		return
	}
	bReq.UserID = uid
	bReq.OrderID = oid

	updateChannel := make(chan client.Response)
	go client.Put(client.NetClient, updateStatusUrl, bReq, updateChannel)
	bResp := <-updateChannel
	if bResp.Err != nil || bResp.StatusCode != http.StatusOK {
		helper.HandleResponse(w, h.render, bResp.StatusCode, bResp.Err.Error(), nil)
		return
	}

	var response string
	if err := json.Unmarshal(bResp.Res, &response); err != nil {
		helper.HandleResponse(w, h.render, http.StatusConflict, "Error unmarshall response", nil)
		return
	}

	helper.HandleResponse(w, h.render, http.StatusCreated, helper.SUCCESS_MESSSAGE, response)
}

func (h *Handler) SellerUpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	role := middleware.GetRole(ctx)
	usrID := middleware.GetUserID(ctx)
	uid, err := uuid.Parse(usrID)
	if err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "Error parse uuid", nil)
		return
	}

	if role != middleware.RoleSeller {
		helper.HandleResponse(w, h.render, http.StatusForbidden, "You are not Seller", nil)
		return
	}

	var bReq order.RequestUpdateShipping
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, err.Error(), nil)
		return
	}

	bReq.UserID = uid

	netClient := client.NetClient
	updateShippingChannel := make(chan client.Response)
	client.Put(netClient, updateShppingUrl, bReq, updateShippingChannel)
	bResp := <-updateShippingChannel
	if bResp.Err != nil || bResp.StatusCode != http.StatusCreated {
		helper.HandleResponse(w, h.render, bResp.StatusCode, bResp.Err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, http.StatusCreated, helper.SUCCESS_MESSSAGE, nil)
}
