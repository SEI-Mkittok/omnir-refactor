import { render, type RenderOptions } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import type { ReactElement } from 'react'
import { Toaster } from '@/components/ui/Toast'

/**
 * Custom render that wraps components with all required providers:
 * - QueryClientProvider (React Query)
 * - MemoryRouter (react-router)
 *
 * Usage:
 *   import { render } from '@/test/utils'
 *   render(<ContactCard contact={...} />)
 */
function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,           // don't retry in tests
        staleTime: Infinity,    // don't refetch during test
      },
    },
  })
}

interface CustomRenderOptions extends Omit<RenderOptions, 'wrapper'> {
  initialRoute?: string
}

function customRender(ui: ReactElement, { initialRoute = '/', ...options }: CustomRenderOptions = {}) {
  const queryClient = makeQueryClient()

  function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <Toaster>
          <MemoryRouter initialEntries={[initialRoute]}>
            {children}
          </MemoryRouter>
        </Toaster>
      </QueryClientProvider>
    )
  }

  return render(ui, { wrapper: Wrapper, ...options })
}

// Re-export everything from RTL so test files only need to import from utils
export * from '@testing-library/react'
export { customRender as render }
