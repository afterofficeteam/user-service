package order

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
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
	clientKey string
	serverKey string
}

const (
	createOrderUrl   = "http://localhost:9993/order/create"
	getproductUrl    = "http://localhost:3000/api/products"
	updateProductUrl = "http://localhost:3000/api/product-stocks"
	paymentUrl       = "http://localhost:4000/api/payments"
)

func NewHandler(r *renderer.Render, validator *validator.Validate, clientKey, serverKey string) *Handler {
	return &Handler{render: r, validator: validator, clientKey: clientKey, serverKey: serverKey}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usrId := middleware.GetUserID(ctx)
	uid := uuid.MustParse(usrId)

	param := r.URL.Query()

	// Membuat request untuk mendapatkan produk di layanan Product
	request := order.CreateOrderRequest{
		Limit: param.Get("limit"),
	}

	var bReq order.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, err.Error(), nil)
		return
	}
	bReq.UserID = uid

	var productIds []string
	for _, product := range bReq.ProductOrder {
		productIds = append(productIds, product.ProductID)
	}
	productId := strings.Join(productIds, ",")

	productChannel := make(chan client.Response)
	netClientProducts := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: getproductUrl,
		QueryParam: []client.QueryParams{
			{Param: "product_ids", Value: productId},
			{Param: "limit", Value: request.Limit},
		},
	}

	netClientProducts.Get(nil, productChannel)
	respProduct := <-productChannel
	if respProduct.Err != nil {
		if err := json.Unmarshal(respProduct.Res, &responseError); err != nil {
			helper.HandleResponse(w, h.render, http.StatusConflict, "Error unmarshall", nil)
			return
		}

		helper.HandleResponse(w, h.render, respProduct.StatusCode, responseError["message"], nil)
		return
	}

	if respProduct.StatusCode != http.StatusOK {
		if err := json.Unmarshal(respProduct.Res, &responseError); err != nil {
			helper.HandleResponse(w, h.render, http.StatusConflict, "Error unmarshall", nil)
			return
		}

		helper.HandleResponse(w, h.render, respProduct.StatusCode, responseError["message"], nil)
		return
	}

	var productData products.DataProduct
	if err := json.Unmarshal(respProduct.Res, &productData); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), productData)
		return
	}

	// Jika jumlah produk tidak tersedia, kembalikan error
	for _, prod := range productData.Data.Items {
		for _, orderProd := range bReq.ProductOrder {
			if prod.Id == orderProd.ProductID && prod.Stock < orderProd.Qty {
				helper.HandleResponse(w, h.render, http.StatusBadRequest, "Product out of stock", nil)
				return
			}
		}
	}

	// Menghitung subtotal harga
	for i, orderProd := range bReq.ProductOrder {
		for _, prod := range productData.Data.Items {
			if prod.Id == orderProd.ProductID {
				bReq.ProductOrder[i].Price = prod.Price
				bReq.ProductOrder[i].SubtotalPrice = prod.Price * float64(orderProd.Qty)
			}
		}
	}

	// Menghitung total harga
	var total float64
	for _, orderProd := range bReq.ProductOrder {
		total += orderProd.SubtotalPrice
	}
	bReq.TotalPrice = total

	// Membuat order di layanan Order
	netClientOrder := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: createOrderUrl,
	}
	createChan := make(chan client.Response)
	go netClientOrder.Post(bReq, createChan)
	responseOrder := <-createChan
	if responseOrder.Err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, responseOrder.Err.Error(), nil)
		return
	}

	if responseOrder.StatusCode != http.StatusCreated {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, responseOrder.Res, nil)
		return
	}

	var orderID string
	if err := json.Unmarshal(responseOrder.Res, &orderID); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Mengupdate jumlah stok produk
	for _, orderProd := range bReq.ProductOrder {
		for _, prod := range productData.Data.Items {
			if prod.Id == orderProd.ProductID {
				var requestUpdate order.UpdateQtyRequest
				requestUpdate.Stock = prod.Stock - orderProd.Qty
				requestUpdate.ProductId = prod.Id
				bReq.UpdateQty = append(bReq.UpdateQty, requestUpdate)
			}
		}
	}

	// Mengupdate stok di layanan Product
	netclientUpdateStock := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: updateProductUrl,
	}
	updateStockChan := make(chan client.Response)
	go netclientUpdateStock.Patch(bReq.UpdateQty, updateStockChan)
	responseUpdateStock := <-updateStockChan

	var responseUpdate string
	if responseUpdateStock.Err != nil {
		if err := json.Unmarshal(respProduct.Res, &responseUpdate); err != nil {
			helper.HandleResponse(w, h.render, http.StatusConflict, "Error unmarshall", nil)
			return
		}

		helper.HandleResponse(w, h.render, respProduct.StatusCode, responseUpdate, nil)
		return
	}

	if responseUpdateStock.StatusCode != http.StatusOK {
		if err := json.Unmarshal(respProduct.Res, &responseUpdate); err != nil {
			helper.HandleResponse(w, h.render, http.StatusConflict, "Error unmarshall", nil)
			return
		}

		helper.HandleResponse(w, h.render, respProduct.StatusCode, responseUpdate, nil)
		return
	}

	// Payment service
	auth := base64.StdEncoding.EncodeToString([]byte(h.serverKey + ":"))
	bReq.BasicAuthHeader = "Basic " + auth
	bReq.PaymentType = "bank_transfer"
	bReq.TransactionDetails.OrderID = orderID
	bReq.TransactionDetails.GrossAmount = bReq.TotalPrice

	netclientPayment := client.NetClientRequest{
		NetClient:  client.NetClient,
		RequestUrl: paymentUrl,
	}
	paymentChan := make(chan client.Response)
	go netclientPayment.Post(bReq, paymentChan)
	responsePayment := <-paymentChan
	var paymentResponse map[string]interface{}
	if responsePayment.Err != nil {
		if err := json.Unmarshal(responsePayment.Res, &paymentResponse); err != nil {
			helper.HandleResponse(w, h.render, http.StatusBadRequest, "Error unmarshall", nil)
			return
		}

		helper.HandleResponse(w, h.render, http.StatusBadRequest, paymentResponse, nil)
		return
	}

	if responsePayment.StatusCode != http.StatusCreated {
		if err := json.Unmarshal(responsePayment.Res, &paymentResponse); err != nil {
			helper.HandleResponse(w, h.render, http.StatusBadRequest, "Error unmarshall", nil)
			return
		}

		helper.HandleResponse(w, h.render, http.StatusInternalServerError, paymentResponse, nil)
		return
	}

	if err := json.Unmarshal(responsePayment.Res, &paymentResponse); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, http.StatusOK, helper.SUCCESS_MESSSAGE, paymentResponse)
}
