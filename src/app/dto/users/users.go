package users

import (
	"errors"
	"math"
	"time"
	"user-service/src/util/helper/jwt"
	"user-service/src/util/repository/model"
	"user-service/src/util/repository/model/users"

	"github.com/google/uuid"
)

type userRepository interface {
	RegisterUser(bReq users.Users) (*uuid.UUID, error)
	GetUserDetails(bReq users.Users) (*users.Users, error)
	GetUsers(bReq users.RequestUsers) (*[]users.Users, int, error)
	UpdateUser(id uuid.UUID, bReq users.Users) error
}

type UserUsecase struct {
	user userRepository
}

func NewUserUsecase(user userRepository) *UserUsecase {
	return &UserUsecase{
		user: user,
	}
}

func (u *UserUsecase) UpdateProfile(id uuid.UUID, bReq users.Users) error {
	if err := u.user.UpdateUser(id, bReq); err != nil {
		return err
	}

	return nil
}

func (u *UserUsecase) Register(bReq users.Users) (*uuid.UUID, error) {
	usrInfo, err := u.user.GetUserDetails(bReq)
	if err != nil {
		return nil, err
	}

	if usrInfo.Email == bReq.Email && usrInfo.Username == bReq.Username {
		return nil, errors.New("user already registered")
	}

	result, err := u.user.RegisterUser(bReq)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (u *UserUsecase) Login(bReq users.Users) (*users.LoginResponse, error) {
	usrLogin, err := u.user.GetUserDetails(bReq)
	if err != nil {
		return nil, err
	}

	if usrLogin.Email != bReq.Email {
		return nil, errors.New("users not yet registered")
	}

	tokenExpiry := time.Minute * 20
	accessToken, payload, err := jwt.CreateAccessToken(usrLogin.Email, usrLogin.Id.String(), tokenExpiry)
	if err != nil {
		return nil, err
	}

	refreshTokenExpiry := time.Hour * 72
	refreshToken, refreshTokenPayload, err := jwt.CreateRefreshToken(usrLogin.Email, usrLogin.Id.String(), refreshTokenExpiry)
	if err != nil {
		return nil, err
	}

	bResp := users.LoginResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: payload.ExpiresAt.Time,
		RefreshToken:         refreshToken,
		RefreshTokenExpiryAt: refreshTokenPayload.ExpiresAt.Time,
		Users:                usrLogin,
	}

	return &bResp, nil
}

func (u *UserUsecase) Get(bReq users.RequestUsers) (*model.BaseModel, error) {
	result, totalData, err := u.user.GetUsers(bReq)
	if err != nil {
		return nil, err
	}

	filteredPage := int(math.Ceil(float64(totalData) / float64(bReq.Limit)))
	bResp := model.BaseModel{
		Items:        result,
		TotalItem:    totalData,
		TotalPage:    filteredPage,
		FilteredItem: bReq.Limit,
		FilteredPage: filteredPage,
	}

	if len(*result) == 0 {
		bResp.Items = []string{}
	}

	return &bResp, nil
}
