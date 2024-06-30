package products

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
