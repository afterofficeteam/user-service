package integrations

import (
	"net/http"
	"strings"
	"time"
	"user-service/src/util/helper"
	"user-service/src/util/helper/integrations"
	"user-service/src/util/helper/jwt"
	"user-service/src/util/repository/model/users"

	"github.com/google/uuid"
	"github.com/thedevsaddam/renderer"
)

type userDto interface {
	Register(bReq users.Users) (*uuid.UUID, error)
}

type userDtoIntegration interface {
	GetUsers(bReq users.RequestUsers) (*[]users.Users, int, error)
	Login(bReq users.Users) (*users.Users, error)
	UserDataSignUp(state, code string) (*users.OauthUserData, error)
	UserDataSignIn(state, code string) (*users.OauthUserData, error)
}

type Handler struct {
	render      *renderer.Render
	dto         userDto
	integration userDtoIntegration
}

func NewHandler(render *renderer.Render, dto userDto, integration userDtoIntegration) *Handler {
	return &Handler{
		render:      render,
		dto:         dto,
		integration: integration,
	}
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, integrations.SSOSignup.AuthCodeURL(integrations.RandomString), http.StatusTemporaryRedirect)
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, integrations.SSOSignin.AuthCodeURL(integrations.RandomString), http.StatusTemporaryRedirect)
}

func (h *Handler) RedirectSignUp(w http.ResponseWriter, r *http.Request) {
	handleOAuthCallback(w, r, h.render, h.dto, h.integration, h.integration.UserDataSignUp, true)
}

func (h *Handler) RedirectSignIn(w http.ResponseWriter, r *http.Request) {
	handleOAuthCallback(w, r, h.render, h.dto, h.integration, h.integration.UserDataSignIn, false)
}

func handleOAuthCallback(w http.ResponseWriter, r *http.Request, render *renderer.Render, dto userDto, integration userDtoIntegration, userDataFunc func(state, code string) (*users.OauthUserData, error), register bool) {
	state, code := r.FormValue("state"), r.FormValue("code")
	if state == "" || code == "" {
		helper.HandleResponse(w, render, http.StatusConflict, "state or code is nil", nil)
		return
	}

	userData, err := userDataFunc(state, code)
	if err != nil {
		helper.HandleResponse(w, render, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if register {
		// Check user already registered
		checkUser, _, err := integration.GetUsers(users.RequestUsers{
			Email: userData.Email,
			Page:  1,
			Limit: 1,
		})
		if err != nil {
			helper.HandleResponse(w, render, http.StatusInternalServerError, err, nil)
			return
		}

		if len(*checkUser) > 0 {
			helper.HandleResponse(w, render, http.StatusConflict, "User already registered", nil)
			return
		}

		// Register user
		userName := strings.ReplaceAll(strings.ToLower(userData.GivenName), " ", "")
		bResp, err := dto.Register(users.Users{
			Email:    userData.Email,
			Username: userName,
			Role:     "Admin",
			CategoryPreferences: []string{
				"Baju",
				"Buku",
			},
			Address: "Jakarta",
		})
		if err != nil {
			helper.HandleResponse(w, render, http.StatusInternalServerError, err, nil)
			return
		}

		helper.HandleResponse(w, render, http.StatusOK, helper.SUCCESS_MESSSAGE, bResp)
	} else {
		checkUser, _, err := integration.GetUsers(users.RequestUsers{
			Email: userData.Email,
			Page:  1,
			Limit: 1,
		})
		if err != nil {
			helper.HandleResponse(w, render, http.StatusInternalServerError, err, nil)
			return
		}

		if len(*checkUser) == 0 {
			helper.HandleResponse(w, render, http.StatusConflict, "User not yet registered", nil)
			return
		}

		usrLogin, err := integration.Login(users.Users{
			Email: userData.Email,
		})
		if err != nil {
			helper.HandleResponse(w, render, http.StatusInternalServerError, err, nil)
			return
		}

		tokenExpiry := time.Minute * 20
		accessToken, payload, err := jwt.CreateAccessToken(usrLogin.Email, usrLogin.Id.String(), tokenExpiry)
		if err != nil {
			return
		}

		refreshTokenExpiry := time.Hour * 72
		refreshToken, refreshTokenPayload, err := jwt.CreateRefreshToken(usrLogin.Email, usrLogin.Id.String(), refreshTokenExpiry)
		if err != nil {
			return
		}

		bResp := users.LoginResponse{
			AccessToken:          accessToken,
			AccessTokenExpiresAt: payload.ExpiresAt.Time,
			RefreshToken:         refreshToken,
			RefreshTokenExpiryAt: refreshTokenPayload.ExpiresAt.Time,
			Users:                usrLogin,
		}

		helper.HandleResponse(w, render, http.StatusOK, helper.SUCCESS_MESSSAGE, bResp)
	}
}
