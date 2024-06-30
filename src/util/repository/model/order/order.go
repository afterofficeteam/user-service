package order

import (
	"github.com/google/uuid"
	"time"
)

type Order struct {
	UserID        uuid.UUID  `json:"user_id"`
	PaymentTypeID uuid.UUID  `json:"payment_type_id"`
	OrderNumber   string     `json:"order_number"`
	TotalPrice    float64    `json:"total_price"`
	Status        string     `json:"status"`
	IsPaid        bool       `json:"is_paid"`
	DiscountID    uuid.UUID  `json:"discount_id"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
	DeleredAt     *time.Time `json:"deleted_at"`
}

type OrderItems struct {
	OrderID       uuid.UUID  `json:"order_id"`
	ProductID     uuid.UUID  `json:"product_id"`
	Qty           int        `json:"qty"`
	Price         float64    `json:"price"`
	ProductName   string     `json:"product_name"`
	SubtotalPrice float64    `json:"subtotal_price"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
}

type OrderItemsLogs struct {
	OrderID    uuid.UUID  `json:"order_id"`
	FromStatus string     `json:"from_status"`
	ToStatus   string     `json:"to_status"`
	Notes      string     `json:"notes"`
	CreatedAt  *time.Time `json:"created_at"`
}
