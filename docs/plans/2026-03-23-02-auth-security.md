---
# Subsystem 2: Authentication & Security (FR-001, FR-007, FR-008, FR-014, FR-015)

## Overview
Phone+OTP authentication, two-factor authentication (TOTP/SMS), session management, password reset, account deletion with grace period, and age verification with welcome bonus.

## Context
- Existing auth: email + password login, OAuth (VK, Yandex, Google) via `internal/service/auth_service.go`
- JWT-based auth middleware: `internal/middleware/auth.go`
- User model: `internal/domain/user.go` — has Email, Name, Role, PasswordHash, social accounts
- Auth handler: `internal/handler/auth_handler.go`
- Redis available for OTP storage, rate limiting
- Config prefix: BANI_

## Dependencies
- No dependencies on other new subsystems
- Wallet system should be called on registration for welcome bonus (but can be added later via hook)

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 2.1: Phone + OTP Authentication (FR-001)

**Files:**
- Modify: `internal/domain/user.go`
- Create: `internal/service/otp_service.go`
- Modify: `internal/service/auth_service.go`
- Modify: `internal/handler/auth_handler.go`
- Create: `migrations/XXXXXX_phone_auth.up.sql`
- Create: `migrations/XXXXXX_phone_auth.down.sql`

- [ ] Add Phone (string) and PhoneVerified (bool) fields to User model
- [ ] Migration: add `phone VARCHAR(20)`, `phone_verified BOOLEAN DEFAULT false` columns to users table, add unique index on phone where phone is not null
- [ ] Create OTPService interface and implementation:
  ```go
  type OTPService interface {
      SendOTP(ctx context.Context, phone string) error
      VerifyOTP(ctx context.Context, phone string, code string) (bool, error)
  }
  ```
- [ ] OTP storage: Redis key `otp:{phone}` with TTL 5 min, value = `{code, attempts}`
- [ ] OTP generation: 6-digit numeric code
- [ ] Max 3 verification attempts per OTP (after that, must request new code)
- [ ] Rate limit: max 3 OTP requests per 15 min per phone (Redis counter `otp_rate:{phone}`)
- [ ] SMS provider integration: define SMSProvider interface, implement SMS.ru adapter
  ```go
  type SMSProvider interface {
      SendSMS(ctx context.Context, phone string, message string) error
  }
  ```
- [ ] Config: BANI_SMS_PROVIDER (default "smsru"), BANI_SMS_API_KEY
- [ ] POST /api/v1/auth/register-phone — register with phone number
  - Request: { phone, name }
  - Sends OTP, returns { message: "OTP sent" }
- [ ] POST /api/v1/auth/login-phone — login with phone
  - Request: { phone }
  - Sends OTP, returns { message: "OTP sent" }
- [ ] POST /api/v1/auth/verify-phone — verify OTP code and complete login/registration
  - Request: { phone, code }
  - Response: { token, user } on success
- [ ] Write tests for OTPService, auth handler changes
- [ ] Run `go test ./... -v` — must pass

### Task 2.2: Two-Factor Authentication (FR-007)

**Files:**
- Create: `internal/service/twofa_service.go`
- Modify: `internal/handler/auth_handler.go`
- Modify: `internal/domain/user.go`
- Create: `migrations/XXXXXX_two_factor_auth.up.sql`
- Create: `migrations/XXXXXX_two_factor_auth.down.sql`

- [ ] Add to User model: TOTPSecret (string, encrypted), TwoFAMethod (enum: none/totp/sms)
- [ ] Migration: add `totp_secret TEXT`, `two_fa_method VARCHAR(10) DEFAULT 'none'` to users
- [ ] TwoFAService implementation using `pquerna/otp` library:
  ```go
  type TwoFAService interface {
      GenerateTOTPSecret(ctx context.Context, userID uuid.UUID) (secret string, qrURL string, err error)
      EnableTOTP(ctx context.Context, userID uuid.UUID, code string) error
      DisableTOTP(ctx context.Context, userID uuid.UUID, code string) error
      VerifyTOTP(ctx context.Context, userID uuid.UUID, code string) (bool, error)
      EnableSMS2FA(ctx context.Context, userID uuid.UUID) error
      SendSMS2FA(ctx context.Context, userID uuid.UUID) error
      VerifySMS2FA(ctx context.Context, userID uuid.UUID, code string) (bool, error)
  }
  ```
- [ ] POST /api/v1/auth/2fa/totp/enable — generate secret + return QR code URL
- [ ] POST /api/v1/auth/2fa/totp/verify — verify code and activate TOTP
- [ ] DELETE /api/v1/auth/2fa/totp — disable TOTP (requires current code in body)
- [ ] POST /api/v1/auth/2fa/sms/enable — enable SMS-based 2FA (requires verified phone)
- [ ] Modify login flow: if 2FA enabled, return partial_token (short-lived, 5 min) + `requires_2fa: true`
- [ ] POST /api/v1/auth/2fa/verify — verify 2FA code during login, return full JWT
- [ ] Add swagger annotations
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 2.3: Session Management (FR-008)

