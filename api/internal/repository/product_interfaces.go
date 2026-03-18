package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

// ProductRepository defines the persistence contract for the product catalog and price books.
type ProductRepository interface {
	CreateProduct(ctx context.Context, p *domain.Product) (*domain.Product, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	ListProducts(ctx context.Context, activeOnly bool) ([]*domain.Product, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, patch domain.ProductPatch) (*domain.Product, error)
	DeactivateProduct(ctx context.Context, id uuid.UUID) error
	CreatePriceBook(ctx context.Context, pb *domain.PriceBook) (*domain.PriceBook, error)
	GetPriceBook(ctx context.Context, id uuid.UUID) (*domain.PriceBook, error)
	ListPriceBooks(ctx context.Context) ([]*domain.PriceBook, error)
	UpsertPriceBookEntry(ctx context.Context, e *domain.PriceBookEntry) (*domain.PriceBookEntry, error)
	DeletePriceBookEntry(ctx context.Context, priceBookID, productID uuid.UUID) error
}

// DealLineItemRepository defines the persistence contract for deal line items.
type DealLineItemRepository interface {
	List(ctx context.Context, dealID uuid.UUID) ([]*domain.DealLineItem, error)
	Upsert(ctx context.Context, item *domain.DealLineItem) (*domain.DealLineItem, error)
	Delete(ctx context.Context, id, dealID uuid.UUID) error
	Reorder(ctx context.Context, dealID uuid.UUID, ids []uuid.UUID) error
}
