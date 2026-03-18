package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

// ProductHandler handles CRUD for the product catalog and price books.
type ProductHandler struct {
	products repository.ProductRepository
}

func NewProductHandler(products repository.ProductRepository) *ProductHandler {
	return &ProductHandler{products: products}
}

func (h *ProductHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListProducts)
	r.Get("/{id}", h.GetProduct)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin))
		r.Post("/", h.CreateProduct)
		r.Patch("/{id}", h.UpdateProduct)
		r.Delete("/{id}", h.DeactivateProduct)
	})
	return r
}

func (h *ProductHandler) PriceBooksRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListPriceBooks)
	r.Get("/{id}", h.GetPriceBook)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(domain.UserRoleAdmin))
		r.Post("/", h.CreatePriceBook)
		r.Put("/{id}/entries/{productId}", h.UpsertPriceBookEntry)
		r.Delete("/{id}/entries/{productId}", h.DeletePriceBookEntry)
	})
	return r
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	activeOnly := true
	if v := r.URL.Query().Get("active_only"); v == "false" {
		activeOnly = false
	}
	products, err := h.products.ListProducts(r.Context(), activeOnly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if products == nil {
		products = []*domain.Product{}
	}
	writeJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	p, err := h.products.GetProduct(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p domain.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	created, err := h.products.CreateProduct(r.Context(), &p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var patch domain.ProductPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	updated, err := h.products.UpdateProduct(r.Context(), id, patch)
	if err != nil {
		if err == domain.ErrNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *ProductHandler) DeactivateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.products.DeactivateProduct(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) ListPriceBooks(w http.ResponseWriter, r *http.Request) {
	pbs, err := h.products.ListPriceBooks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pbs == nil {
		pbs = []*domain.PriceBook{}
	}
	writeJSON(w, http.StatusOK, pbs)
}

func (h *ProductHandler) GetPriceBook(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	pb, err := h.products.GetPriceBook(r.Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, pb)
}

func (h *ProductHandler) CreatePriceBook(w http.ResponseWriter, r *http.Request) {
	var pb domain.PriceBook
	if err := json.NewDecoder(r.Body).Decode(&pb); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	created, err := h.products.CreatePriceBook(r.Context(), &pb)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *ProductHandler) UpsertPriceBookEntry(w http.ResponseWriter, r *http.Request) {
	pbID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid price book id", http.StatusBadRequest)
		return
	}
	productID, err := uuid.Parse(chi.URLParam(r, "productId"))
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}
	var body struct {
		PriceOverride *float64 `json:"price_override"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	entry := &domain.PriceBookEntry{
		PriceBookID:   pbID,
		ProductID:     productID,
		PriceOverride: body.PriceOverride,
	}
	updated, err := h.products.UpsertPriceBookEntry(r.Context(), entry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *ProductHandler) DeletePriceBookEntry(w http.ResponseWriter, r *http.Request) {
	pbID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid price book id", http.StatusBadRequest)
		return
	}
	productID, err := uuid.Parse(chi.URLParam(r, "productId"))
	if err != nil {
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}
	if err := h.products.DeletePriceBookEntry(r.Context(), pbID, productID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DealLineItemHandler handles line items on a deal.
type DealLineItemHandler struct {
	items repository.DealLineItemRepository
}

func NewDealLineItemHandler(items repository.DealLineItemRepository) *DealLineItemHandler {
	return &DealLineItemHandler{items: items}
}

func (h *DealLineItemHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Upsert)
	r.Put("/{itemId}", h.Upsert)
	r.Delete("/{itemId}", h.Delete)
	r.Post("/reorder", h.Reorder)
	return r
}

func (h *DealLineItemHandler) List(w http.ResponseWriter, r *http.Request) {
	dealID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid deal id", http.StatusBadRequest)
		return
	}
	items, err := h.items.List(r.Context(), dealID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if items == nil {
		items = []*domain.DealLineItem{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *DealLineItemHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	dealID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid deal id", http.StatusBadRequest)
		return
	}
	var item domain.DealLineItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	item.DealID = dealID
	if idStr := chi.URLParam(r, "itemId"); idStr != "" {
		if id, err := uuid.Parse(idStr); err == nil {
			item.ID = id
		}
	}
	saved, err := h.items.Upsert(r.Context(), &item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	status := http.StatusCreated
	if chi.URLParam(r, "itemId") != "" {
		status = http.StatusOK
	}
	writeJSON(w, status, saved)
}

func (h *DealLineItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	dealID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid deal id", http.StatusBadRequest)
		return
	}
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		http.Error(w, "invalid item id", http.StatusBadRequest)
		return
	}
	if err := h.items.Delete(r.Context(), itemID, dealID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DealLineItemHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	dealID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid deal id", http.StatusBadRequest)
		return
	}
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ids := make([]uuid.UUID, 0, len(body.IDs))
	for _, s := range body.IDs {
		id, err := uuid.Parse(s)
		if err != nil {
			http.Error(w, "invalid id: "+s, http.StatusBadRequest)
			return
		}
		ids = append(ids, id)
	}
	if err := h.items.Reorder(r.Context(), dealID, ids); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
