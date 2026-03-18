import apiClient from './client'
import type {
  Product,
  CreateProductRequest,
  UpdateProductRequest,
  PriceBook,
  CreatePriceBookRequest,
  PriceBookEntry,
  DealLineItem,
  UpsertLineItemRequest,
} from './types'

export const productsApi = {
  list: async (activeOnly = true): Promise<Product[]> => {
    const { data } = await apiClient.get<Product[]>('/products', {
      params: { active_only: activeOnly },
    })
    return data
  },

  get: async (id: string): Promise<Product> => {
    const { data } = await apiClient.get<Product>(`/products/${id}`)
    return data
  },

  create: async (payload: CreateProductRequest): Promise<Product> => {
    const { data } = await apiClient.post<Product>('/products', payload)
    return data
  },

  update: async (id: string, payload: UpdateProductRequest): Promise<Product> => {
    const { data } = await apiClient.patch<Product>(`/products/${id}`, payload)
    return data
  },

  deactivate: async (id: string): Promise<void> => {
    await apiClient.delete(`/products/${id}`)
  },
}

export const priceBooksApi = {
  list: async (): Promise<PriceBook[]> => {
    const { data } = await apiClient.get<PriceBook[]>('/price-books')
    return data
  },

  get: async (id: string): Promise<PriceBook> => {
    const { data } = await apiClient.get<PriceBook>(`/price-books/${id}`)
    return data
  },

  create: async (payload: CreatePriceBookRequest): Promise<PriceBook> => {
    const { data } = await apiClient.post<PriceBook>('/price-books', payload)
    return data
  },

  upsertEntry: async (
    priceBookId: string,
    productId: string,
    priceOverride?: number
  ): Promise<PriceBookEntry> => {
    const { data } = await apiClient.put<PriceBookEntry>(
      `/price-books/${priceBookId}/entries/${productId}`,
      { price_override: priceOverride ?? null }
    )
    return data
  },

  deleteEntry: async (priceBookId: string, productId: string): Promise<void> => {
    await apiClient.delete(`/price-books/${priceBookId}/entries/${productId}`)
  },
}

export const lineItemsApi = {
  list: async (dealId: string): Promise<DealLineItem[]> => {
    const { data } = await apiClient.get<DealLineItem[]>(`/deals/${dealId}/line-items`)
    return data
  },

  upsert: async (dealId: string, item: UpsertLineItemRequest): Promise<DealLineItem> => {
    if (item.id) {
      const { data } = await apiClient.put<DealLineItem>(
        `/deals/${dealId}/line-items/${item.id}`,
        item
      )
      return data
    }
    const { data } = await apiClient.post<DealLineItem>(`/deals/${dealId}/line-items`, item)
    return data
  },

  delete: async (dealId: string, itemId: string): Promise<void> => {
    await apiClient.delete(`/deals/${dealId}/line-items/${itemId}`)
  },

  reorder: async (dealId: string, ids: string[]): Promise<void> => {
    await apiClient.post(`/deals/${dealId}/line-items/reorder`, { ids })
  },
}
