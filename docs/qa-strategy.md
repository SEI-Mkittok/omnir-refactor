# Omnir CRM — QA Testing Strategy

> Status: Approved  
> Issue: OMN-9  
> Author: Forge (AI Dev Agent)  
> Date: 2026-03-17

---

## 1. Philosophy

Test at the layer where bugs live. Unit tests for logic; integration tests for DB correctness; E2E for critical user journeys. Avoid testing framework glue—test behavior.

**Coverage targets:**

| Layer | Target | CI Gate |
|---|---|---|
| Go unit tests | ≥ 80% | Block merge below 70% |
| Go integration tests | critical paths (auth, CRUD) | Block merge on failure |
| Frontend component tests | ≥ 75% | Block merge below 60% |
| E2E (Playwright) | happy paths for 5 core flows | Block merge on failure |

---

## 2. Go API — Unit Tests

### Convention: Table-Driven Tests

All Go unit tests use table-driven style with `t.Run`. This minimizes boilerplate and makes it easy to add edge cases.

```go
func TestContactValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   domain.Contact
        wantErr bool
    }{
        {"valid contact", domain.Contact{FirstName: "Ada", Email: "ada@example.com"}, false},
        {"missing first name", domain.Contact{Email: "ada@example.com"}, true},
        {"invalid email", domain.Contact{FirstName: "Ada", Email: "not-an-email"}, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.input.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Mock Strategy

Repository interfaces are mocked using `testify/mock` (or hand-rolled stubs for simple cases). **Never** hit a real DB in unit tests.

```go
// internal/repository/mock/contacts.go
type MockContactRepository struct {
    mock.Mock
}

func (m *MockContactRepository) Create(ctx context.Context, c *domain.Contact) (*domain.Contact, error) {
    args := m.Called(ctx, c)
    return args.Get(0).(*domain.Contact), args.Error(1)
}
```

Handlers are tested by injecting mock repositories:

```go
func TestContactHandler_Create(t *testing.T) {
    mockRepo := new(mock.MockContactRepository)
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Contact")).
        Return(&domain.Contact{ID: uuid.New(), FirstName: "Ada"}, nil)

    h := handler.NewContactHandler(mockRepo)
    // ... http.NewRecorder, call handler, assert
}
```

### File Conventions

- Test files co-located: `internal/handler/contacts_test.go`
- Package: `package handler_test` (black-box) for handler tests; `package domain` (white-box) for domain logic tests
- Shared test fixtures in `internal/testutil/fixtures.go`
- Shared mock setup in `internal/testutil/mocks/`

---

## 3. Go API — Integration Tests

### Approach

Integration tests run against a real PostgreSQL instance (the `omnir_crm_test` DB defined in `docker-compose.test.yml`). They test the full request-to-DB round-trip.

### Test DB Setup

```go
// internal/testutil/db.go
func NewTestDB(t *testing.T) *pgxpool.Pool {
    t.Helper()
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "postgres://omnir:omnir_dev@localhost:5433/omnir_crm_test?sslmode=disable"
    }
    pool, err := pgxpool.New(context.Background(), dsn)
    require.NoError(t, err)
    t.Cleanup(func() { pool.Close() })
    return pool
}

func TruncateAll(t *testing.T, pool *pgxpool.Pool) {
    t.Helper()
    _, err := pool.Exec(context.Background(),
        "TRUNCATE contacts, accounts, deals, activities, users, pipelines, notes RESTART IDENTITY CASCADE")
    require.NoError(t, err)
}
```

Each integration test truncates tables in `TestMain` setup and after each test to ensure isolation.

### API Contract Tests

The full HTTP stack is tested using `httptest.NewServer` with a real Chi router and real DB pool:

```go
func TestContactsAPI(t *testing.T) {
    db := testutil.NewTestDB(t)
    testutil.TruncateAll(t, db)

    srv := httptest.NewServer(buildRouter(db))
    defer srv.Close()

    // POST /api/v1/contacts
    body := `{"firstName":"Ada","lastName":"Lovelace","email":"ada@example.com"}`
    resp, err := http.Post(srv.URL+"/api/v1/contacts", "application/json", strings.NewReader(body))
    require.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
    // ... decode and assert response fields
}
```

### Tagging Convention

Integration tests are tagged with `//go:build integration` so they're skipped during fast unit test runs:

```go
//go:build integration
// +build integration

package handler_test
```

Run with: `go test -tags=integration ./...`

---

## 4. Frontend — Component Tests (Vitest + React Testing Library)

### Stack

- **Vitest** — fast, Vite-native test runner (replaces Jest in Vite projects)
- **React Testing Library (RTL)** — test behavior, not implementation
- **@testing-library/user-event** — realistic user interaction simulation
- **MSW (Mock Service Worker)** — intercept API calls at the network layer

### Philosophy

> Test what the user sees and does, not internal state or component structure.

