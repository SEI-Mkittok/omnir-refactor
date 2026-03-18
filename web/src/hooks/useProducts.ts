import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { productsApi, priceBooksApi, lineItemsApi } from '../api/products'
import type {
  CreateProductRequest,
  UpdateProductRequest,
  CreatePriceBookRequest,
  UpsertLineItemRequest,
} from '../api/types'

export function useProducts(activeOnly = true) {
  return useQuery({
    queryKey: ['products', { activeOnly }],
    queryFn: () => productsApi.list(activeOnly),
  })
}

export function useProduct(id: string) {
  return useQuery({
    queryKey: ['products', id],
    queryFn: () => productsApi.get(id),
    enabled: !!id,
  })
}

export function useCreateProduct() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateProductRequest) => productsApi.create(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['products'] }),
  })
}

export function useUpdateProduct() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...payload }: UpdateProductRequest & { id: string }) =>
      productsApi.update(id, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['products'] }),
  })
}

export function useDeactivateProduct() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => productsApi.deactivate(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['products'] }),
  })
}

export function usePriceBooks() {
  return useQuery({
    queryKey: ['price-books'],
    queryFn: priceBooksApi.list,
  })
}

export function usePriceBook(id: string) {
  return useQuery({
    queryKey: ['price-books', id],
    queryFn: () => priceBooksApi.get(id),
    enabled: !!id,
  })
}

export function useCreatePriceBook() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreatePriceBookRequest) => priceBooksApi.create(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['price-books'] }),
  })
}

export function useUpsertPriceBookEntry(priceBookId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      productId,
      priceOverride,
    }: {
      productId: string
      priceOverride?: number
    }) => priceBooksApi.upsertEntry(priceBookId, productId, priceOverride),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['price-books'] }),
  })
}

export function useDeletePriceBookEntry(priceBookId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (productId: string) => priceBooksApi.deleteEntry(priceBookId, productId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['price-books'] }),
  })
}

export function useDealLineItems(dealId: string) {
  return useQuery({
    queryKey: ['line-items', dealId],
    queryFn: () => lineItemsApi.list(dealId),
    enabled: !!dealId,
  })
}

export function useUpsertLineItem(dealId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (item: UpsertLineItemRequest) => lineItemsApi.upsert(dealId, item),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['line-items', dealId] }),
  })
}

export function useDeleteLineItem(dealId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (itemId: string) => lineItemsApi.delete(dealId, itemId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['line-items', dealId] }),
  })
}

export function useReorderLineItems(dealId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (ids: string[]) => lineItemsApi.reorder(dealId, ids),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['line-items', dealId] }),
  })
}
