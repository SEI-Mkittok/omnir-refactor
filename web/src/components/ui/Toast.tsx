import * as React from 'react'
import * as ToastPrimitive from '@radix-ui/react-toast'
import { X } from 'lucide-react'

// ── Minimal toast context ─────────────────────────────────────────────────────

interface ToastMessage {
  id: string
  title: string
  description?: string
  variant?: 'default' | 'destructive'
}

interface ToastContextValue {
  toast: (opts: Omit<ToastMessage, 'id'>) => void
}

const ToastContext = React.createContext<ToastContextValue | null>(null)

export function useToast(): ToastContextValue {
  const ctx = React.useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used inside <Toaster>')
  return ctx
}

// ── Provider + Viewport ───────────────────────────────────────────────────────

export function Toaster({ children }: { children: React.ReactNode }) {
  const [messages, setMessages] = React.useState<ToastMessage[]>([])

  const toast = React.useCallback((opts: Omit<ToastMessage, 'id'>) => {
    const id = Math.random().toString(36).slice(2)
    setMessages((prev) => [...prev, { ...opts, id }])
  }, [])

  const dismiss = (id: string) =>
    setMessages((prev) => prev.filter((m) => m.id !== id))

  return (
    <ToastContext.Provider value={{ toast }}>
      <ToastPrimitive.Provider swipeDirection="right">
        {children}
        {messages.map((m) => (
          <ToastPrimitive.Root
            key={m.id}
            open
            onOpenChange={(open) => { if (!open) dismiss(m.id) }}
            duration={4000}
            className={[
              'group pointer-events-auto relative flex w-full items-center justify-between',
              'gap-4 overflow-hidden rounded-md border px-4 py-3 shadow-md transition-all',
              'data-[swipe=move]:translate-x-[var(--radix-toast-swipe-move-x)]',
              'data-[swipe=cancel]:translate-x-0',
              'data-[swipe=end]:translate-x-[var(--radix-toast-swipe-end-x)]',
              'data-[state=open]:animate-in data-[state=open]:slide-in-from-bottom-4',
              'data-[state=closed]:animate-out data-[state=closed]:fade-out-80 data-[state=closed]:slide-out-to-right-full',
              m.variant === 'destructive'
                ? 'border-red-200 bg-red-50 text-red-900'
                : 'border-slate-200 bg-white text-slate-900',
            ].join(' ')}
          >
            <div className="flex flex-col gap-0.5">
              <ToastPrimitive.Title className="text-sm font-semibold">
                {m.title}
              </ToastPrimitive.Title>
              {m.description && (
                <ToastPrimitive.Description className="text-xs text-slate-500">
                  {m.description}
                </ToastPrimitive.Description>
              )}
            </div>
            <ToastPrimitive.Close
              className="shrink-0 rounded p-0.5 opacity-60 hover:opacity-100 focus:outline-none focus:ring-1 focus:ring-slate-400"
              aria-label="Dismiss"
            >
              <X className="h-4 w-4" />
            </ToastPrimitive.Close>
          </ToastPrimitive.Root>
        ))}
        <ToastPrimitive.Viewport className="fixed bottom-4 right-4 z-50 flex w-80 flex-col gap-2 outline-none" />
      </ToastPrimitive.Provider>
    </ToastContext.Provider>
  )
}
