# Testing Patterns

**Analysis Date:** 2026-04-17

## Test Framework

**Backend (Go):**
- Runner: `testing` package (Go standard library)
- Config: `.github/workflows/ci.yml` (GitHub Actions) - runs `go test ./... -race -coverprofile=coverage.out -covermode=atomic`
- Coverage tool: built-in `go tool cover`

**Frontend (React/TypeScript):**
- Runner: Vitest 4.1.0
- Config: `frontend/vitest.config.ts` (referenced in package.json scripts)
- Assertion library: Vitest (chai assertions + Testing Library matchers)
- Environment: jsdom (browser-like DOM)

**Run Commands:**
```bash
# Backend
go test ./... -v                            # all tests
go test ./internal/service/ -v -run TestBooking  # single test
make test                                   # shortcut
make test-hurl                              # hurl integration tests (requires running server)

# Frontend
cd frontend && npx vitest run               # all tests
cd frontend && npx vitest run Dashboard.test.tsx  # single file
cd frontend && npx vitest --ui              # UI mode (not used in CI)
```

## Test File Organization

**Location:**
- Backend: co-located with source (e.g., `internal/service/booking_service_test.go` next to `booking_service.go`)
- Frontend: `src/__tests__/` directory (separate from source)

**Naming:**
- Backend: `{source_name}_test.go`
- Frontend: `{ComponentName}.test.tsx` or `{moduleName}.test.ts`

**Structure:**
```
# Backend
internal/
├── service/
│   ├── booking_service.go
│   ├── booking_service_test.go        # tests for BookingService
│   ├── wallet_service.go
│   └── wallet_service_test.go
├── handler/
│   ├── booking_handler.go
│   └── booking_handler_test.go        # tests for HTTP handlers
└── repository/
    ├── postgres/                      # real implementations
    │   ├── booking_repo.go
    │   └── ...
    └── mock/                          # in-memory mocks
        ├── booking.go
        ├── addon_repo.go
        ├── wallet_repo.go
        └── ...                        # 40+ mock files

# Frontend
src/
├── __tests__/
│   ├── Dashboard.test.tsx
│   ├── auth-store.test.ts
│   ├── BroadcastCreate.test.tsx
│   └── ...
├── stores/
│   ├── auth.ts
│   └── bathhouse.ts
└── components/
    └── (components tested in __tests__/)
```

## Test Structure

**Backend Test Suite (Go):**
```go
package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rekurt/relax-hub/internal/domain"
	"github.com/rekurt/relax-hub/internal/logger"
	"github.com/rekurt/relax-hub/internal/repository/mock"
	"github.com/rekurt/relax-hub/internal/service"
)

// 1. Test environment setup (helper)
func newBookingService() (service.BookingService, *mock.BathhouseRepo, *mock.BookingRepo, ...) {
	bhRepo := mock.NewBathhouseRepo()
	bookingRepo := mock.NewBookingRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewBookingService(bookingRepo, bhRepo, ...)
	return svc, bhRepo, bookingRepo, ...
}

// 2. Table-driven test (optional for simpler tests)
func TestBookingService_Create_Success(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()
	clientID := uuid.New()
	bh := createBathhouse(t, bhRepo, ownerID)

	// 3. Arrange: Set up test data
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, now.Location())
	end := start.Add(2 * time.Hour)

	// 4. Act: Call the function under test
	result, err := svc.Create(context.Background(), clientID, service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     end,
		GuestCount:  5,
	})

	// 5. Assert: Verify results
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Booking.Status != domain.BookingPending {
		t.Errorf("status = %q, want %q", result.Booking.Status, domain.BookingPending)
	}
	if result.Booking.TotalPrice != bh.PricePerHour*2 {
		t.Errorf("totalPrice = %d, want %d", result.Booking.TotalPrice, bh.PricePerHour*2)
	}
}

// 6. Error case test
func TestBookingService_Create_InactiveBathhouse(t *testing.T) {
	svc, bhRepo, _, _, _, _, _, _ := newBookingService()
	ownerID := uuid.New()

	bh := &domain.Bathhouse{
		ID: uuid.New(), OwnerID: ownerID, Name: "Pending Bath",
		Address: "123 St", CityID: 1, PricePerHour: 5000,
		MinDuration: 1, MaxGuests: 10, Status: domain.BathhouseStatusPending,
	}
	_ = bhRepo.Create(context.Background(), bh)

	start := time.Now().Add(24 * time.Hour)
	_, err := svc.Create(context.Background(), uuid.New(), service.CreateBookingInput{
		BathhouseID: bh.ID,
		StartTime:   start,
		EndTime:     start.Add(2 * time.Hour),
		GuestCount:  5,
	})

	if !errors.Is(err, domain.ErrBathhouseNotActive) {
		t.Errorf("should fail for inactive bathhouse, got: %v", err)
	}
}
```

