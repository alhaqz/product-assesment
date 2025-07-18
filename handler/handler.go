package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"be-assesment-product/model"
	"be-assesment-product/service"
)

type ProductHandler struct {
	svc service.ProductService
}

func NewProductHandler(svc service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (h *ProductHandler) AddProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result := &model.CreateProductResponse{
		Error:   false,
		Code:    http.StatusOK,
		Message: "Success Create Product",
	}

	var p model.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		result.Error = true
		result.Code = http.StatusInternalServerError
		result.Message = "Invalid Argument"
		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(result)
		return
	}

	req := &model.CreateProductRequest{
		Name:        p.Name,
		Price:       p.Price,
		Description: p.Description,
		Quantity:    p.Quantity,
	}
	if err := h.svc.AddProduct(ctx, req); err != nil {
		result.Error = true
		result.Code = http.StatusForbidden
		result.Message = err.Error()
		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(result)
		return
	}

	result.Error = false
	result.Code = http.StatusOK
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result := &model.ListProductResponse{
		Error:   false,
		Code:    http.StatusOK,
		Message: "Success",
	}

	queryParams := r.URL.Query()

	page, _ := strconv.Atoi(queryParams.Get("page"))
	limit, _ := strconv.Atoi(queryParams.Get("limit"))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	req := &model.ListProductRequest{
		Page:  int32(page),
		Limit: int32(limit),
		Sort:  queryParams.Get("sort"),
		Dir:   queryParams.Get("dir"),
		Query: queryParams.Get("query"),
	}

	products, err := h.svc.ListProducts(ctx, req)

	if err != nil {
		result.Error = true
		result.Code = http.StatusInternalServerError
		result.Message = err.Error()
		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(result)
		return
	}

	result.Error = false
	result.Code = http.StatusOK
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}
