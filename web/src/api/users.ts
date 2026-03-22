import apiClient from './client'
import type {
  User,
  CreateUserRequest,
  UpdateUserRequest,
  UserListParams,
  PaginatedResponse,
} from './types'

export const usersApi = {
  list: async (params?: UserListParams): Promise<PaginatedResponse<User>> => {
    const { data } = await apiClient.get('/users', { params })
    return data
  },

  get: async (id: string): Promise<User> => {
    const { data } = await apiClient.get(`/users/${id}`)
    return data
  },

  me: async (): Promise<User> => {
    const { data } = await apiClient.get('/users/me')
    return data
  },

  create: async (payload: CreateUserRequest): Promise<User> => {
    const { data } = await apiClient.post('/users', payload)
    return data
  },

  update: async (id: string, payload: UpdateUserRequest): Promise<User> => {
    const { data } = await apiClient.patch(`/users/${id}`, payload)
    return data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/users/${id}`)
  },
}
