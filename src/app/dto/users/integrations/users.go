package integrations

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"user-service/src/util/helper/integrations"
	"user-service/src/util/repository/model/users"

	"github.com/coreos/go-oidc"
)

type userRepository interface {
	GetUsers(bReq users.RequestUsers) (*[]users.Users, int, error)
	GetUserDetails(bReq users.Users) (*users.Users, error)
}

type UserUsecase struct {
	user userRepository
}

func NewUserUsecase(user userRepository) *UserUsecase {
	return &UserUsecase{
		user: user,
	}
}

func (u *UserUsecase) GetUsers(bReq users.RequestUsers) (*[]users.Users, int, error) {
	result, _, err := u.user.GetUsers(bReq)
	if err != nil {
		return nil, 0, err
	}

	return result, 0, nil
}

func (u *UserUsecase) UserDataSignUp(state, code string) (*users.OauthUserData, error) {
	if state != integrations.RandomString {
		return nil, errors.New("invalid user state")
	}

	token, err := integrations.SSOSignup.Exchange(context.Background(), code)
	if err != nil {
		return nil, errors.New("cannot retrieve token")
	}

	provider, err := oidc.NewProvider(context.Background(), integrations.Provider)
	if err != nil {
		return nil, errors.New("invalid token signature")
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: integrations.SSOSignup.ClientID,
	})
	_, err = verifier.Verify(context.Background(), token.Extra("id_token").(string))
	if err != nil {
		return nil, errors.New("invalid token signature")
	}

	result, err := http.Get(integrations.UserInfoURL + token.AccessToken)
	if err != nil {
		return nil, errors.New("cannot retrieve response")
	}
	defer result.Body.Close()

	var bResp users.OauthUserData
	if err := json.NewDecoder(result.Body).Decode(&bResp); err != nil {
		return nil, err
	}

	return &bResp, nil
}

func (u *UserUsecase) UserDataSignIn(state, code string) (*users.OauthUserData, error) {
	if state != integrations.RandomString {
		return nil, errors.New("invalid user state")
	}

	token, err := integrations.SSOSignin.Exchange(context.Background(), code)
	if err != nil {
		return nil, errors.New("cannot retrieve token")
	}

	provider, err := oidc.NewProvider(context.Background(), integrations.Provider)
	if err != nil {
		return nil, errors.New("invalid token signature")
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: integrations.SSOSignin.ClientID,
	})
	_, err = verifier.Verify(context.Background(), token.Extra("id_token").(string))
	if err != nil {
		return nil, errors.New("invalid token signature")
	}

	result, err := http.Get(integrations.UserInfoURL + token.AccessToken)
	if err != nil {
		return nil, errors.New("cannot retrieve response")
	}
	defer result.Body.Close()

	var bResp users.OauthUserData
	if err := json.NewDecoder(result.Body).Decode(&bResp); err != nil {
		return nil, err
	}

	return &bResp, nil
}

func (u *UserUsecase) Login(bReq users.Users) (*users.Users, error) {
	result, err := u.user.GetUserDetails(bReq)
	if err != nil {
		return nil, err
	}

	return result, err
}
