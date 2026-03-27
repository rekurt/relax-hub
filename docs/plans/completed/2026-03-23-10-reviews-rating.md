---
# Subsystem 10: Reviews & Rating Enhancements (FR-129-138)

## Overview
Bayesian average rating, multi-criteria ratings (cleanliness, accuracy, communication, value), auto review requests after check-out, quality monitoring with auto-depublish, and NLP auto-moderation placeholder.

## Context
- Existing reviews: `internal/service/review_service.go`, `internal/handler/review_handler.go`
- Review model: `internal/domain/review.go` — has Rating (float64), Text, UserID, BathhouseID
- Existing review media: photo/video attachments
- Existing moderation: complaint-based with auto-hide at 3+ reports
- Bathhouse has AvgRating and ReviewCount fields

## Dependencies
- No dependencies on other new subsystems
- Uses existing review, notification, and bathhouse subsystems

## Development Approach
- **Testing approach**: Regular (code first, then tests)
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 10.1: Multi-criteria Rating (FR-132)

**Files:**
- Modify: `internal/domain/review.go`
- Modify: `internal/service/review_service.go`
- Modify: `internal/handler/review_handler.go`
- Modify: `internal/repository/postgres/review_repo.go`
- Create: `migrations/XXXXXX_review_criteria.up.sql`

- [x] Add criteria fields to Review model:
  - Cleanliness (float64, 1.0-5.0, step 0.5)
  - Accuracy (float64, 1.0-5.0, step 0.5) — matches listing description
  - Communication (float64, 1.0-5.0, step 0.5) — owner responsiveness
  - ValueForMoney (float64, 1.0-5.0, step 0.5)
- [x] Overall Rating = average of 4 criteria (rounded to 1 decimal)
- [x] Migration: add `cleanliness DECIMAL(2,1)`, `accuracy DECIMAL(2,1)`, `communication DECIMAL(2,1)`, `value_for_money DECIMAL(2,1)` to reviews
- [x] Modify review creation: require all 4 criteria (or overall rating as before for backwards compat)
- [x] Modify review response: include criteria breakdown
- [x] Modify bathhouse detail: include average per criteria
  ```json
  {
    "avg_rating": 4.3,
    "avg_cleanliness": 4.5,
    "avg_accuracy": 4.2,
    "avg_communication": 4.4,
    "avg_value_for_money": 4.1,
    "review_count": 42
  }
  ```
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 10.2: Bayesian Average Rating (FR-136)

**Files:**
- Modify: `internal/service/review_service.go`
- Modify: `internal/repository/postgres/review_repo.go`
- Modify: `internal/domain/bathhouse.go`

- [x] Implement Bayesian average calculation:
  ```
  R_bayesian = (n * R + m * C) / (n + m)
  where:
    n = review count for this bathhouse
    R = simple average rating for this bathhouse
    m = minimum reviews threshold (configurable, default 5)
    C = platform-wide average rating
  ```
- [x] Display rules:
  - < 3 reviews: show "Новое" badge instead of rating
  - >= 3 reviews: show Bayesian average
- [x] Recalculate on every review create/update/delete:
  - Update bathhouse.AvgRating with Bayesian value
  - Recalculate platform average C (cache in Redis, recalc daily)
- [x] Add BayesianRating field to Bathhouse (or rename existing AvgRating)
- [x] Write tests with edge cases (0 reviews, 1 review, many reviews)
- [x] Run `go test ./... -v` — must pass

### Task 10.3: Auto Review Request (FR-129)

**Files:**
- Create: `internal/cron/review_jobs.go`
- Modify: `internal/service/review_service.go`

- [x] Cron job (hourly): find bookings where:
  - Status = completed
  - CheckedOutAt is not null
  - CheckedOutAt + 2 hours < now (configurable: BANI_REVIEW_REQUEST_DELAY_HOURS, default 2, range 2-24)
  - No review exists for this booking
  - No review request sent yet (track in Redis: `review_request_sent:{bookingID}`)
- [x] Send notification:
  - Push: "Как вам визит в {bathhouse_name}? Оставьте отзыв!"
  - Email: include link to review form
- [x] Include pre-filled data: booking ID, bathhouse name
- [x] Max 1 review request per booking
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 10.4: Quality Monitoring (FR-137, FR-138)

**Files:**
- Modify: `internal/service/review_service.go`
- Modify: `internal/service/bathhouse_service.go`
- Modify: `internal/domain/bathhouse.go`

- [x] On rating recalculation, check thresholds:
  - Rating < 3.0 (and >= 10 reviews): send warning to owner
    - "Рейтинг {bathhouse_name} опустился ниже 3.0. Обратите внимание на отзывы."
  - Rating < 2.0 (and >= 10 reviews): auto-depublish listing
    - Set status to "inactive"
    - Notify owner: "Объявление снято с публикации из-за низкого рейтинга"
    - Notify admin
- [x] Badges system:
  - "Verified" — all photos moderated + KYC approved
  - "Top" — Bayesian rating >= 4.5 AND review_count >= 10
  - "Premium" — active premium/promoted subscription
  - "New" — created within last 30 days AND < 3 reviews
- [x] Add Badges ([]string) to bathhouse response (computed, not stored)
- [x] Write tests
- [x] Run `go test ./... -v` — must pass

### Task 10.5: NLP Auto-moderation Placeholder (FR-133)

**Files:**
- Create: `internal/moderation/text_moderator.go`
- Modify: `internal/service/review_service.go`

- [x] Define TextModerationService interface:
  ```go
  type TextModerationService interface {
      Analyze(ctx context.Context, text string) (*ModerationResult, error)
  }
  type ModerationResult struct {
      Flagged  bool     `json:"flagged"`
      Score    float64  `json:"score"`    // 0.0 (clean) - 1.0 (definitely bad)
      Flags    []string `json:"flags"`    // "profanity", "spam", "contact_info", etc.
  }
  ```
- [x] Regex-based implementation (v1):
  - Profanity filter: common Russian profanity word list
  - URL detection: http/https/www patterns
  - Phone detection: Russian phone patterns (+7, 8-, etc.)
  - Email detection
  - Spam patterns: repeated characters, ALL CAPS > 50%
- [x] Integration with review creation:
  - If flagged (score > 0.7): set review status to "pending_moderation" instead of auto-publish
  - If clean: auto-publish as before
- [x] Store moderation result for admin reference
- [x] Future: replace regex with ML model after accumulating 10,000+ moderated texts
- [x] Write tests with various text samples
- [x] Run `go test ./... -v -race` — must pass
- [x] Run linter: `make lint`
