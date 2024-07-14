package order

import (
	"time"

	"github.com/google/uuid"
)

type CreateOrderRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`

	// Product Get
	ProductIds string `json:"product_ids"`
	Limit      string `json:"limit"`

	// Product Update
	UpdateQty []UpdateQtyRequest `json:"update_qty"`

	// User
	PaymentTypeID uuid.UUID      `json:"payment_type_id" validate:"required"`
	OrderNumber   string         `json:"order_number" validate:"required"`
	TotalPrice    float64        `json:"total_price" validate:"required"`
	ProductOrder  []ProductOrder `json:"product_order"`
	Status        string         `json:"status" validate:"required"`
	IsPaid        bool           `json:"is_paid"`
	RefCode       string         `json:"ref_code"`
	CreatedAt     *time.Time     `json:"created_at"`
}

type UpdateQtyRequest struct {
	ProductId string `json:"product_id" validate:"required"`
	Stock     int    `json:"stock" validate:"required"`
}

type Order struct {
	UserID        uuid.UUID   `json:"user_id" validate:"required"`
	PaymentTypeID uuid.UUID   `json:"payment_type_id" validate:"required"`
	OrderNumber   string      `json:"order_number" validate:"required"`
	SubtotalPrice float64     `json:"subtotal_price" validate:"required"`
	TotalPrice    float64     `json:"total_price" validate:"required"`
	ProductOrder  interface{} `json:"product_order"`
	Status        string      `json:"status" validate:"required"`
	IsPaid        bool        `json:"is_paid"`
	RefCode       string      `json:"ref_code"`
	CreatedAt     *time.Time  `json:"created_at"`
	UpdatedAt     *time.Time  `json:"updated_at"`
	DeleredAt     *time.Time  `json:"deleted_at"`
}

type ProductOrder struct {
	ProductID     string  `json:"product_id"`
	ProductName   string  `json:"product_name"`
	Price         float64 `json:"price"`
	Qty           int     `json:"qty"`
	SubtotalPrice float64 `json:"subtotal_price"`
}

type OrderItemsLogs struct {
	OrderID    uuid.UUID  `json:"order_id"`
	RefCode    string     `json:"ref_code"`
	FromStatus string     `json:"from_status"`
	ToStatus   string     `json:"to_status"`
	Notes      string     `json:"notes"`
	CreatedAt  *time.Time `json:"created_at"`
}
