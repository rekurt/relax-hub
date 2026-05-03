package pms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rekurt/relax-hub/internal/domain"
)

const (
	yclientsBaseURL    = "https://api.yclients.com/api/v1"
	yclientsAPITimeout = 30 * time.Second
)

// YclientsCredentials holds the OAuth2 token and partner token for Yclients API.
type YclientsCredentials struct {
	PartnerToken string `json:"partner_token"`
	UserToken    string `json:"user_token"`
}

// yclientsProvider implements PMSProvider for Yclients integration.
type yclientsProvider struct {
	httpClient *http.Client
}

// NewYclientsProvider creates a new Yclients PMS provider.
func NewYclientsProvider() PMSProvider {
	return &yclientsProvider{
		httpClient: &http.Client{
			Timeout: yclientsAPITimeout,
		},
	}
}

func (y *yclientsProvider) Name() domain.PMSProvider {
	return domain.PMSProviderYclients
}

func (y *yclientsProvider) TestConnection(ctx context.Context, credentials string) error {
	creds, err := parseYclientsCredentials(credentials)
	if err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, yclientsBaseURL+"/company", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	y.setHeaders(req, creds)

	resp, err := y.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("yclients api request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("yclients authentication failed: status %d", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("yclients api error: status %d", resp.StatusCode)
	}

	return nil
}

func (y *yclientsProvider) PullBookings(ctx context.Context, credentials string, externalID string, from, to time.Time) ([]domain.PMSBooking, error) {
	creds, err := parseYclientsCredentials(credentials)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	url := fmt.Sprintf("%s/records/%s?start_date=%s&end_date=%s",
		yclientsBaseURL, externalID,
		from.Format("2006-01-02"), to.Format("2006-01-02"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	y.setHeaders(req, creds)

	resp, err := y.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("yclients api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("yclients api error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp yclientsRecordsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	bookings := make([]domain.PMSBooking, 0, len(apiResp.Data))
	for _, r := range apiResp.Data {
		bookings = append(bookings, domain.PMSBooking{
			ExternalID: fmt.Sprintf("%d", r.ID),
			StartTime:  parseYclientsDateTime(r.Date, r.Datetime),
			EndTime:    parseYclientsDateTime(r.Date, r.Datetime).Add(time.Duration(r.Length) * time.Second),
			GuestName:  r.Client.Name,
			GuestPhone: r.Client.Phone,
			Status:     mapYclientsStatus(r.Attendance),
			Notes:      r.Comment,
		})
	}

	return bookings, nil
}

func (y *yclientsProvider) PushBooking(ctx context.Context, credentials string, externalID string, booking domain.PMSBooking) error {
	creds, err := parseYclientsCredentials(credentials)
	if err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	payload := map[string]interface{}{
		"staff_id":   0,
		"datetime":   booking.StartTime.Format("2006-01-02T15:04:05-07:00"),
		"comment":    booking.Notes,
		"client":     map[string]string{"name": booking.GuestName, "phone": booking.GuestPhone},
		"attendance": 0,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/records/%s", yclientsBaseURL, externalID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	y.setHeaders(req, creds)
	req.Header.Set("Content-Type", "application/json")

	resp, err := y.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("yclients api request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("yclients push booking failed: status %d", resp.StatusCode)
	}

	return nil
}

func (y *yclientsProvider) SyncSchedule(ctx context.Context, credentials string, externalID string, from, to time.Time) ([]domain.PMSScheduleSlot, error) {
	creds, err := parseYclientsCredentials(credentials)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	url := fmt.Sprintf("%s/schedule/%s?start_date=%s&end_date=%s",
		yclientsBaseURL, externalID,
		from.Format("2006-01-02"), to.Format("2006-01-02"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	y.setHeaders(req, creds)

	resp, err := y.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("yclients api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("yclients schedule api error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp yclientsScheduleResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode schedule response: %w", err)
	}

	var slots []domain.PMSScheduleSlot
	for _, day := range apiResp.Data {
		for _, slot := range day.Slots {
			slots = append(slots, domain.PMSScheduleSlot{
				Date:      day.Date,
				TimeFrom:  slot.TimeFrom,
				TimeTo:    slot.TimeTo,
				Available: slot.Available,
			})
		}
	}

	return slots, nil
}

func (y *yclientsProvider) CancelBooking(ctx context.Context, credentials string, externalID string, bookingExternalID string) error {
	creds, err := parseYclientsCredentials(credentials)
	if err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	url := fmt.Sprintf("%s/records/%s/%s", yclientsBaseURL, externalID, bookingExternalID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	y.setHeaders(req, creds)

	resp, err := y.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("yclients api request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("yclients cancel booking failed: status %d", resp.StatusCode)
	}

	return nil
}

func (y *yclientsProvider) setHeaders(req *http.Request, creds YclientsCredentials) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s, User %s", creds.PartnerToken, creds.UserToken))
	req.Header.Set("Accept", "application/vnd.yclients.v2+json")
}

func parseYclientsCredentials(raw string) (YclientsCredentials, error) {
	var creds YclientsCredentials
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return creds, fmt.Errorf("parse yclients credentials: %w", err)
	}
	if creds.PartnerToken == "" || creds.UserToken == "" {
		return creds, fmt.Errorf("missing required yclients credentials")
	}
	return creds, nil
}

// Yclients API response types

type yclientsRecordsResponse struct {
	Data []yclientsRecord `json:"data"`
}

type yclientsRecord struct {
	ID         int            `json:"id"`
	Date       string         `json:"date"`
	Datetime   string         `json:"datetime"`
	Length     int            `json:"length"`
	Comment    string         `json:"comment"`
	Attendance int            `json:"attendance"`
	Client     yclientsClient `json:"client"`
}

type yclientsClient struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type yclientsScheduleResponse struct {
	Data []yclientsScheduleDay `json:"data"`
}

type yclientsScheduleDay struct {
	Date  string         `json:"date"`
	Slots []yclientsSlot `json:"slots"`
}

type yclientsSlot struct {
	TimeFrom  string `json:"time_from"`
	TimeTo    string `json:"time_to"`
	Available bool   `json:"available"`
}

func parseYclientsDateTime(date, datetime string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05-07:00", datetime)
	if err != nil {
		t, _ = time.Parse("2006-01-02", date)
	}
	return t
}

func mapYclientsStatus(attendance int) string {
	switch attendance {
	case 1:
		return "confirmed"
	case 2:
		return "completed"
	case -1:
		return "cancelled"
	default:
		return "pending"
	}
}
