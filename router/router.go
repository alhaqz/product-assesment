package router

import (
	"net/http"

	"be-assesment-product/handler"

	"github.com/gorilla/mux"
)

func NewRouter(h *handler.ProductHandler) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/product/create", h.AddProduct).Methods(http.MethodPost)
	r.HandleFunc("/product/list", h.ListProducts).Methods(http.MethodGet)
	return r
}