**Frontend Test Suite (Vitest + React Testing Library):**
```typescript
import { render, screen, fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import Dashboard from '@/pages/Dashboard'
import { useGetMyBathhousesIdAnalytics } from '@/api/generated/analytics/analytics'

// 1. Mock external dependencies
vi.mock('@/api/generated/analytics/analytics', () => ({
  useGetMyBathhousesIdAnalytics: vi.fn(),
}))

// 2. Helper for rendering with providers
function renderDashboard() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <Dashboard />
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

// 3. Test suite
describe('Dashboard', () => {
  // 4. Setup before each test
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // 5. Happy path test
  it('renders analytics data when loaded', () => {
    const mockData = {
      data: {
        data: {
          bookings: 42,
          revenue: 1500000,
          views: 1200,
          rating: 4.7,
        },
      },
    }
    vi.mocked(useGetMyBathhousesIdAnalytics).mockReturnValue({
      data: mockData.data.data,
      isLoading: false,
      error: null,
    } as any)

    renderDashboard()
    expect(screen.getByText('42')).toBeInTheDocument()
  })

  // 6. Error case test
  it('handles loading error gracefully', () => {
    vi.mocked(useGetMyBathhousesIdAnalytics).mockReturnValue({
      error: new Error('API failed'),
      isLoading: false,
      data: null,
    } as any)

    renderDashboard()
    expect(screen.getByText(/ошибка/i)).toBeInTheDocument()
  })

  // 7. Store mutation test
  it('updates auth state on login', () => {
    const user = { id: '1', email: 'test@test.com', role: 'owner' }
    useAuthStore.getState().setAuth('jwt-token', user)

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(true)
    expect(state.token).toBe('jwt-token')
    expect(state.user).toEqual(user)
  })
})
```

## Mocking

**Backend Mocking Framework:** Manual interface-based mocks in `internal/repository/mock/`

**Mock Implementation Pattern:**
- All mocks implement repository interfaces (e.g., `*AddOnRepo` implements `repository.AddOnRepository`)
- In-memory storage using maps and sync.RWMutex for thread safety
- Automatic ID generation and timestamp management

**Example Mock Structure (`internal/repository/mock/addon_repo.go`):**
```go
type AddOnRepo struct {
	mu            sync.RWMutex
	addons        map[uuid.UUID]*domain.AddOn
	bookingAddOns map[uuid.UUID]*domain.BookingAddOn
}

func NewAddOnRepo() *AddOnRepo {
	return &AddOnRepo{
		addons:        make(map[uuid.UUID]*domain.AddOn),
		bookingAddOns: make(map[uuid.UUID]*domain.BookingAddOn),
	}
}

var _ repository.AddOnRepository = (*AddOnRepo)(nil)  // Interface conformance check

func (r *AddOnRepo) Create(_ context.Context, addon *domain.AddOn) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// ... copy addon to map
}

func (r *AddOnRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.AddOn, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.addons[id]
	if !ok {
		return nil, domain.ErrAddOnNotFound
	}
	return a, nil
}
```

