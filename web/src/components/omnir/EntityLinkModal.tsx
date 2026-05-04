import { useEffect, useId, useMemo, useRef, useState, type ReactNode } from 'react'
import { Loader2, Search } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/Dialog'
import { Input } from '@/components/ui/Input'

interface EntityLinkModalProps<T> {
  open: boolean
  onClose: () => void
  title: string
  placeholder: string
  search: (query: string) => Promise<T[]>
  onSelect: (item: T) => Promise<unknown>
  renderItem: (item: T, state: { isHighlighted: boolean }) => ReactNode
  getKey: (item: T) => string
  emptyStateText?: string
  noResultsText?: string
  loadingText?: string
}

const SEARCH_DEBOUNCE_MS = 300

export function EntityLinkModal<T>({
  open,
  onClose,
  title,
  placeholder,
  search,
  onSelect,
  renderItem,
  getKey,
  emptyStateText = 'Start typing to search.',
  noResultsText = 'No matches found.',
  loadingText = 'Searching…',
}: EntityLinkModalProps<T>) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<T[]>([])
  const [isSearching, setIsSearching] = useState(false)
  const [isPending, setIsPending] = useState(false)
  const [highlightedIndex, setHighlightedIndex] = useState(-1)
  const inputRef = useRef<HTMLInputElement>(null)
  const contentRef = useRef<HTMLDivElement>(null)
  const optionRefs = useRef<Array<HTMLButtonElement | null>>([])
  const requestIdRef = useRef(0)
  const listboxId = useId()
  const optionIdPrefix = useId()

  useEffect(() => {
    if (!open) return

    setQuery('')
    setResults([])
    setIsSearching(false)
    setIsPending(false)
    setHighlightedIndex(-1)
    requestIdRef.current += 1
    optionRefs.current = []
  }, [open])

  useEffect(() => {
    if (!open) return

    const trimmedQuery = query.trim()
    if (!trimmedQuery) {
      setResults([])
      setIsSearching(false)
      setHighlightedIndex(-1)
      requestIdRef.current += 1
      return
    }

    const currentRequestId = requestIdRef.current + 1
    requestIdRef.current = currentRequestId
    setIsSearching(true)

    const timeoutId = window.setTimeout(async () => {
      try {
        const nextResults = await search(trimmedQuery)
        if (requestIdRef.current !== currentRequestId) return
        setResults(nextResults)
        setHighlightedIndex(nextResults.length > 0 ? 0 : -1)
      } finally {
        if (requestIdRef.current === currentRequestId) {
          setIsSearching(false)
        }
      }
    }, SEARCH_DEBOUNCE_MS)

    return () => {
      window.clearTimeout(timeoutId)
    }
  }, [open, query, search])

  useEffect(() => {
    if (highlightedIndex < 0) return
    const option = optionRefs.current[highlightedIndex]
    if (option && typeof option.scrollIntoView === 'function') {
      option.scrollIntoView({ block: 'nearest' })
    }
  }, [highlightedIndex])

  const hasQuery = query.trim().length > 0
  const listboxOpen = open && (hasQuery || isSearching)
  const activeOptionId =
    highlightedIndex >= 0 && results[highlightedIndex]
      ? `${optionIdPrefix}-${getKey(results[highlightedIndex])}`
      : undefined

  const statusMessage = useMemo(() => {
    if (isSearching) return loadingText
    if (!hasQuery) return emptyStateText
    if (results.length === 0) return noResultsText
    return null
  }, [emptyStateText, hasQuery, isSearching, loadingText, noResultsText, results.length])

  const selectItem = async (item: T) => {
    setIsPending(true)
    try {
      await onSelect(item)
      onClose()
    } catch {
      // Leave the modal open so mutation-level UI can surface errors and allow retry.
    } finally {
      setIsPending(false)
    }
  }

  const moveHighlight = (direction: 1 | -1) => {
    if (!results.length) return
    setHighlightedIndex((current) => {
      if (current < 0) return direction === 1 ? 0 : results.length - 1
      if (direction === 1) return current >= results.length - 1 ? 0 : current + 1
      return current <= 0 ? results.length - 1 : current - 1
    })
  }

  const trapFocusInsideModal = (event: React.KeyboardEvent<HTMLDivElement>) => {
    if (event.key !== 'Tab') return
    const container = contentRef.current
    if (!container) return

    const focusable = Array.from(
      container.querySelectorAll<HTMLElement>(
        'button:not([disabled]), input:not([disabled]), textarea:not([disabled]), select:not([disabled]), [href], [tabindex]:not([tabindex="-1"])'
      )
    ).filter((el) => !el.hasAttribute('disabled') && !el.getAttribute('aria-hidden'))

    if (focusable.length === 0) return

    const activeElement = document.activeElement as HTMLElement | null
    const currentIndex = activeElement ? focusable.indexOf(activeElement) : -1

    if (event.shiftKey) {
      if (currentIndex <= 0) {
        event.preventDefault()
        focusable[focusable.length - 1]?.focus()
      }
      return
    }

    if (currentIndex === focusable.length - 1) {
      event.preventDefault()
      focusable[0]?.focus()
    }
  }

  const handleInputKeyDown = async (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Escape') {
      event.preventDefault()
      onClose()
      return
    }

    if (event.key === 'ArrowDown') {
      event.preventDefault()
      moveHighlight(1)
      return
    }

    if (event.key === 'ArrowUp') {
      event.preventDefault()
      moveHighlight(-1)
      return
    }

    if (event.key === 'Home' && results.length > 0) {
      event.preventDefault()
      setHighlightedIndex(0)
      return
    }

    if (event.key === 'End' && results.length > 0) {
      event.preventDefault()
      setHighlightedIndex(results.length - 1)
      return
    }

    if (event.key === 'Tab' && !event.shiftKey && results.length > 0) {
      event.preventDefault()
      const nextIndex = highlightedIndex >= 0 ? highlightedIndex : 0
      optionRefs.current[nextIndex]?.focus()
      return
    }

    if (event.key === 'Enter' && highlightedIndex >= 0 && results[highlightedIndex]) {
      event.preventDefault()
      await selectItem(results[highlightedIndex])
    }
  }

  return (
    <Dialog open={open} onOpenChange={(nextOpen) => !nextOpen && onClose()}>
      <DialogContent
        ref={contentRef}
        className="max-w-sm border p-0 shadow-xl"
        style={{ background: 'var(--surface-card)', borderColor: 'var(--border-default)' }}
        onOpenAutoFocus={(event) => {
          event.preventDefault()
          inputRef.current?.focus()
        }}
        onKeyDown={trapFocusInsideModal}
        aria-describedby={undefined}
      >
        <DialogHeader
          className="border-b px-4 py-3"
          style={{ borderColor: 'var(--border-default)' }}
        >
          <DialogTitle
            className="text-sm font-semibold"
            style={{ color: 'var(--text-primary)' }}
          >
            {title}
          </DialogTitle>
        </DialogHeader>

        <div className="px-4 py-3">
          <div className="relative">
            <Search
              className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2"
              style={{ color: 'var(--text-label)' }}
            />
            <Input
              ref={inputRef}
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              onKeyDown={(event) => void handleInputKeyDown(event)}
              placeholder={placeholder}
              className="h-10 border pl-8 pr-8 text-sm shadow-none"
              style={{
                borderColor: 'var(--border-default)',
                color: 'var(--text-primary)',
                background: 'var(--surface-app)',
              }}
              aria-label={placeholder}
              role="combobox"
              aria-haspopup="listbox"
              aria-expanded={listboxOpen}
              aria-controls={listboxId}
              aria-activedescendant={activeOptionId}
              aria-autocomplete="list"
            />
            {isSearching && (
              <Loader2
                className="absolute right-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 animate-spin"
                style={{ color: 'var(--text-label)' }}
                aria-hidden="true"
              />
            )}
          </div>
        </div>

        <div className="max-h-64 overflow-y-auto px-4 pb-4">
          {statusMessage ? (
            <div
              id={listboxId}
              role="listbox"
              aria-label={`${title} results`}
            >
              <p className="py-6 text-center text-sm" style={{ color: 'var(--text-label)' }}>
                {statusMessage}
              </p>
            </div>
          ) : (
            <div
              id={listboxId}
              role="listbox"
              aria-label={`${title} results`}
              className="space-y-2"
            >
              {results.map((item, index) => {
                const key = getKey(item)
                const isHighlighted = index === highlightedIndex

                return (
                  <button
                    key={key}
                    id={`${optionIdPrefix}-${key}`}
                    ref={(node) => {
                      optionRefs.current[index] = node
                    }}
                    type="button"
                    role="option"
                    aria-selected={isHighlighted}
                    disabled={isPending}
                    className="w-full rounded-lg border px-3 py-2 text-left transition-colors disabled:opacity-50"
                    style={{
                      borderColor: isHighlighted ? 'var(--border-focus)' : 'var(--border-subtle)',
                      background: isHighlighted ? 'var(--surface-app)' : 'transparent',
                    }}
                    onMouseEnter={() => setHighlightedIndex(index)}
                    onMouseDown={(event) => {
                      event.preventDefault()
                      void selectItem(item)
                    }}
                    onKeyDown={(event) => {
                      if (event.key === 'Escape') {
                        event.preventDefault()
                        onClose()
                        return
                      }
                      if (event.key === 'ArrowDown') {
                        event.preventDefault()
                        moveHighlight(1)
                        optionRefs.current[index === results.length - 1 ? 0 : index + 1]?.focus()
                        return
                      }
                      if (event.key === 'ArrowUp') {
                        event.preventDefault()
                        moveHighlight(-1)
                        optionRefs.current[index === 0 ? results.length - 1 : index - 1]?.focus()
                        return
                      }
                      if (event.key === 'Home') {
                        event.preventDefault()
                        setHighlightedIndex(0)
                        optionRefs.current[0]?.focus()
                        return
                      }
                      if (event.key === 'End') {
                        event.preventDefault()
                        setHighlightedIndex(results.length - 1)
                        optionRefs.current[results.length - 1]?.focus()
                        return
                      }
                      if (event.key === 'Tab' && event.shiftKey) {
                        event.preventDefault()
                        inputRef.current?.focus()
                        return
                      }
                      if ((event.key === 'Enter' || event.key === ' ') && !isPending) {
                        event.preventDefault()
                        void selectItem(item)
                      }
                    }}
                  >
                    {renderItem(item, { isHighlighted })}
                  </button>
                )
              })}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}

export type { EntityLinkModalProps }
