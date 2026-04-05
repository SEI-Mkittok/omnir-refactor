import { Component, type ErrorInfo, type ReactNode } from 'react'
import { AlertCircle } from 'lucide-react'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('[ErrorBoundary]', error, info.componentStack)
  }

  render() {
    if (this.state.error) {
      if (this.props.fallback) return this.props.fallback
      return (
        <div className="flex flex-col items-center justify-center py-24 text-center gap-4 px-4">
          <AlertCircle className="h-12 w-12 text-[var(--color-danger)]" />
          <p className="text-[18px] font-semibold text-[var(--text-primary)]">Something went wrong</p>
          <p className="text-[14px] text-[var(--text-secondary)] max-w-md">
            {this.state.error.message}
          </p>
          <button
            onClick={() => this.setState({ error: null })}
            className="h-9 rounded-md border border-[var(--border-default)] px-4 text-[13px] text-[var(--text-primary)] hover:bg-[var(--surface-app)] focus:outline-none focus:ring-2 focus:ring-[var(--border-focus)]"
          >
            Try again
          </button>
        </div>
      )
    }
    return this.props.children
  }
}
