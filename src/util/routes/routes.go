package routes

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
	"user-service/src/util/config"
	"user-service/src/util/middleware"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	cart "user-service/src/handlers/cart"
	order "user-service/src/handlers/order"
	product "user-service/src/handlers/products"
	user "user-service/src/handlers/users"
	integration "user-service/src/handlers/users/integrations"
)

type Routes struct {
	Router      *mux.Router
	Integration *integration.Handler
	User        *user.Handler
	Product     *product.Handler
	Cart        *cart.Handler
	Order       *order.Handler
}

func EnabledCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", "DELETE, POST, GET, OPTIONS, PUT, PATCH")
		header.Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func URLRewriter(router *mux.Router, baseURLPath string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = func(url string) string {
			if strings.Index(url, baseURLPath) == 0 {
				url = url[len(baseURLPath):]
			}
			return url
		}(r.URL.Path)

		router.ServeHTTP(w, r)
	}
}

func LoggerMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "notifications") {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			recorder := httptest.NewRecorder()
			next.ServeHTTP(recorder, r)

			for k, v := range recorder.Header() {
				w.Header()[k] = v
			}
			w.WriteHeader(recorder.Code)
			recorder.Body.WriteTo(w)

			responseTime := time.Since(start).Seconds()
			formattedResponseTime := fmt.Sprintf("%.9f", responseTime)
			formattedResponseTime = fmt.Sprintf("%sµs", formattedResponseTime)

			log.Printf("%s - [%s] - [%s] \"%s %s %s\" %d %s\n",
				r.RemoteAddr,
				time.Now().Format(time.RFC1123),
				formattedResponseTime,
				r.Method,
				r.URL.Path,
				r.Proto,
				recorder.Code,
				r.UserAgent(),
			)
		})
	}
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
	r.Router.Use(EnabledCors, LoggerMiddleware())

	r.SetupBaseURL()
	r.SetupIntegration()
	r.SetupUser()
	r.SetupProduct()
	r.SetupCart()
	r.setupOrder()
}

func (r *Routes) SetupBaseURL() {
	baseURL := viper.GetString("BASE_URL_PATH")
	if baseURL != "" && baseURL != "/" {
		r.Router.PathPrefix(baseURL).HandlerFunc(URLRewriter(r.Router, baseURL))
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
	productRoutes.Use(middleware.Authentication)
	productRoutes.HandleFunc("/create", r.Product.CreateShop).Methods(http.MethodPost, http.MethodOptions)
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
	orderRoutes.HandleFunc("/create/items", r.Order.CreateOrderItems).Methods(http.MethodPost, http.MethodOptions)
	orderRoutes.HandleFunc("/create/items/logs", r.Order.CreateOrderItemlogs).Methods(http.MethodPost, http.MethodOptions)
}
