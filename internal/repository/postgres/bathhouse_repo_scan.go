package postgres

import (
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rekurt/relax-hub/internal/domain"
)

func (r *bathhouseRepo) scanBathhouseMinimal(rows pgx.Rows) (*domain.Bathhouse, error) {
	var (
		bh         domain.Bathhouse
		imagesJSON []byte
		whJSON     []byte
	)
	err := rows.Scan(
		&bh.ID, &bh.OwnerID, &bh.Name, &bh.Slug, &bh.Description, &bh.Address, &bh.CityID,
		&bh.Latitude, &bh.Longitude, &bh.PricePerHour, &bh.MinDuration, &bh.MaxGuests,
		&bh.HasPool, &bh.HasSauna, &bh.HasSteamRoom, &bh.HasHotTub, &bh.HasBBQ, &bh.HasKaraoke,
		&bh.Rating, &bh.BayesianRating, &bh.ReviewCount, &imagesJSON, &whJSON, &bh.Status,
		&bh.CreatedAt, &bh.UpdatedAt, &bh.IsPhotoVerified,
		&bh.LongSessionThresholdHours, &bh.LongSessionDiscountPercent, &bh.BaseCapacity, &bh.ExtraGuestSurcharge,
		&bh.LastMinuteEnabled, &bh.LastMinuteDiscountPercent, &bh.LastMinuteHoursThreshold,
		&bh.BufferMinutes, &bh.LeadTimeHours, &bh.MaxAdvanceDays,
		&bh.BookingMode, &bh.RequestTimeout,
		&bh.ResponseRate, &bh.AvgResponseTimeMinutes,
		&bh.CancellationPolicy, &bh.SecurityDepositPercent,
		&bh.LowResponseRateSince,
	)
	if err != nil {
		return nil, fmt.Errorf("scan bathhouse row: %w", err)
	}
	if err := json.Unmarshal(imagesJSON, &bh.Images); err != nil {
		return nil, fmt.Errorf("unmarshal images: %w", err)
	}
	if err := json.Unmarshal(whJSON, &bh.WorkingHours); err != nil {
		return nil, fmt.Errorf("unmarshal working hours: %w", err)
	}
	return &bh, nil
}

func (r *bathhouseRepo) scanBathhouseFromRowWithSubscription(rows pgx.Rows) (*domain.Bathhouse, error) {
	var (
		bh         domain.Bathhouse
		imagesJSON []byte
		whJSON     []byte
		isPromoted bool
		promoRank  *int64 // used for ORDER BY only, ignored in domain
	)
	err := rows.Scan(
		&bh.ID, &bh.OwnerID, &bh.Name, &bh.Slug, &bh.Description, &bh.Address, &bh.CityID,
		&bh.Latitude, &bh.Longitude, &bh.PricePerHour, &bh.MinDuration, &bh.MaxGuests,
		&bh.HasPool, &bh.HasSauna, &bh.HasSteamRoom, &bh.HasHotTub, &bh.HasBBQ, &bh.HasKaraoke,
		&bh.Rating, &bh.BayesianRating, &bh.ReviewCount, &imagesJSON, &whJSON, &bh.Status, &bh.CreatedAt, &bh.UpdatedAt,
		&bh.IsPhotoVerified,
		&bh.LongSessionThresholdHours, &bh.LongSessionDiscountPercent, &bh.BaseCapacity, &bh.ExtraGuestSurcharge,
		&bh.LastMinuteEnabled, &bh.LastMinuteDiscountPercent, &bh.LastMinuteHoursThreshold,
		&bh.BufferMinutes, &bh.LeadTimeHours, &bh.MaxAdvanceDays,
		&bh.BookingMode, &bh.RequestTimeout,
		&bh.ResponseRate, &bh.AvgResponseTimeMinutes,
		&bh.CancellationPolicy, &bh.SecurityDepositPercent,
		&isPromoted,
		&promoRank,
	)
	if err != nil {
		return nil, fmt.Errorf("scan bathhouse row: %w", err)
	}

	if err := json.Unmarshal(imagesJSON, &bh.Images); err != nil {
		return nil, fmt.Errorf("unmarshal images: %w", err)
	}
	if err := json.Unmarshal(whJSON, &bh.WorkingHours); err != nil {
		return nil, fmt.Errorf("unmarshal working hours: %w", err)
	}

	bh.IsPromoted = isPromoted

	return &bh, nil
}