**Frontend Mocking (`vitest`):**
```typescript
// Mock API calls
vi.mock('@/api/generated/analytics/analytics', () => ({
  useGetMyBathhousesIdAnalytics: vi.fn(),
}))

// Mock localStorage
const mockLocalStorage = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value }),
    removeItem: vi.fn((key: string) => { delete store[key] }),
    clear: () => { store = {} },
  }
})()
Object.defineProperty(window, 'localStorage', { value: mockLocalStorage })

// Mock React Query data
const mockData = {
  data: { bookings: 42, revenue: 1500000 },
  isLoading: false,
  error: null,
}
vi.mocked(useGetMyBathhousesIdAnalytics).mockReturnValue(mockData as any)
```

**What to Mock:**
- Repository interfaces (mock storage)
- External API calls (via vi.mock)
- localStorage/sessionStorage
- HTTP clients
- Time-dependent functions (if needed, use vi.useFakeTimers)

**What NOT to Mock:**
- Domain types (use real instances)
- Business logic (service layer should be tested)
- Error types and validation
- Helper functions (format.ts, constants)

**Handler Testing Pattern (`internal/handler/*_test.go`):**
```go
// Mock service interface
type mockDisputeService struct {
	openDisputeFn    func(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, reason domain.DisputeReason, description string) (*domain.Dispute, error)
	// ... other methods
}

func (m *mockDisputeService) OpenDispute(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, reason domain.DisputeReason, description string) (*domain.Dispute, error) {
	if m.openDisputeFn != nil {
		return m.openDisputeFn(ctx, userID, bookingID, reason, description)
	}
	return nil, nil
}

// Use in test
func TestDisputeHandler_OpenDispute(t *testing.T) {
	mockSvc := &mockDisputeService{
		openDisputeFn: func(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID, reason domain.DisputeReason, description string) (*domain.Dispute, error) {
			return &domain.Dispute{ID: uuid.New(), Status: domain.DisputeOpen}, nil
		},
	}
	handler := handler.NewDisputeHandler(mockSvc)

	req := httptest.NewRequest("POST", "/disputes", nil)
	w := httptest.NewRecorder()

	handler.OpenDispute(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}
```

## Fixtures and Factories

**Test Data Helpers (`internal/service/*_test.go`):**
```go
// Factory function
func createBathhouse(t *testing.T, bhRepo *mock.BathhouseRepo, ownerID uuid.UUID) *domain.Bathhouse {
	t.Helper()
	bh := &domain.Bathhouse{
		ID:           uuid.New(),
		OwnerID:      ownerID,
		Name:         "Test Bath",
		Address:      "123 St",
		CityID:       1,
		PricePerHour: 10000, // 100 RUB
		MinDuration: 1,
		MaxGuests:   10,
		Status:       domain.BathhouseStatusActive,
	}
	if err := bhRepo.Create(context.Background(), bh); err != nil {
		t.Fatalf("create bathhouse: %v", err)
	}
	return bh
}

// Environment setup
func newWalletTestEnv() *walletTestEnv {
	walletRepo := mock.NewWalletRepo()
	log := logger.New(logger.LevelWarn)
	svc := service.NewWalletService(walletRepo, log)
	return &walletTestEnv{svc: svc}
}

// Reusable test wallet
func createTestWallet(t *testing.T, env *walletTestEnv) *domain.Wallet {
	t.Helper()
	userID := uuid.New()
	wallet, err := env.svc.CreateWallet(context.Background(), userID, domain.WalletCurrencyRUB)
	if err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	return wallet
}
```

**Location:** Helpers defined in same test file or in separate `_test.go` files per package.

## Coverage

**Requirements:**
- Target: 80% (noted in CI script comments)
- Baseline: ~45-50% (due to postgres repos requiring real DB, not tested)
- Enforce: CI fails if coverage < 45% (see `.github/workflows/ci.yml`)

