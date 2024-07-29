package users

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"user-service/src/util/helper"
	model "user-service/src/util/repository/model"
	users "user-service/src/util/repository/model/users"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/thedevsaddam/renderer"
)

func TestHandler_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDto := NewMockuserDto(ctrl)
	rend := renderer.New()
	h := NewUserHandler(
		mockUserDto,
		rend,
	)

	t.Run("successful update", func(t *testing.T) {
		usrId := uuid.New()
		user := users.Users{
			Id:                  usrId,
			Email:               "",
			Username:            "",
			Role:                "",
			Address:             "",
			CategoryPreferences: []string{},
			CreatedAt:           nil,
			UpdatedAt:           nil,
			DeletedAt:           nil,
		}

		mockUserDto.EXPECT().UpdateProfile(usrId, user).Return(nil)

		body, _ := json.Marshal(user)
		req, err := http.NewRequest("PUT", "/users/"+usrId.String(), bytes.NewBuffer(body))
		assert.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"user_id": usrId.String()})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.UpdateProfile)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Contains(t, rr.Body.String(), helper.SUCCESS_MESSSAGE)
	})

	t.Run("invalid user id", func(t *testing.T) {
		req, err := http.NewRequest("PUT", "/users/invalid-uuid", nil)
		assert.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"user_id": "invalid-uuid"})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.UpdateProfile)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid user id")
	})

	t.Run("decode error", func(t *testing.T) {
		usrId := uuid.New()

		req, err := http.NewRequest("PUT", "/users/"+usrId.String(), bytes.NewBuffer([]byte("invalid body")))
		assert.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"user_id": usrId.String()})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.UpdateProfile)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("update profile error", func(t *testing.T) {
		usrId := uuid.New()
		user := users.Users{
			Id:                  usrId,
			Email:               "",
			Username:            "",
			Role:                "",
			Address:             "",
			CategoryPreferences: []string{},
			CreatedAt:           nil,
			UpdatedAt:           nil,
			DeletedAt:           nil,
		}

		mockUserDto.EXPECT().UpdateProfile(usrId, user).Return(errors.New("update error"))

		body, _ := json.Marshal(user)
		req, err := http.NewRequest("PUT", "/users/"+usrId.String(), bytes.NewBuffer(body))
		assert.NoError(t, err)

		req = mux.SetURLVars(req, map[string]string{"user_id": usrId.String()})

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.UpdateProfile)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestHandler_GetUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDto := NewMockuserDto(ctrl)
	rend := renderer.New()
	h := NewUserHandler(
		mockUserDto,
		rend,
	)

	t.Run("successful get users", func(t *testing.T) {
		search := "test"
		role := "admin"
		userId := uuid.New()
		page := 1
		limit := 10

		req, err := http.NewRequest("GET", "/users?search="+search+"&role="+role+"&user_id="+userId.String()+"&page="+strconv.Itoa(page)+"&limit="+strconv.Itoa(limit), nil)
		assert.NoError(t, err)

		expectedResponse := &model.BaseModel{
			// isi field sesuai dengan struct BaseModel
		}

		mockUserDto.EXPECT().Get(users.RequestUsers{
			Search: search,
			Role:   role,
			UserId: userId,
			Page:   page,
			Limit:  limit,
		}).Return(expectedResponse, nil)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.GetUsers)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), helper.SUCCESS_MESSSAGE)
	})

	t.Run("invalid user id", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/users?user_id=invalid-uuid", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.GetUsers)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid user id")
	})

	t.Run("invalid page parameter", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/users?page=invalid", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.GetUsers)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "page cant not nil")
	})

	t.Run("invalid limit parameter", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/users?page=1&limit=invalid", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.GetUsers)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "limit cant not nil")
	})

	t.Run("get users error", func(t *testing.T) {
		search := "test"
		role := "admin"
		userId := uuid.New()
		page := 1
		limit := 10

		req, err := http.NewRequest("GET", "/users?search="+search+"&role="+role+"&user_id="+userId.String()+"&page="+strconv.Itoa(page)+"&limit="+strconv.Itoa(limit), nil)
		assert.NoError(t, err)

		mockUserDto.EXPECT().Get(users.RequestUsers{
			Search: search,
			Role:   role,
			UserId: userId,
			Page:   page,
			Limit:  limit,
		}).Return(nil, errors.New("get users error"))

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.GetUsers)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestHandler_SignUpByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDto := NewMockuserDto(ctrl)
	rend := renderer.New()
	h := NewUserHandler(
		mockUserDto,
		rend,
	)

	t.Run("successful sign up", func(t *testing.T) {
		user := users.Users{
			// isi field user sesuai dengan struct Users
		}

		newUUID := uuid.New()
		mockUserDto.EXPECT().Register(user).Return(&newUUID, nil)

		body, _ := json.Marshal(user)
		req, err := http.NewRequest("POST", "/signup", bytes.NewBuffer(body))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.SignUpByEmail)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), helper.SUCCESS_MESSSAGE)
	})

	t.Run("decode error", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/signup", bytes.NewBuffer([]byte("invalid body")))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.SignUpByEmail)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("register error", func(t *testing.T) {
		user := users.Users{
			// isi field user sesuai dengan struct Users
		}

		mockUserDto.EXPECT().Register(user).Return(nil, errors.New("register error"))

		body, _ := json.Marshal(user)
		req, err := http.NewRequest("POST", "/signup", bytes.NewBuffer(body))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.SignUpByEmail)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestHandler_SignInByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDto := NewMockuserDto(ctrl)
	rend := renderer.New()
	h := NewUserHandler(
		mockUserDto,
		rend,
	)

	t.Run("successful sign in", func(t *testing.T) {
		user := users.Users{
			// isi field user sesuai dengan struct Users
		}

		expectedResponse := &users.LoginResponse{
			// isi field sesuai dengan struct LoginResponse
		}

		mockUserDto.EXPECT().Login(user).Return(expectedResponse, nil)

		body, _ := json.Marshal(user)
		req, err := http.NewRequest("POST", "/signin", bytes.NewBuffer(body))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.SignInByEmail)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), helper.SUCCESS_MESSSAGE)
	})

	t.Run("decode error", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/signin", bytes.NewBuffer([]byte("invalid body")))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.SignInByEmail)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("login error", func(t *testing.T) {
		user := users.Users{
			// isi field user sesuai dengan struct Users
		}

		mockUserDto.EXPECT().Login(user).Return(nil, errors.New("login error"))

		body, _ := json.Marshal(user)
		req, err := http.NewRequest("POST", "/signin", bytes.NewBuffer(body))
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(h.SignInByEmail)

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
