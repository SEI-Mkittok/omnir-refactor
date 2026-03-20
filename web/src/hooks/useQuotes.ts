import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { quotesApi } from '@/api/quotes'
import type { QuoteListParams, CreateQuoteRequest, UpdateQuoteRequest, SendQuoteRequest } from '@/api/types'

export const quoteKeys = {
  all: ['quotes'] as const,
  lists: () => [...quoteKeys.all, 'list'] as const,
  list: (params?: QuoteListParams) => [...quoteKeys.lists(), params] as const,
  byDeal: (dealId: string) => [...quoteKeys.all, 'deal', dealId] as const,
  details: () => [...quoteKeys.all, 'detail'] as const,
  detail: (id: string) => [...quoteKeys.details(), id] as const,
}

export function useQuotes(params?: QuoteListParams) {
  return useQuery({
    queryKey: quoteKeys.list(params),
    queryFn: () => quotesApi.list(params),
    staleTime: 30_000,
  })
}

export function useDealQuotes(dealId: string) {
  return useQuery({
    queryKey: quoteKeys.byDeal(dealId),
    queryFn: () => quotesApi.listByDeal(dealId),
    staleTime: 30_000,
    enabled: !!dealId,
  })
}

export function useQuote(id: string) {
  return useQuery({
    queryKey: quoteKeys.detail(id),
    queryFn: () => quotesApi.get(id),
    staleTime: 30_000,
    enabled: !!id,
  })
}

export function useCreateQuote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateQuoteRequest) => quotesApi.create(payload),
    onSuccess: (quote) => {
      qc.invalidateQueries({ queryKey: quoteKeys.lists() })
      if (quote.deal_id) {
        qc.invalidateQueries({ queryKey: quoteKeys.byDeal(quote.deal_id) })
      }
    },
  })
}

export function useUpdateQuote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateQuoteRequest }) =>
      quotesApi.update(id, payload),
    onSuccess: (quote, { id }) => {
      qc.invalidateQueries({ queryKey: quoteKeys.lists() })
      qc.invalidateQueries({ queryKey: quoteKeys.detail(id) })
      if (quote.deal_id) {
        qc.invalidateQueries({ queryKey: quoteKeys.byDeal(quote.deal_id) })
      }
    },
  })
}

export function useDeleteQuote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => quotesApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: quoteKeys.lists() })
    },
  })
}

export function useSendQuote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: SendQuoteRequest }) =>
      quotesApi.send(id, payload),
    onSuccess: (quote, { id }) => {
      qc.invalidateQueries({ queryKey: quoteKeys.detail(id) })
      qc.invalidateQueries({ queryKey: quoteKeys.lists() })
      if (quote.deal_id) {
        qc.invalidateQueries({ queryKey: quoteKeys.byDeal(quote.deal_id) })
      }
    },
  })
}
