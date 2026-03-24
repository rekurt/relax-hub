package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
	"github.com/redis/go-redis/v9"
)

func (cs *CronScheduler) handleAutoRejectTimedOutRequests() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting auto-reject timed out booking requests")

	rejected, err := cs.bookingSvc.AutoRejectTimedOutRequests(ctx)
	if err != nil {
		cs.logger.Error("Auto-reject timed out requests failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("Auto-reject timed out requests completed", "rejected", rejected, "duration", time.Since(start))
}

func (cs *CronScheduler) handleNoShowDetection() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting no-show detection")

	marked, err := cs.bookingSvc.MarkNoShows(ctx)
	if err != nil {
		cs.logger.Error("No-show detection failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("No-show detection completed", "marked", marked, "duration", time.Since(start))
}

func (cs *CronScheduler) handleBookingReminders() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting booking reminders")

	now := time.Now()
	sent := 0

	// 24h reminders: bookings starting between now+23h45m and now+24h15m
	reminder24hFrom := now.Add(23*time.Hour + 45*time.Minute)
	reminder24hTo := now.Add(24*time.Hour + 15*time.Minute)
	sent += cs.sendReminders(ctx, reminder24hFrom, reminder24hTo, "24h", func(info reminderInfo) {
		mapsURL := fmt.Sprintf("https://yandex.ru/maps/?pt=%.6f,%.6f&z=16", info.Longitude, info.Latitude)
		body := fmt.Sprintf("Завтра в %s — бронирование в %s. Адрес: %s",
			info.StartTime.Format("15:04"), info.BathhouseName, info.Address)
		data := map[string]string{
			"booking_id":   info.BookingID.String(),
			"bathhouse_id": info.BathhouseID.String(),
			"maps_url":     mapsURL,
		}
		if err := cs.notifSvc.Send(ctx, info.UserID, domain.NotifBookingReminder24h,
			"Напоминание о бронировании", body, data); err != nil {
			cs.logger.Warn("failed to send 24h reminder", "booking_id", info.BookingID, "error", err)
		}
	})

	// 2h reminders: bookings starting between now+1h45m and now+2h15m
	reminder2hFrom := now.Add(1*time.Hour + 45*time.Minute)
	reminder2hTo := now.Add(2*time.Hour + 15*time.Minute)
	sent += cs.sendReminders(ctx, reminder2hFrom, reminder2hTo, "2h", func(info reminderInfo) {
		body := fmt.Sprintf("Через 2 часа — %s. Не забудьте!", info.BathhouseName)
		data := map[string]string{
			"booking_id":   info.BookingID.String(),
			"bathhouse_id": info.BathhouseID.String(),
		}
		if err := cs.notifSvc.Send(ctx, info.UserID, domain.NotifBookingReminder2h,
			"Скоро бронирование", body, data); err != nil {
			cs.logger.Warn("failed to send 2h reminder", "booking_id", info.BookingID, "error", err)
		}
	})

	// Owner 5-min reminders: bookings starting between now-1m and now+5m15s
	reminderOwnerFrom := now.Add(-1 * time.Minute)
	reminderOwnerTo := now.Add(5*time.Minute + 15*time.Second)
	sent += cs.sendReminders(ctx, reminderOwnerFrom, reminderOwnerTo, "owner_5min", func(info reminderInfo) {
		body := fmt.Sprintf("Гость скоро прибудет в %s!", info.BathhouseName)
		if err := cs.notifSvc.Send(ctx, info.OwnerID, domain.NotifBookingReminderOwner5min,
			"Скоро прибытие гостя", body, nil); err != nil {
			cs.logger.Warn("failed to send owner 5min reminder", "booking_id", info.BookingID, "error", err)
		}
	})

	cs.logger.Info("Booking reminders completed", "sent", sent, "duration", time.Since(start))
}

func (cs *CronScheduler) handleResponseRateRecalculation() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	start := time.Now()
	cs.logger.Info("Starting response rate recalculation")

	updated, err := cs.bookingSvc.RecalculateResponseRates(ctx)
	if err != nil {
		cs.logger.Error("Response rate recalculation failed", "error", err, "duration", time.Since(start))
		return
	}

	cs.logger.Info("Response rate recalculation completed", "updated", updated, "duration", time.Since(start))
}

type reminderInfo struct {
	BookingID     uuid.UUID
	UserID        uuid.UUID
	OwnerID       uuid.UUID
	BathhouseID   uuid.UUID
	BathhouseName string
	Address       string
	Latitude      float64
	Longitude     float64
	StartTime     time.Time
}

func (cs *CronScheduler) sendReminders(ctx context.Context, from, to time.Time, reminderType string, sendFn func(info reminderInfo)) int {
	bookings, err := cs.bookingSvc.ListUpcomingWithBathhouse(ctx, from, to)
	if err != nil {
		cs.logger.Error("Failed to list upcoming bookings for reminders",
			"type", reminderType, "error", err)
		return 0
	}

	sent := 0
	for _, b := range bookings {
		redisKey := fmt.Sprintf("reminder_sent:%s:%s", b.Booking.ID.String(), reminderType)

		if cs.redisClient != nil {
			// Atomic check-and-set deduplication using SET with NX mode
			wasSet, err := cs.redisClient.SetArgs(ctx, redisKey, "1", redis.SetArgs{
				Mode: "NX",
				TTL:  48 * time.Hour,
			}).Result()
			if err != nil && err != redis.Nil {
				cs.logger.Warn("Redis SET NX failed for reminder dedup, sending anyway",
					"key", redisKey, "error", err)
			} else if wasSet != "OK" {
				continue // already sent
			}
		}

		sendFn(reminderInfo{
			BookingID:     b.Booking.ID,
			UserID:        b.Booking.UserID,
			OwnerID:       b.OwnerID,
			BathhouseID:   b.Booking.BathhouseID,
			BathhouseName: b.BathhouseName,
			Address:       b.Address,
			Latitude:      b.Latitude,
			Longitude:     b.Longitude,
			StartTime:     b.Booking.StartTime,
		})
		sent++
	}
	return sent
}
