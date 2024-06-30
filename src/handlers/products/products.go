package products

import (
	"encoding/json"
	"github.com/thedevsaddam/renderer"
	"net/http"
	"user-service/src/util/client"
	"user-service/src/util/helper"
	"user-service/src/util/middleware"
	"user-service/src/util/repository/model/products"
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
		helper.HandleResponse(w, h.render, http.StatusBadRequest, bResp.Err.Error(), nil)
		return
	}

	if bResp.StatusCode != http.StatusCreated {
		helper.HandleResponse(w, h.render, bResp.StatusCode, bResp.Res, nil)
		return
	}

	var resp products.Response
	if err := json.Unmarshal(bResp.Res, &resp); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, bResp.StatusCode, "Shop created successfully", resp)
}
