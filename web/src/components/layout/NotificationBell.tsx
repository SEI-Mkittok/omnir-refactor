import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { AlertTriangle, Bell, BellRing, CheckCheck, Activity, TrendingUp, AtSign, UserPlus } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { notificationsApi } from '@/api/notifications'
import { formatRelativeTime } from '@/lib/utils'
import type { Notification, NotificationKind } from '@/api/types'

function kindIcon(kind: NotificationKind) {
  switch (kind) {
    case 'activity_reminder': return Activity
    case 'deal_stage_changed': return TrendingUp
    case 'mention': return AtSign
    case 'assignment': return UserPlus
    case 'sla_warning':
    case 'sla_breached': return AlertTriangle
  }
}

function entityPath(n: Notification): string | null {
  if (!n.entity_id || !n.entity_type) return null
  switch (n.entity_type) {
    case 'deal': return `/deals?openId=${n.entity_id}`
    case 'contact': return `/contacts?openId=${n.entity_id}`
    case 'ticket': return `/tickets?openId=${n.entity_id}`
    case 'activity': return `/deals`
    default: return null
  }
}

export function NotificationBell() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const { data: countData } = useQuery({
    queryKey: ['notifications-unread-count'],
    queryFn: notificationsApi.unreadCount,
    refetchInterval: 60_000,
    staleTime: 30_000,
  })

  const { data: listData } = useQuery({
    queryKey: ['notifications-dropdown'],
    queryFn: () => notificationsApi.list({ limit: 10 }),
    enabled: open,
    staleTime: 30_000,
  })

  const markRead = useMutation({
    mutationFn: notificationsApi.markRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications-unread-count'] })
      queryClient.invalidateQueries({ queryKey: ['notifications-dropdown'] })
      queryClient.invalidateQueries({ queryKey: ['notifications-list'] })
    },
  })

  const markAllRead = useMutation({
    mutationFn: notificationsApi.markAllRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications-unread-count'] })
      queryClient.invalidateQueries({ queryKey: ['notifications-dropdown'] })
      queryClient.invalidateQueries({ queryKey: ['notifications-list'] })
    },
  })

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const unreadCount = countData?.count ?? 0
  const notifications = listData?.data ?? []

  function handleNotificationClick(n: Notification) {
    if (!n.read_at) markRead.mutate(n.id)
    const path = entityPath(n)
    if (path) navigate(path)
    setOpen(false)
  }

  return (
    <div ref={containerRef} className="relative">
      <button
        onClick={() => setOpen((v) => !v)}
        className="relative flex h-8 w-8 items-center justify-center rounded-full text-slate-500 hover:bg-slate-100 hover:text-slate-900 transition-colors"
        aria-label="Notifications"
      >
        {unreadCount > 0 ? (
          <BellRing className="h-5 w-5" />
        ) : (
          <Bell className="h-5 w-5" />
        )}
        {unreadCount > 0 && (
          <span className="absolute -top-0.5 -right-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-[var(--color-primary)] text-[10px] font-bold text-white leading-none">
            {unreadCount > 99 ? '99+' : unreadCount}
          </span>
        )}
      </button>

      {open && (
        <div className="absolute right-0 top-full mt-2 w-80 rounded-lg border border-slate-200 bg-white shadow-xl z-50 overflow-hidden">
          {/* Header */}
          <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
            <span className="text-sm font-semibold text-slate-800">Notifications</span>
            {unreadCount > 0 && (
              <button
                onClick={() => markAllRead.mutate()}
                disabled={markAllRead.isPending}
                className="flex items-center gap-1 text-xs text-[var(--color-primary)] hover:text-[var(--color-primary)] disabled:opacity-50 transition-colors"
              >
                <CheckCheck className="h-3.5 w-3.5" />
                Mark all read
              </button>
            )}
          </div>

          {/* Notification list */}
          <div className="max-h-96 overflow-y-auto">
            {notifications.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-10 text-slate-400">
                <Bell className="h-8 w-8 mb-2 opacity-40" />
                <p className="text-sm">No notifications</p>
              </div>
            ) : (
              notifications.map((n) => {
                const Icon = kindIcon(n.kind)
                const isUnread = !n.read_at
                return (
                  <button
                    key={n.id}
                    onClick={() => handleNotificationClick(n)}
                    className={`w-full flex items-start gap-3 px-4 py-3 text-left transition-colors hover:bg-slate-50 border-b border-slate-50 last:border-0 ${
                      isUnread ? 'bg-[var(--color-primary-light)]/50' : ''
                    }`}
                  >
                    <div className={`mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full ${
                      isUnread ? 'bg-[var(--color-primary-light)] text-[var(--color-primary)]' : 'bg-slate-100 text-slate-500'
                    }`}>
                      <Icon className="h-3.5 w-3.5" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className={`text-sm leading-snug ${isUnread ? 'font-semibold text-slate-800' : 'text-slate-700'}`}>
                        {n.title}
                      </p>
                      {n.body && (
                        <p className="mt-0.5 text-xs text-slate-500 truncate">{n.body}</p>
                      )}
                      <p className="mt-1 text-xs text-slate-400">{formatRelativeTime(n.created_at)}</p>
                    </div>
                    {isUnread && (
                      <span className="mt-2 h-2 w-2 shrink-0 rounded-full bg-[var(--color-primary-light)]0" />
                    )}
                  </button>
                )
              })
            )}
          </div>

          {/* Footer */}
          <div className="border-t border-slate-100">
            <button
              onClick={() => { navigate('/notifications'); setOpen(false) }}
              className="w-full px-4 py-2.5 text-center text-xs font-medium text-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-colors"
            >
              View all notifications
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
