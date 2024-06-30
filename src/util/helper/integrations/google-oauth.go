package integrations

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	UserInfoURL          = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	Provider             = "https://accounts.google.com"
	ScopesURLUserInfo    = "https://www.googleapis.com/auth/userinfo.email"
	ScopesURLUserProfile = "https://www.googleapis.com/auth/userinfo.profile"
	RandomString         = "123qwerty"
)

var (
	SSOSignup *oauth2.Config
	SSOSignin *oauth2.Config

	ClientIDSignup     = "1046398925079-55m6p5ivhnc6961d61ln8vjctjtnk9t0.apps.googleusercontent.com"
	ClientSecretSignup = "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"
	RedirectURLSignup  = "http://localhost:9990/users/signup/callback"

	ClientIDSignin     = "1046398925079-as4gpoto2dp978a4akav5ak0v5pfccg8.apps.googleusercontent.com"
	ClientSecretSignin = "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"
	RedirectURLSignin  = "http://localhost:9990/users/signin/callback"
)

func init() {
	SSOSignup = initOAuthConfig(ClientIDSignup, ClientSecretSignup, RedirectURLSignup)
	SSOSignin = initOAuthConfig(ClientIDSignin, ClientSecretSignin, RedirectURLSignin)
}

func initOAuthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			ScopesURLUserInfo,
			ScopesURLUserProfile,
		},
		Endpoint: google.Endpoint,
	}
}
