package main

import (
	"database/sql"
	"sync"
	"user-service/src/handlers/cart"
	"user-service/src/handlers/order"
	"user-service/src/util/config"
	"user-service/src/util/routes"

	"github.com/go-playground/validator/v10"
	"github.com/thedevsaddam/renderer"

	userUsecase "user-service/src/app/dto/users"
	userHandler "user-service/src/handlers/users"
	userStore "user-service/src/util/repository/users"

	productHandler "user-service/src/handlers/products"

	shopHandler "user-service/src/handlers/shop"

	integrationUseCase "user-service/src/app/dto/users/integrations"
	integrationHandler "user-service/src/handlers/users/integrations"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		return
	}

	sqlDb, err := config.ConnectToDatabase(config.Connection{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	})
	if err != nil {
		return
	}
	defer sqlDb.Close()

	mutex := &sync.Mutex{}
	validator := validator.New()
	render := renderer.New()
	routes := setupRoutes(render, sqlDb, validator, cfg, mutex)
	routes.Run(cfg.AppPort)
}

func setupRoutes(render *renderer.Render, myDb *sql.DB, validator *validator.Validate, config *config.Config, mutex *sync.Mutex) *routes.Routes {
	userStore := userStore.NewStore(myDb)
	userUsecase := userUsecase.NewUserUsecase(userStore)
	userHandler := userHandler.NewUserHandler(userUsecase, render)

	integrationUseCase := integrationUseCase.NewUserUsecase(userStore)
	integrationHandler := integrationHandler.NewHandler(render, userUsecase, integrationUseCase)

	productHandler := productHandler.NewHandler(render)

	shopHandler := shopHandler.NewHandler(render)

	cartHandler := cart.NewHandler(render)

	orderHandler := order.NewHandler(render, validator, mutex, config.ClientKey, config.ServerKey)

	return &routes.Routes{
		Integration: integrationHandler,
		User:        userHandler,
		Product:     productHandler,
		Shop:        shopHandler,
		Cart:        cartHandler,
		Order:       orderHandler,
	}
}
