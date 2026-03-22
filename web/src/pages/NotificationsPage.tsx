import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Bell, Activity, TrendingUp, AtSign, UserPlus, CheckCheck } from 'lucide-react'
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { notificationsApi } from '@/api/notifications'
import { formatRelativeTime } from '@/lib/utils'
import { Spinner } from '@/components/ui/Spinner'
import type { Notification, NotificationKind } from '@/api/types'

type FilterMode = 'all' | 'unread'

function kindIcon(kind: NotificationKind) {
  switch (kind) {
    case 'activity_reminder': return Activity
    case 'deal_stage_changed': return TrendingUp
    case 'mention': return AtSign
    case 'assignment': return UserPlus
  }
}

function kindLabel(kind: NotificationKind): string {
  switch (kind) {
    case 'activity_reminder': return 'Reminder'
    case 'deal_stage_changed': return 'Deal Update'
    case 'mention': return 'Mention'
    case 'assignment': return 'Assignment'
  }
}

function entityPath(n: Notification): string | null {
  if (!n.entity_id || !n.entity_type) return null
  switch (n.entity_type) {
    case 'deal': return `/deals?openId=${n.entity_id}`
    case 'contact': return `/contacts?openId=${n.entity_id}`
    default: return null
  }
}

export function NotificationsPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [filter, setFilter] = useState<FilterMode>('all')

  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
  } = useInfiniteQuery({
    queryKey: ['notifications-list', filter],
    queryFn: ({ pageParam }) =>
      notificationsApi.list({
        limit: 25,
        before: pageParam as string | undefined,
        unread_only: filter === 'unread' ? true : undefined,
      }),
    getNextPageParam: (lastPage) => {
      const items = lastPage.data
      if (!items || items.length < 25) return undefined
      return items[items.length - 1].created_at
    },
    initialPageParam: undefined as string | undefined,
    staleTime: 30_000,
  })

  const markRead = useMutation({
    mutationFn: notificationsApi.markRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications-unread-count'] })
      queryClient.invalidateQueries({ queryKey: ['notifications-list'] })
      queryClient.invalidateQueries({ queryKey: ['notifications-dropdown'] })
    },
  })

  const markAllRead = useMutation({
    mutationFn: notificationsApi.markAllRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications-unread-count'] })
      queryClient.invalidateQueries({ queryKey: ['notifications-list'] })
      queryClient.invalidateQueries({ queryKey: ['notifications-dropdown'] })
    },
  })

  const notifications: Notification[] = data?.pages.flatMap((p) => p.data ?? []) ?? []

  function handleClick(n: Notification) {
    if (!n.read_at) markRead.mutate(n.id)
    const path = entityPath(n)
    if (path) navigate(path)
  }

  return (
    <div className="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-900">Notifications</h1>
        <button
          onClick={() => markAllRead.mutate()}
          disabled={markAllRead.isPending || notifications.every((n) => !!n.read_at)}
          className="flex items-center gap-1.5 rounded-md border border-slate-200 px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-50 disabled:opacity-40 transition-colors"
        >
          <CheckCheck className="h-4 w-4" />
          Mark all as read
        </button>
      </div>

      {/* Filter tabs */}
      <div className="mb-4 flex gap-1 rounded-lg border border-slate-200 bg-slate-50 p-1 w-fit">
        {(['all', 'unread'] as FilterMode[]).map((mode) => (
          <button
            key={mode}
            onClick={() => setFilter(mode)}
            className={`rounded-md px-4 py-1.5 text-sm font-medium transition-colors capitalize ${
              filter === mode
                ? 'bg-white text-slate-900 shadow-sm'
                : 'text-slate-500 hover:text-slate-700'
            }`}
          >
            {mode}
          </button>
        ))}
      </div>

      {/* Content */}
      {isLoading ? (
        <div className="flex justify-center py-20">
          <Spinner className="h-6 w-6 text-slate-400" />
        </div>
      ) : notifications.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed border-slate-200 py-20 text-slate-400">
          <Bell className="h-10 w-10 mb-3 opacity-40" />
          <p className="text-sm font-medium">
            {filter === 'unread' ? 'No unread notifications' : 'No notifications yet'}
          </p>
        </div>
      ) : (
        <div className="divide-y divide-slate-100 rounded-lg border border-slate-200 bg-white overflow-hidden">
          {notifications.map((n) => {
            const Icon = kindIcon(n.kind)
            const isUnread = !n.read_at
            const path = entityPath(n)
            return (
              <div
                key={n.id}
                onClick={() => handleClick(n)}
                className={`flex cursor-pointer items-start gap-4 px-5 py-4 transition-colors hover:bg-slate-50 ${
                  isUnread ? 'bg-[var(--color-primary-light)]/40' : ''
                }`}
              >
                <div className={`mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full ${
                  isUnread ? 'bg-[var(--color-primary-light)] text-[var(--color-primary)]' : 'bg-slate-100 text-slate-500'
                }`}>
                  <Icon className="h-4 w-4" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-baseline gap-2">
                    <span className={`text-sm ${isUnread ? 'font-semibold text-slate-900' : 'text-slate-700'}`}>
                      {n.title}
                    </span>
                    <span className="shrink-0 rounded-full bg-slate-100 px-2 py-0.5 text-[10px] font-medium text-slate-500 uppercase tracking-wide">
                      {kindLabel(n.kind)}
                    </span>
                  </div>
                  {n.body && (
                    <p className="mt-0.5 text-sm text-slate-500 line-clamp-2">{n.body}</p>
                  )}
                  <p className="mt-1 text-xs text-slate-400">{formatRelativeTime(n.created_at)}</p>
                </div>
                <div className="flex shrink-0 flex-col items-end gap-2">
                  {isUnread && <span className="h-2 w-2 rounded-full bg-[var(--color-primary-light)]0" />}
                  {path && (
                    <span className="text-xs text-[var(--color-primary)] hover:underline">View →</span>
                  )}
                </div>
              </div>
            )
          })}

          {/* Load more */}
          {hasNextPage && (
            <div className="flex justify-center py-4">
              <button
                onClick={() => fetchNextPage()}
                disabled={isFetchingNextPage}
                className="flex items-center gap-2 text-sm text-[var(--color-primary)] hover:text-[var(--color-primary)] disabled:opacity-50 transition-colors"
              >
                {isFetchingNextPage ? (
                  <>
                    <Spinner className="h-3.5 w-3.5" />
                    Loading…
                  </>
                ) : (
                  'Load more'
                )}
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
