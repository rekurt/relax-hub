package pms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rekurt/relax-hub/internal/domain"
)

const (
	restoplaceBaseURL    = "https://api.restoplace.ws/api/v1"
	restoplaceAPITimeout = 30 * time.Second
)

// RestoplaceCredentials holds the API key for Restoplace.
type RestoplaceCredentials struct {
	APIKey string `json:"api_key"`
}

// restoplaceProvider implements PMSProvider for Restoplace integration.
type restoplaceProvider struct {
	httpClient *http.Client
}

// NewRestoplaceProvider creates a new Restoplace PMS provider.
func NewRestoplaceProvider() PMSProvider {
	return &restoplaceProvider{
		httpClient: &http.Client{
			Timeout: restoplaceAPITimeout,
		},
	}
}

func (rp *restoplaceProvider) Name() domain.PMSProvider {
	return domain.PMSProviderRestoplace
}

func (rp *restoplaceProvider) TestConnection(ctx context.Context, credentials string) error {
	creds, err := parseRestoplaceCredentials(credentials)
	if err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, restoplaceBaseURL+"/venues", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	rp.setHeaders(req, creds)

	resp, err := rp.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("restoplace api request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("restoplace authentication failed: status %d", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("restoplace api error: status %d", resp.StatusCode)
	}

	return nil
}

func (rp *restoplaceProvider) PullBookings(ctx context.Context, credentials string, externalID string, from, to time.Time) ([]domain.PMSBooking, error) {
	creds, err := parseRestoplaceCredentials(credentials)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	url := fmt.Sprintf("%s/venues/%s/reservations?date_from=%s&date_to=%s",
		restoplaceBaseURL, externalID,
		from.Format("2006-01-02"), to.Format("2006-01-02"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	rp.setHeaders(req, creds)

	resp, err := rp.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("restoplace api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("restoplace api error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp restoplaceReservationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	bookings := make([]domain.PMSBooking, 0, len(apiResp.Reservations))
	for _, r := range apiResp.Reservations {
		startTime, _ := time.Parse("2006-01-02T15:04:05Z", r.StartAt)
		endTime, _ := time.Parse("2006-01-02T15:04:05Z", r.EndAt)

		bookings = append(bookings, domain.PMSBooking{
			ExternalID: r.ID,
			StartTime:  startTime,
			EndTime:    endTime,
			GuestName:  r.GuestName,
			GuestPhone: r.GuestPhone,
			Status:     r.Status,
			Notes:      r.Notes,
		})
	}

	return bookings, nil
}

func (rp *restoplaceProvider) PushBooking(ctx context.Context, credentials string, externalID string, booking domain.PMSBooking) error {
	creds, err := parseRestoplaceCredentials(credentials)
	if err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	payload := restoplaceCreateReservation{
		StartAt:    booking.StartTime.Format("2006-01-02T15:04:05Z"),
		EndAt:      booking.EndTime.Format("2006-01-02T15:04:05Z"),
		GuestName:  booking.GuestName,
		GuestPhone: booking.GuestPhone,
		Notes:      booking.Notes,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/venues/%s/reservations", restoplaceBaseURL, externalID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	rp.setHeaders(req, creds)
	req.Header.Set("Content-Type", "application/json")

	resp, err := rp.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("restoplace api request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("restoplace push booking failed: status %d", resp.StatusCode)
	}

	return nil
}

func (rp *restoplaceProvider) SyncSchedule(ctx context.Context, credentials string, externalID string, from, to time.Time) ([]domain.PMSScheduleSlot, error) {
	creds, err := parseRestoplaceCredentials(credentials)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	url := fmt.Sprintf("%s/venues/%s/availability?date_from=%s&date_to=%s",
		restoplaceBaseURL, externalID,
		from.Format("2006-01-02"), to.Format("2006-01-02"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	rp.setHeaders(req, creds)

	resp, err := rp.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("restoplace api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("restoplace schedule api error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp restoplaceAvailabilityResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode schedule response: %w", err)
	}

	var slots []domain.PMSScheduleSlot
	for _, day := range apiResp.Days {
		for _, slot := range day.Slots {
			slots = append(slots, domain.PMSScheduleSlot{
				Date:      day.Date,
				TimeFrom:  slot.From,
				TimeTo:    slot.To,
				Available: slot.Available,
			})
		}
	}

	return slots, nil
}

func (rp *restoplaceProvider) CancelBooking(ctx context.Context, credentials string, externalID string, bookingExternalID string) error {
	creds, err := parseRestoplaceCredentials(credentials)
	if err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	url := fmt.Sprintf("%s/venues/%s/reservations/%s/cancel", restoplaceBaseURL, externalID, bookingExternalID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	rp.setHeaders(req, creds)

	resp, err := rp.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("restoplace api request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("restoplace cancel booking failed: status %d", resp.StatusCode)
	}

	return nil
}

func (rp *restoplaceProvider) setHeaders(req *http.Request, creds RestoplaceCredentials) {
	req.Header.Set("X-API-Key", creds.APIKey)
	req.Header.Set("Accept", "application/json")
}

func parseRestoplaceCredentials(raw string) (RestoplaceCredentials, error) {
	var creds RestoplaceCredentials
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return creds, fmt.Errorf("parse restoplace credentials: %w", err)
	}
	if creds.APIKey == "" {
		return creds, fmt.Errorf("missing required restoplace api_key")
	}
	return creds, nil
}

// Restoplace API response types

type restoplaceReservationsResponse struct {
	Reservations []restoplaceReservation `json:"reservations"`
}

type restoplaceReservation struct {
	ID         string `json:"id"`
	StartAt    string `json:"start_at"`
	EndAt      string `json:"end_at"`
	GuestName  string `json:"guest_name"`
	GuestPhone string `json:"guest_phone"`
	Status     string `json:"status"`
	Notes      string `json:"notes"`
}

type restoplaceCreateReservation struct {
	StartAt    string `json:"start_at"`
	EndAt      string `json:"end_at"`
	GuestName  string `json:"guest_name"`
	GuestPhone string `json:"guest_phone"`
	Notes      string `json:"notes"`
}

type restoplaceAvailabilityResponse struct {
	Days []restoplaceAvailabilityDay `json:"days"`
}

type restoplaceAvailabilityDay struct {
	Date  string                       `json:"date"`
	Slots []restoplaceAvailabilitySlot `json:"slots"`
}

type restoplaceAvailabilitySlot struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Available bool   `json:"available"`
}
