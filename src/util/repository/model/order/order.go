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

	// Payment
	BasicAuthHeader    string             `json:"basic_auth_header"`
	PaymentType        string             `json:"payment_type"`
	TransactionDetails TransactionDetails `json:"transaction_details"`
	BankTransfer       BankTransfer       `json:"bank_transfer"`
}

type BankTransfer struct {
	Bank string `json:"bank"`
}

type TransactionDetails struct {
	OrderID     string  `json:"order_id"`
	GrossAmount float64 `json:"gross_amount"`
}

type UpdateQtyRequest struct {
	ProductId string `json:"product_id" validate:"required"`
	Stock     int    `json:"stock" validate:"required"`
}

type ProductOrder struct {
	ProductID     string  `json:"product_id"`
	ProductName   string  `json:"product_name"`
	Price         float64 `json:"price"`
	Qty           int     `json:"qty"`
	SubtotalPrice float64 `json:"subtotal_price"`
}
