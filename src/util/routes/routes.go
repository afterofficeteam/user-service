package routes

import (
	"log"
	"net/http"
	"time"
	"user-service/src/util/config"
	"user-service/src/util/helper"
	"user-service/src/util/middleware"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	cart "user-service/src/handlers/cart"
	order "user-service/src/handlers/order"
	product "user-service/src/handlers/products"
	shop "user-service/src/handlers/shop"
	user "user-service/src/handlers/users"
	integration "user-service/src/handlers/users/integrations"
)

type Routes struct {
	Router      *mux.Router
	Integration *integration.Handler
	User        *user.Handler
	Product     *product.Handler
	Shop        *shop.Handler
	Cart        *cart.Handler
	Order       *order.Handler
}

func (r *Routes) Run(port string) {
	r.SetupRouter()

	log.Printf("[HTTP SRV] clients on localhost port :%s", port)
	srv := &http.Server{
		Handler:      r.Router,
		Addr:         "localhost:" + port,
		WriteTimeout: config.WriteTimeout() * time.Second,
		ReadTimeout:  config.ReadTimeout() * time.Second,
	}

	log.Panic(srv.ListenAndServe())
}

func (r *Routes) SetupRouter() {
	r.Router = mux.NewRouter()
	r.Router.Use(helper.EnabledCors, helper.LoggerMiddleware())

	r.SetupBaseURL()
	r.SetupIntegration()
	r.SetupUser()
	r.SetupProduct()
	r.SetupShop()
	r.SetupCart()
	r.setupOrder()
}

func (r *Routes) SetupBaseURL() {
	baseURL := viper.GetString("BASE_URL_PATH")
	if baseURL != "" && baseURL != "/" {
		r.Router.PathPrefix(baseURL).HandlerFunc(helper.URLRewriter(r.Router, baseURL))
	}
}

func (r *Routes) SetupIntegration() {
	path := r.Router.PathPrefix("/users").Subrouter()
	path.HandleFunc("/signup", r.Integration.SignUp).Methods(http.MethodGet, http.MethodOptions)
	path.HandleFunc("/signup/callback", r.Integration.RedirectSignUp).Methods(http.MethodGet, http.MethodOptions)
	path.HandleFunc("/signin", r.Integration.SignIn).Methods(http.MethodGet, http.MethodOptions)
	path.HandleFunc("/signin/callback", r.Integration.RedirectSignIn).Methods(http.MethodGet, http.MethodOptions)
}

func (r *Routes) SetupUser() {
	userRoutes := r.Router.PathPrefix("/users").Subrouter()
	userRoutes.HandleFunc("/signup/email", r.User.SignUpByEmail).Methods(http.MethodPost, http.MethodOptions)
	userRoutes.HandleFunc("/signin/email", r.User.SignInByEmail).Methods(http.MethodPost, http.MethodOptions)

	authenticatedRoutes := userRoutes.PathPrefix("").Subrouter()
	authenticatedRoutes.Use(middleware.Authentication)
	authenticatedRoutes.HandleFunc("", r.User.GetUsers).Methods(http.MethodGet, http.MethodOptions)
	authenticatedRoutes.HandleFunc("/{user_id}/update", r.User.UpdateProfile).Methods(http.MethodPut, http.MethodOptions)
}

func (r *Routes) SetupProduct() {
	productRoutes := r.Router.PathPrefix("/products").Subrouter()
	productRoutes.HandleFunc("", r.Product.GetProducts).Methods(http.MethodGet, http.MethodOptions)

	authenticatedProductRoutes := productRoutes.PathPrefix("").Subrouter()
	authenticatedProductRoutes.Use(middleware.Authentication)
	authenticatedProductRoutes.HandleFunc("", r.Product.GetProducts).Methods(http.MethodGet, http.MethodOptions)
	authenticatedProductRoutes.HandleFunc("/{product_id}", r.Product.UpdateProduct).Methods(http.MethodPut, http.MethodOptions)
	authenticatedProductRoutes.HandleFunc("/create", r.Product.CreateProduct).Methods(http.MethodPost, http.MethodOptions)
	authenticatedProductRoutes.HandleFunc("/{product_id}/delete", r.Product.DeleteProduct).Methods(http.MethodDelete, http.MethodOptions)
}

func (r *Routes) SetupShop() {
	shopRoutes := r.Router.PathPrefix("/shops").Subrouter()
	shopRoutes.HandleFunc("", r.Shop.GetShops).Methods(http.MethodGet, http.MethodOptions)

	authenticatedShopRoutes := shopRoutes.PathPrefix("").Subrouter()
	authenticatedShopRoutes.Use(middleware.Authentication)
	authenticatedShopRoutes.HandleFunc("/create", r.Shop.CreateShop).Methods(http.MethodPost, http.MethodOptions)
}

func (r *Routes) SetupCart() {
	cartRoutes := r.Router.PathPrefix("/cart").Subrouter()
	cartRoutes.Use(middleware.Authentication)
	cartRoutes.HandleFunc("/details", r.Cart.GetCartByUserID).Methods(http.MethodGet, http.MethodOptions)
	cartRoutes.HandleFunc("/update", r.Cart.UpdateCart).Methods(http.MethodPut, http.MethodOptions)
	cartRoutes.HandleFunc("/add", r.Cart.AddCart).Methods(http.MethodPost, http.MethodOptions)
	cartRoutes.HandleFunc("/delete", r.Cart.DeleteCart).Methods(http.MethodDelete, http.MethodOptions)
}

func (r *Routes) setupOrder() {
	orderRoutes := r.Router.PathPrefix("/order").Subrouter()
	orderRoutes.Use(middleware.Authentication)
	orderRoutes.HandleFunc("/create", r.Order.CreateOrder).Methods(http.MethodPost, http.MethodOptions)
	orderRoutes.HandleFunc("/status/{order_id}", r.Order.CheckStatusPayment).Methods(http.MethodGet, http.MethodOptions)
	orderRoutes.HandleFunc("/status/{order_id}/update", r.Order.UpdateStatus).Methods(http.MethodPut, http.MethodOptions)
	orderRoutes.HandleFunc("/status/{order_id}/shipping/update", r.Order.SellerUpdateStatus).Methods(http.MethodPut, http.MethodOptions)

	callbackRoutes := r.Router.PathPrefix("/order/callback").Subrouter()
	callbackRoutes.HandleFunc("", r.Order.CallbackPayment).Methods(http.MethodPost, http.MethodOptions)
}