**Files:**
- Create: `internal/domain/session.go`
- Create: `internal/service/session_service.go`
- Create: `internal/handler/session_handler.go`
- Create: `internal/repository/postgres/session_repo.go`
- Create: `internal/repository/mock/session_repo.go`
- Modify: `internal/repository/interfaces.go`
- Modify: `internal/middleware/auth.go`
- Create: `migrations/XXXXXX_sessions.up.sql`

- [ ] Session model: ID (uuid), UserID, DeviceInfo (string), Browser (string), IP (string), LastActiveAt, CreatedAt, ExpiresAt
- [ ] SessionRepository interface: Create, GetByID, ListByUser, Delete, DeleteAllExcept, UpdateLastActive, DeleteExpired
- [ ] SessionService:
  - CreateSession(ctx, userID, deviceInfo, browser, ip) — create session record, embed session ID in JWT
  - ListSessions(ctx, userID) — return all active sessions
  - TerminateSession(ctx, userID, sessionID) — delete specific session
  - TerminateAllExceptCurrent(ctx, userID, currentSessionID) — delete all other sessions
  - CleanExpired(ctx) — cron job to clean expired sessions
- [ ] Modify auth middleware: validate session exists and not expired, update LastActiveAt
- [ ] GET /api/v1/my/sessions — list all active sessions
- [ ] DELETE /api/v1/my/sessions/{id} — terminate specific session
- [ ] DELETE /api/v1/my/sessions — terminate all except current (current session ID from JWT)
- [ ] Auto-expire after 30 days of inactivity (LastActiveAt + 30 days)
- [ ] Migration: sessions table with indexes on user_id and expires_at
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 2.4: Password Reset (FR-015)

**Files:**
- Modify: `internal/service/auth_service.go`
- Modify: `internal/handler/auth_handler.go`

- [ ] POST /api/v1/auth/forgot-password — send reset link via email
  - Request: { email }
  - Generate secure token (32 bytes, hex encoded), store in Redis with key `password_reset:{token}` and value userID, TTL 30 min
  - Send email with reset link (BANI_FRONTEND_URL + "/reset-password?token=" + token)
  - Rate limit: max 3 requests per 15 min per email (Redis counter)
- [ ] POST /api/v1/auth/reset-password — reset with token
  - Request: { token, new_password }
  - Validate token exists in Redis, get userID
  - Update password hash
  - Delete token from Redis
  - Terminate all active sessions for this user
  - Return success
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 2.5: Account Deletion (FR-014)

**Files:**
- Create: `internal/service/account_deletion_service.go`
- Modify: `internal/handler/auth_handler.go`
- Modify: `internal/domain/user.go`
- Create: `migrations/XXXXXX_account_deletion.up.sql`

- [ ] Add to User model: DeletionRequestedAt (*time.Time), DeletionScheduledAt (*time.Time)
- [ ] Migration: add `deletion_requested_at TIMESTAMPTZ`, `deletion_scheduled_at TIMESTAMPTZ` to users
- [ ] AccountDeletionService:
  - RequestDeletion(ctx, userID) — set DeletionRequestedAt=now, DeletionScheduledAt=now+30days, send confirmation notification
  - RestoreAccount(ctx, userID) — clear deletion fields (only during grace period)
  - ExecuteDeletion(ctx, userID) — anonymize personal data (name="Deleted User", email=hash, phone=null), return topup funds to original card, burn bonus funds, keep financial history for 5 years
  - CheckPendingDeletions(ctx) — cron: find users where deletion_scheduled_at <= now
- [ ] POST /api/v1/auth/delete-account — request deletion (RequireAuth)
- [ ] POST /api/v1/auth/restore-account — restore during grace period (RequireAuth)
- [ ] Cron reminders: day 0 (confirmation), day 14 (reminder), day 27 (final warning)
- [ ] Cron: execute deletion at day 30
- [ ] Block login for accounts pending deletion (show message with restore option)
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 2.6: Age Verification & Welcome Bonus (FR-005, FR-016)

**Files:**
- Modify: `internal/service/auth_service.go`
- Modify: `internal/handler/auth_handler.go`
- Modify: `internal/domain/user.go`

- [ ] Add AgeConfirmed (bool) to registration request validation
- [ ] Reject registration if age_confirmed != true (return ErrInvalidInput: "Age confirmation required")
- [ ] On first registration (any method: email, phone, OAuth):
  - Auto-create wallet (call WalletService.CreateWallet)
  - Credit welcome bonus: 500 RUB (50,000 kopecks), expires in 30 days
  - Config: BANI_WELCOME_BONUS_AMOUNT (default 50000), BANI_WELCOME_BONUS_EXPIRY_DAYS (default 30)
- [ ] Write tests
- [ ] Run `go test ./... -v` — must pass

### Task 2.7: Verify & Finalize

- [ ] Run full test suite: `go test ./... -v -race`
- [ ] Run linter: `make lint`
- [ ] Run `go build ./...`
- [ ] Regenerate swagger: `make swagger`
