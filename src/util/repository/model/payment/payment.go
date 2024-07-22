package payment

type CreatePaymentResponse struct {
	StatusCode    string     `json:"status_code"`
	StatusMessage string     `json:"status_message"`
	TxId          string     `json:"transaction_id"`
	OrderId       string     `json:"order_id"`
	MerchantId    string     `json:"merchant_id"`
	GrossAmount   string     `json:"gross_amount"`
	Currency      string     `json:"currency"`
	PaymentType   string     `json:"payment_type"`
	TxTime        string     `json:"transaction_time"`
	TxStatus      string     `json:"transaction_status"`
	VANumbers     []VANumber `json:"va_numbers"`
	FraudStatus   string     `json:"fraud_status"`
}

type VANumber struct {
	Bank     string `json:"bank"`
	VANumber string `json:"va_number"`
}