```tsx
// src/components/ContactForm.test.tsx
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ContactForm } from './ContactForm'

describe('ContactForm', () => {
  it('submits with valid data', async () => {
    const onSubmit = vi.fn()
    render(<ContactForm onSubmit={onSubmit} />)

    await userEvent.type(screen.getByLabelText(/first name/i), 'Ada')
    await userEvent.type(screen.getByLabelText(/email/i), 'ada@example.com')
    await userEvent.click(screen.getByRole('button', { name: /save/i }))

    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({ firstName: 'Ada', email: 'ada@example.com' })
    )
  })

  it('shows validation errors for empty required fields', async () => {
    render(<ContactForm onSubmit={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: /save/i }))
    expect(screen.getByText(/first name is required/i)).toBeInTheDocument()
  })
})
```

### MSW API Mocking

```ts
// src/test/mocks/handlers.ts
import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v1/contacts', () =>
    HttpResponse.json({ data: [], total: 0, page: 1 })
  ),
  http.post('/api/v1/contacts', async ({ request }) => {
    const body = await request.json()
    return HttpResponse.json({ id: 'uuid-123', ...body }, { status: 201 })
  }),
]
```

```ts
// src/test/setup.ts
import { server } from './mocks/server'
beforeAll(() => server.listen())
afterEach(() => server.resetHandlers())
afterAll(() => server.close())
```

### Vitest Config

```ts
// vitest.config.ts
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      thresholds: { lines: 60, branches: 60, functions: 60 },
    },
  },
})
```

### File Conventions

- Co-located: `src/components/ContactCard.test.tsx`
- Shared test utilities: `src/test/utils.tsx` (custom render with providers)
- MSW handlers: `src/test/mocks/handlers.ts`

---

## 5. E2E Tests — Playwright

### Scope

E2E tests cover the 5 critical user flows only. Fast and focused—not a full regression suite.

**Critical flows:**

1. **Auth** — Login, token refresh, logout
2. **Contact CRUD** — Create → view → edit → delete a contact
3. **Deal pipeline** — Create deal, move through stages, close as won
4. **Account + Contact linkage** — Create account, add contact, verify relationship
5. **Search** — Global search returns correct results across entities

### Config

```ts
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,          // sequential for CRM state isolation
  retries: process.env.CI ? 2 : 0,
  use: {
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
  webServer: {
    command: 'docker compose up --wait',
    url: 'http://localhost:5173',
    reuseExistingServer: !process.env.CI,
    timeout: 60_000,
  },
})
```

### Seed & Teardown

```ts
// e2e/fixtures/auth.ts
import { test as base, expect } from '@playwright/test'

export const test = base.extend({
  authenticatedPage: async ({ page }, use) => {
    await page.goto('/login')
    await page.fill('[name=email]', 'admin@omnir.test')
    await page.fill('[name=password]', 'testpassword')
    await page.click('button[type=submit]')
    await expect(page).toHaveURL('/dashboard')
    await use(page)
  },
})
```

### Example E2E Test

```ts
// e2e/contacts.spec.ts
import { test, expect } from './fixtures/auth'

test('creates and deletes a contact', async ({ authenticatedPage: page }) => {
  await page.goto('/contacts/new')
  await page.fill('[name=firstName]', 'Ada')
  await page.fill('[name=lastName]', 'Lovelace')
  await page.fill('[name=email]', 'ada@omnir.test')
  await page.click('button[type=submit]')

  await expect(page.getByText('Ada Lovelace')).toBeVisible()

  await page.click('[aria-label="Delete contact"]')
  await page.click('button:has-text("Confirm")')
  await expect(page.getByText('Ada Lovelace')).not.toBeVisible()
})
```

---

## 6. CI Enforcement

```yaml
# .github/workflows/test.yml (reference)

jobs:
  go-unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: go test ./... -coverprofile=coverage.out
      - run: go tool cover -func=coverage.out | grep total | awk '{if ($3+0 < 70) exit 1}'

  go-integration:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: omnir
          POSTGRES_PASSWORD: omnir_dev
          POSTGRES_DB: omnir_crm_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: go test -tags=integration ./...
        env:
          DATABASE_URL: postgres://omnir:omnir_dev@localhost:5432/omnir_crm_test?sslmode=disable

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: '20' }
      - run: cd omnir-frontend && npm ci
      - run: cd omnir-frontend && npm run test:coverage

  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: docker compose up -d --wait
      - uses: actions/setup-node@v4
        with: { node-version: '20' }
      - run: cd omnir-frontend && npm ci && npx playwright install --with-deps chromium
      - run: cd omnir-frontend && npx playwright test
      - uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: playwright-report
          path: omnir-frontend/playwright-report/
```

---

## 7. Quick Reference

| Command | What it runs |
|---|---|
| `go test ./...` | All Go unit tests |
| `go test -tags=integration ./...` | Unit + integration |
| `npm run test` | Vitest watch |
| `npm run test:coverage` | Vitest with coverage report |
| `npx playwright test` | Full E2E suite |
| `npx playwright test --ui` | Playwright interactive UI |
| `docker compose -f docker-compose.yml -f docker-compose.test.yml up` | Full test stack |
