package products

import "time"

type Data struct {
	Id        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Response struct {
	Data Data `json:"data"`
}

type ResponseError struct {
	Errors  map[string]any `json:"errors"`
	Message string         `json:"message"`
	Success bool           `json:"success"`
}

type RequesstCreate struct {
	Name string `json:"name"`
}

type GetProductsRequest struct {
	ShopId      string `json:"shop_id"`
	CategoryId  string `json:"category_id"`
	Name        string `json:"name"`
	PriceMinStr string `json:"price_min"`
	PriceMaxStr string `json:"price_max"`
	IsAvailable string `json:"is_available"`
	Page        string `json:"page"`
	Limit       string `json:"limit"`
}

type UpdateProductRequest struct {
	Id          string  `params:"id" validate:"required,uuid"`
	CategoryId  string  `json:"category_id" validate:"omitempty,uuid"`
	Name        string  `json:"name" validate:"required,max=255,min=3"`
	Description *string `json:"description" validate:"omitempty,max=255,min=3"`
	ImageUrl    *string `json:"image_url" validate:"omitempty,url"`
	Price       float64 `json:"price" validate:"required,numeric"`
	Stock       int64   `json:"stock" validate:"required,numeric"`
}

type UpsertProductResponse struct {
	Id          string    `json:"id" db:"id"`
	UserId      string    `json:"user_id" db:"user_id"`
	ShopId      string    `json:"shop_id" db:"shop_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	ImageUrl    *string   `json:"image_url" db:"image_url"`
	Price       float64   `json:"price" db:"price"`
	Stock       int       `json:"stock" db:"stock"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateProductRequest struct {
	ShopId      string  `json:"shop_id" validate:"required,uuid"`
	CategoryId  string  `json:"category_id" validate:"required,uuid"`
	Name        string  `json:"name" validate:"required,max=255,min=3"`
	Description *string `json:"description" validate:"omitempty,max=255,min=3"`
	ImageUrl    *string `json:"image_url" validate:"omitempty,url"`
	Price       float64 `json:"price" validate:"required,numeric"`
	Stock       int64   `json:"stock" validate:"required,numeric"`
}

type DeleteProductRequest struct {
	ProductId string `params:"product_id" validate:"required,uuid"`
	UserId    string `query:"user_id" validate:"required,uuid"`
}
