package users

import (
	"encoding/json"
	"net/http"
	"strconv"
	"user-service/src/util/helper"
	"user-service/src/util/repository/model"
	"user-service/src/util/repository/model/users"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/thedevsaddam/renderer"
)

type userDto interface {
	Register(bReq users.Users) (*uuid.UUID, error)
	Get(bReq users.RequestUsers) (*model.BaseModel, error)
	UpdateProfile(id uuid.UUID, bReq users.Users) error
	Login(bReq users.Users) (*users.LoginResponse, error)
}

type Handler struct {
	render *renderer.Render
	dto    userDto
}

func NewUserHandler(dto userDto, render *renderer.Render) *Handler {
	return &Handler{
		dto:    dto,
		render: render,
	}
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	param := mux.Vars(r)
	id := param["user_id"]

	usrId, err := uuid.Parse(id)
	if err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "Invalid user id", nil)
		return
	}

	var bReq users.Users
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, err, nil)
		return
	}

	if err := h.dto.UpdateProfile(usrId, bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err, nil)
		return
	}

	helper.HandleResponse(w, h.render, http.StatusCreated, helper.SUCCESS_MESSSAGE, nil)
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query()
	search := param.Get("search")
	role := param.Get("role")
	userIdStr := param.Get("user_id")

	var userId uuid.UUID
	var userIdPtr uuid.UUID

	if userIdStr != "" {
		var err error
		userId, err = uuid.Parse(userIdStr)
		if err != nil {
			helper.HandleResponse(w, h.render, http.StatusBadRequest, "Invalid user id", nil)
			return
		}
		userIdPtr = userId
	}

	page, err := strconv.Atoi(param.Get("page"))
	if err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "page cant not nil", nil)
		return
	}
	limit, err := strconv.Atoi(param.Get("limit"))
	if err != nil {
		helper.HandleResponse(w, h.render, http.StatusBadRequest, "limit cant not nil", nil)
		return
	}

	bResp, err := h.dto.Get(users.RequestUsers{
		Search: search,
		Role:   role,
		UserId: userIdPtr,
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err, nil)
		return
	}

	helper.HandleResponse(w, h.render, http.StatusOK, helper.SUCCESS_MESSSAGE, bResp)
}

func (h *Handler) SignUpByEmail(w http.ResponseWriter, r *http.Request) {
	var bReq users.Users
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusConflict, err.Error(), nil)
		return
	}

	bResp, err := h.dto.Register(bReq)
	if err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, http.StatusOK, helper.SUCCESS_MESSSAGE, bResp)
}

func (h *Handler) SignInByEmail(w http.ResponseWriter, r *http.Request) {
	var bReq users.Users
	if err := json.NewDecoder(r.Body).Decode(&bReq); err != nil {
		helper.HandleResponse(w, h.render, http.StatusConflict, err.Error(), nil)
		return
	}

	bResp, err := h.dto.Login(bReq)
	if err != nil {
		helper.HandleResponse(w, h.render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	helper.HandleResponse(w, h.render, http.StatusOK, helper.SUCCESS_MESSSAGE, bResp)
}