func (r *bathhouseRepo) scanBathhouseFromRowWithAPIKey(rows pgx.Rows) (*domain.Bathhouse, error) {
	var (
		bh         domain.Bathhouse
		imagesJSON []byte
		whJSON     []byte
		isPromoted bool
	)
	err := rows.Scan(
		&bh.ID, &bh.OwnerID, &bh.Name, &bh.Slug, &bh.Description, &bh.Address, &bh.CityID,
		&bh.Latitude, &bh.Longitude, &bh.PricePerHour, &bh.MinDuration, &bh.MaxGuests,
		&bh.HasPool, &bh.HasSauna, &bh.HasSteamRoom, &bh.HasHotTub, &bh.HasBBQ, &bh.HasKaraoke,
		&bh.Rating, &bh.BayesianRating, &bh.ReviewCount, &imagesJSON, &whJSON, &bh.Status, &bh.CreatedAt, &bh.UpdatedAt,
		&bh.IsPhotoVerified, &bh.ApiKey,
		&bh.LongSessionThresholdHours, &bh.LongSessionDiscountPercent, &bh.BaseCapacity, &bh.ExtraGuestSurcharge,
		&bh.LastMinuteEnabled, &bh.LastMinuteDiscountPercent, &bh.LastMinuteHoursThreshold,
		&bh.BufferMinutes, &bh.LeadTimeHours, &bh.MaxAdvanceDays,
		&bh.BookingMode, &bh.RequestTimeout,
		&bh.ResponseRate, &bh.AvgResponseTimeMinutes,
		&bh.CancellationPolicy, &bh.SecurityDepositPercent,
		&isPromoted,
	)
	if err != nil {
		return nil, fmt.Errorf("scan bathhouse row: %w", err)
	}

	if err := json.Unmarshal(imagesJSON, &bh.Images); err != nil {
		return nil, fmt.Errorf("unmarshal images: %w", err)
	}
	if err := json.Unmarshal(whJSON, &bh.WorkingHours); err != nil {
		return nil, fmt.Errorf("unmarshal working hours: %w", err)
	}

	bh.IsPromoted = isPromoted

	return &bh, nil
}

func (r *bathhouseRepo) scanBathhouseFromRowWithAPIKeyOnly(rows pgx.Rows) (*domain.Bathhouse, error) {
	var (
		bh         domain.Bathhouse
		imagesJSON []byte
		whJSON     []byte
	)
	err := rows.Scan(
		&bh.ID, &bh.OwnerID, &bh.Name, &bh.Slug, &bh.Description, &bh.Address, &bh.CityID,
		&bh.Latitude, &bh.Longitude, &bh.PricePerHour, &bh.MinDuration, &bh.MaxGuests,
		&bh.HasPool, &bh.HasSauna, &bh.HasSteamRoom, &bh.HasHotTub, &bh.HasBBQ, &bh.HasKaraoke,
		&bh.Rating, &bh.BayesianRating, &bh.ReviewCount, &imagesJSON, &whJSON, &bh.Status, &bh.ApiKey, &bh.IsPhotoVerified,
		&bh.LongSessionThresholdHours, &bh.LongSessionDiscountPercent, &bh.BaseCapacity, &bh.ExtraGuestSurcharge,
		&bh.LastMinuteEnabled, &bh.LastMinuteDiscountPercent, &bh.LastMinuteHoursThreshold,
		&bh.BufferMinutes, &bh.LeadTimeHours, &bh.MaxAdvanceDays,
		&bh.BookingMode, &bh.RequestTimeout,
		&bh.ResponseRate, &bh.AvgResponseTimeMinutes,
		&bh.CancellationPolicy, &bh.SecurityDepositPercent,
		&bh.CreatedAt, &bh.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan bathhouse row: %w", err)
	}

	if err := json.Unmarshal(imagesJSON, &bh.Images); err != nil {
		return nil, fmt.Errorf("unmarshal images: %w", err)
	}
	if err := json.Unmarshal(whJSON, &bh.WorkingHours); err != nil {
		return nil, fmt.Errorf("unmarshal working hours: %w", err)
	}

	return &bh, nil
}
