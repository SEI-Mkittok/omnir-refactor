import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { usersApi } from '@/api/users'
import type { CreateUserRequest, UpdateUserRequest, UserListParams } from '@/api/types'

export const userKeys = {
  all: ['users'] as const,
  lists: () => [...userKeys.all, 'list'] as const,
  list: (params?: UserListParams) => [...userKeys.lists(), params] as const,
  details: () => [...userKeys.all, 'detail'] as const,
  detail: (id: string) => [...userKeys.details(), id] as const,
  me: () => [...userKeys.all, 'me'] as const,
  assignmentOptions: () => [...userKeys.all, 'assignment-options'] as const,
}

export function useUsers(params?: UserListParams) {
  return useQuery({
    queryKey: userKeys.list(params),
    queryFn: () => usersApi.list(params),
    staleTime: 30_000,
  })
}

export function useUser(id: string) {
  return useQuery({
    queryKey: userKeys.detail(id),
    queryFn: () => usersApi.get(id),
    staleTime: 30_000,
    enabled: !!id,
  })
}

export function useMe() {
  return useQuery({
    queryKey: userKeys.me(),
    queryFn: () => usersApi.me(),
    staleTime: 60_000,
  })
}

export function useUserAssignmentOptions(enabled = true) {
  return useQuery({
    queryKey: userKeys.assignmentOptions(),
    queryFn: usersApi.assignmentOptions,
    staleTime: 60_000,
    enabled,
  })
}

export function useCreateUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateUserRequest) => usersApi.create(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: userKeys.lists() })
    },
  })
}

export function useUpdateUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateUserRequest }) =>
      usersApi.update(id, payload),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: userKeys.lists() })
      qc.invalidateQueries({ queryKey: userKeys.detail(id) })
      qc.invalidateQueries({ queryKey: userKeys.me() })
    },
  })
}

export function useDeleteUser() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => usersApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: userKeys.lists() })
    },
  })
}