**View Coverage:**
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out   # opens in browser
go tool cover -func=coverage.out   # shows per-function coverage
```

**Coverage Gaps:**
- PostgreSQL repository implementations (require real database, excluded from coverage baseline)
- Integration tests (hurl tests in `tests/hurl/` run separately via `make test-hurl`)
- Third-party integrations (payment providers, PMS systems)

## Test Types

**Unit Tests:**
- Scope: Single service method or handler endpoint
- Dependencies: Mocked repositories, mocked external services
- Pattern: Arrange-Act-Assert with clear variable names
- Location: `internal/service/*_test.go`, `internal/handler/*_test.go`
- Examples: `TestBookingService_Create_Success`, `TestWalletService_CreateWallet_Duplicate`

**Integration Tests (Hurl):**
- Scope: Full HTTP endpoint request/response cycles
- Setup: Requires running server (started via `make docker-up`)
- Format: `.hurl` files in `tests/hurl/` (HTTP request syntax)
- Run: `make test-hurl`
- Purpose: Test auth flow, complex state transitions, edge cases

**E2E Tests:**
- Status: Not currently implemented
- Frontend: Could use Playwright or Cypress (not in codebase)
- No E2E test framework configured

## Common Patterns

**Async Testing (Go):**
```go
func TestAsyncOperation(t *testing.T) {
	done := make(chan error, 1)
	go func() {
		err := svc.AsyncWork(context.Background())
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("async work failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("async work timed out")
	}
}
```

**Error Testing (Go):**
```go
func TestError_Condition(t *testing.T) {
	_, err := svc.Operation(context.Background())
	
	if !errors.Is(err, domain.ErrSpecificError) {
		t.Errorf("expected ErrSpecificError, got: %v", err)
	}
}
```

**React Component Testing:**
```typescript
it('submits form with validation', async () => {
	render(<LoginForm />)
	
	// Act
	const submitBtn = screen.getByRole('button', { name: /submit/i })
	fireEvent.click(submitBtn)
	
	// Assert
	await waitFor(() => {
		expect(screen.getByText(/email required/i)).toBeInTheDocument()
	})
})

it('calls API on successful submission', async () => {
	const mockSubmit = vi.fn()
	render(<LoginForm onSubmit={mockSubmit} />)
	
	// ... fill and submit form
	
	await waitFor(() => {
		expect(mockSubmit).toHaveBeenCalledWith(expect.objectContaining({ email: 'test@test.com' }))
	})
})
```

**Store Testing (Zustand):**
```typescript
describe('useAuthStore', () => {
	beforeEach(() => {
		mockLocalStorage.clear()
		useAuthStore.setState({
			user: null,
			token: null,
			isLoading: false,
			isAuthenticated: false,
		})
	})

	it('initial state is unauthenticated', () => {
		const state = useAuthStore.getState()
		expect(state.isAuthenticated).toBe(false)
		expect(state.user).toBeNull()
	})

	it('setAuth stores token and user', () => {
		const user = { id: '1', email: 'test@test.com', role: 'owner' }
		useAuthStore.getState().setAuth('jwt-token', user)

		const state = useAuthStore.getState()
		expect(state.isAuthenticated).toBe(true)
		expect(state.token).toBe('jwt-token')
		expect(mockLocalStorage.setItem).toHaveBeenCalledWith('bani_token', 'jwt-token')
	})
})
```

## Best Practices

1. **Use helper functions** to set up common test environments (e.g., `newBookingService()`, `renderDashboard()`)
2. **Name tests descriptively:** `Test{Package}_{Function}_{Condition}` or `it('{behavior}')`
3. **Isolate tests:** Use setup/teardown (beforeEach) to reset state and clear mocks
4. **Mock at boundaries:** Mock repositories, APIs, time.Now() if needed; not business logic
5. **Test error cases:** Ensure error handling works correctly (e.g., `TestBookingService_Create_InactiveBathhouse`)
6. **Use context.Background()** for testing (no timeout unless testing timeout behavior)
7. **Keep mocks simple:** Don't replicate entire database logic; just track state and calls
8. **Test behavior, not implementation:** Focus on inputs/outputs, not internal state
9. **Use type assertions for mocks:** Verify interface compliance with `var _ Interface = (*Mock)(nil)`
10. **Avoid flakiness:** Don't rely on timing; use deterministic test data and time.Now() sparingly

---

*Testing analysis: 2026-04-17*
