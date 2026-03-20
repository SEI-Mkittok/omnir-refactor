import apiClient from './client'
import type {
  Product,
  CreateProductRequest,
  UpdateProductRequest,
  ProductListParams,
  PaginatedResponse,
} from './types'

export const productsApi = {
  list: async (params?: ProductListParams): Promise<PaginatedResponse<Product>> => {
    const { data } = await apiClient.get('/products', { params })
    return data
  },

  get: async (id: string): Promise<Product> => {
    const { data } = await apiClient.get(`/products/${id}`)
    return data
  },

  create: async (payload: CreateProductRequest): Promise<Product> => {
    const { data } = await apiClient.post('/products', payload)
    return data
  },

  update: async (id: string, payload: UpdateProductRequest): Promise<Product> => {
    const { data } = await apiClient.patch(`/products/${id}`, payload)
    return data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/products/${id}`)
  },
}
