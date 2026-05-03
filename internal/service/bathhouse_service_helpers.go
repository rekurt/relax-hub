package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/rekurt/relax-hub/internal/domain"
)

func buildChangedFields(old, new *domain.Bathhouse) map[string]interface{} {
	changes := make(map[string]interface{})

	if old.Name != new.Name {
		changes["name"] = map[string]string{"old": old.Name, "new": new.Name}
	}
	if old.Description != new.Description {
		changes["description"] = map[string]string{"old": old.Description, "new": new.Description}
	}
	if old.Address != new.Address {
		changes["address"] = map[string]string{"old": old.Address, "new": new.Address}
	}
	if old.CityID != new.CityID {
		changes["city_id"] = map[string]int64{"old": old.CityID, "new": new.CityID}
	}
	if old.Latitude != new.Latitude {
		changes["latitude"] = map[string]float64{"old": old.Latitude, "new": new.Latitude}
	}
	if old.Longitude != new.Longitude {
		changes["longitude"] = map[string]float64{"old": old.Longitude, "new": new.Longitude}
	}
	if old.PricePerHour != new.PricePerHour {
		changes["price_per_hour"] = map[string]int64{"old": old.PricePerHour, "new": new.PricePerHour}
	}
	if old.MinDuration != new.MinDuration {
		changes["min_duration"] = map[string]int{"old": old.MinDuration, "new": new.MinDuration}
	}
	if old.MaxGuests != new.MaxGuests {
		changes["max_guests"] = map[string]int{"old": old.MaxGuests, "new": new.MaxGuests}
	}
	if old.HasPool != new.HasPool {
		changes["has_pool"] = map[string]bool{"old": old.HasPool, "new": new.HasPool}
	}
	if old.HasSauna != new.HasSauna {
		changes["has_sauna"] = map[string]bool{"old": old.HasSauna, "new": new.HasSauna}
	}
	if old.HasSteamRoom != new.HasSteamRoom {
		changes["has_steam_room"] = map[string]bool{"old": old.HasSteamRoom, "new": new.HasSteamRoom}
	}
	if old.HasHotTub != new.HasHotTub {
		changes["has_hot_tub"] = map[string]bool{"old": old.HasHotTub, "new": new.HasHotTub}
	}
	if old.HasBBQ != new.HasBBQ {
		changes["has_bbq"] = map[string]bool{"old": old.HasBBQ, "new": new.HasBBQ}
	}
	if old.HasKaraoke != new.HasKaraoke {
		changes["has_karaoke"] = map[string]bool{"old": old.HasKaraoke, "new": new.HasKaraoke}
	}
	if !stringSlicesEqual(old.Images, new.Images) {
		changes["images"] = map[string]interface{}{"old": old.Images, "new": new.Images}
	}
	if !workingHoursEqual(old.WorkingHours, new.WorkingHours) {
		oldWH, _ := json.Marshal(old.WorkingHours)
		newWH, _ := json.Marshal(new.WorkingHours)
		changes["working_hours"] = map[string]string{"old": string(oldWH), "new": string(newWH)}
	}

	return changes
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func workingHoursEqual(a, b []domain.WorkingHours) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

const (
	badgeVerified   = "verified"
	badgeTop        = "top"
	badgePremium    = "premium"
	badgeNew        = "new"
	badgeLastMinute = "last_minute"

	badgeTopMinRating  = 4.5
	badgeTopMinReviews = 10
	badgeNewMaxDays    = 30
	badgeNewMaxReviews = 3
)

func (s *bathhouseService) ComputeBadges(ctx context.Context, bh *domain.Bathhouse) []string {
	var badges []string

	// "Verified": photos moderated + KYC approved
	if bh.IsPhotoVerified {
		if approved, err := s.kycSvc.IsApproved(ctx, bh.OwnerID); err == nil && approved {
			badges = append(badges, badgeVerified)
		}
	}

	// "Top": Bayesian >= 4.5 AND review count >= 10
	if bh.BayesianRating >= badgeTopMinRating && bh.ReviewCount >= badgeTopMinReviews {
		badges = append(badges, badgeTop)
	}

	// "Premium": active subscription (premium or promoted plan)
	if s.subRepo != nil {
		sub, err := s.subRepo.GetActiveBybathhouse(ctx, bh.ID)
		if err == nil && sub != nil && (sub.Plan == domain.PlanPremium || sub.Plan == domain.PlanPromoted) {
			badges = append(badges, badgePremium)
		}
	}

	// "New": < 30 days old AND < 3 reviews
	if time.Since(bh.CreatedAt) < badgeNewMaxDays*24*time.Hour && bh.ReviewCount < badgeNewMaxReviews {
		badges = append(badges, badgeNew)
	}

	// "LastMinute": last-minute discount is currently active
	if s.IsLastMinuteActive(bh) {
		badges = append(badges, badgeLastMinute)
	}

	return badges
}

// IsLastMinuteActive checks whether a bathhouse currently qualifies for its last-minute discount.
// Returns true when the bathhouse has last-minute enabled and there are open working hours
// starting within the threshold window from now.
func (s *bathhouseService) IsLastMinuteActive(bh *domain.Bathhouse) bool {
	if !bh.LastMinuteEnabled || bh.LastMinuteHoursThreshold <= 0 {
		return false
	}

	now := time.Now()
	threshold := now.Add(time.Duration(bh.LastMinuteHoursThreshold) * time.Hour)

	for dayOffset := 0; dayOffset <= 1; dayOffset++ {
		checkDate := now.AddDate(0, 0, dayOffset)
		wd := checkDate.Weekday()
		dayOfWeek := int(wd) - 1
		if wd == time.Sunday {
			dayOfWeek = 6
		}

		for _, wh := range bh.WorkingHours {
			if wh.DayOfWeek != dayOfWeek {
				continue
			}
			if len(wh.OpenTime) != 5 {
				continue
			}
			h := (int(wh.OpenTime[0]-'0') * 10) + int(wh.OpenTime[1]-'0')
			m := (int(wh.OpenTime[3]-'0') * 10) + int(wh.OpenTime[4]-'0')
			slotStart := time.Date(checkDate.Year(), checkDate.Month(), checkDate.Day(), h, m, 0, 0, now.Location())

			if slotStart.After(now) && !slotStart.After(threshold) {
				return true
			}
		}
	}
	return false
}

// GetAreaAvgPrice returns the average price per hour for active bathhouses in the same city
// within a 5km radius of the given coordinates. Returns 0 if no data available.
// Results are cached in Redis for 1 hour by city_id.
func (s *bathhouseService) GetAreaAvgPrice(ctx context.Context, cityID int64, lat, lng float64) (int64, error) {
	cacheKey := fmt.Sprintf("%s%d", areaAvgPriceCachePrefix, cityID)

	if s.redisClient != nil {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			if v, parseErr := strconv.ParseInt(cached, 10, 64); parseErr == nil {
				return v, nil
			}
		}
	}

	avg, err := s.bhRepo.GetAreaAvgPrice(ctx, cityID, lat, lng)
	if err != nil {
		return 0, err
	}

	if s.redisClient != nil {
		if cacheErr := s.redisClient.Set(ctx, cacheKey, strconv.FormatInt(avg, 10), areaAvgPriceCacheTTL).Err(); cacheErr != nil {
			s.logger.Error("failed to cache area avg price", "city_id", cityID, "error", cacheErr)
		}
	}

	return avg, nil
}

func countAmenities(bh *domain.Bathhouse) int {
	count := 0
	if bh.HasPool {
		count++
	}
	if bh.HasSauna {
		count++
	}
	if bh.HasSteamRoom {
		count++
	}
	if bh.HasHotTub {
		count++
	}
	if bh.HasBBQ {
		count++
	}
	if bh.HasKaraoke {
		count++
	}
	return count
}
