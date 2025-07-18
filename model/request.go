package model

type ListProductRequest struct {
	Page  int32  `json:"page"`
	Limit int32  `json:"limit"`
	Sort  string `json:"sort"`
	Dir   string `json:"dir"`
	Query string `json:"query"`
}

type ListProductResponse struct {
	Error      bool
	Code       int32
	Message    string
	Data       []*Product
	Pagination *PaginationResponse
}

type CreateProductRequest struct {
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Quantity    int32   `json:"quantity"`
}

type CreateProductResponse struct {
	Error   bool
	Code    int32
	Message string
}
