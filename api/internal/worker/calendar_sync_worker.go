package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

const (
	googleEventsURL     = "https://www.googleapis.com/calendar/v3/calendars/%s/events"
	googleRefreshURL    = "https://oauth2.googleapis.com/token"
	microsoftEventsURL  = "https://graph.microsoft.com/v1.0/me/calendarView"
	microsoftRefreshURL = "https://login.microsoftonline.com/%s/oauth2/v2.0/token"
)

// CalendarSyncWorker polls all calendar connections every interval and
// creates/updates activities from calendar events.
type CalendarSyncWorker struct {
	connections repository.CalendarConnectionRepository
	activities  repository.ActivityRepository
	interval    time.Duration
	log         *slog.Logger

	// OAuth credentials for token refresh.
	googleClientID        string
	googleClientSecret    string
	microsoftClientID     string
	microsoftClientSecret string
	microsoftTenantID     string
	scheduler             *SchedulerRegistry
	schedulerKey          string
}

// NewCalendarSyncWorker creates a new CalendarSyncWorker.
func NewCalendarSyncWorker(
	connections repository.CalendarConnectionRepository,
	activities repository.ActivityRepository,
	interval time.Duration,
	log *slog.Logger,
	googleClientID, googleClientSecret string,
	microsoftClientID, microsoftClientSecret, microsoftTenantID string,
) *CalendarSyncWorker {
	return &CalendarSyncWorker{
		connections:           connections,
		activities:            activities,
		interval:              interval,
		log:                   log,
		googleClientID:        googleClientID,
		googleClientSecret:    googleClientSecret,
		microsoftClientID:     microsoftClientID,
		microsoftClientSecret: microsoftClientSecret,
		microsoftTenantID:     microsoftTenantID,
	}
}

func (w *CalendarSyncWorker) WithScheduler(registry *SchedulerRegistry, key string) *CalendarSyncWorker {
	w.scheduler = registry
	w.schedulerKey = key
	return w
}

// Start launches the sync loop in a background goroutine.
func (w *CalendarSyncWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	go func() {
		defer ticker.Stop()
		w.log.Info("calendar sync worker started", "interval", w.interval)
		for {
			select {
			case <-ticker.C:
				if err := w.scheduler.TrackRun(w.schedulerKey, w.interval, func() error {
					return w.runSync(ctx)
				}); err != nil {
					w.log.Error("calendar sync worker failed", "err", err)
				}
			case <-ctx.Done():
				w.log.Info("calendar sync worker stopped")
				return
			}
		}
	}()
}

// runSync fetches all connections and syncs each one.
func (w *CalendarSyncWorker) runSync(ctx context.Context) error {
	conns, err := w.connections.ListAllActive(ctx)
	if err != nil {
		return fmt.Errorf("list connections: %w", err)
	}
	var firstErr error
	for _, conn := range conns {
		if err := w.syncConnection(ctx, conn); err != nil {
			w.log.Warn("calendar sync: connection failed",
				"connection_id", conn.ID,
				"provider", conn.Provider,
				"user_id", conn.UserID,
				"err", err,
			)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// syncConnection syncs a single calendar connection.
func (w *CalendarSyncWorker) syncConnection(ctx context.Context, conn *domain.CalendarConnection) error {
	// Refresh token if it expires within 5 minutes.
	if conn.TokenExpiry != nil && time.Until(*conn.TokenExpiry) < 5*time.Minute {
		if err := w.refreshToken(ctx, conn); err != nil {
			return fmt.Errorf("token refresh: %w", err)
		}
	}

	var events []domain.CalendarEvent
	var newCursor string
	var err error

	switch conn.Provider {
	case domain.CalendarProviderGoogle:
		events, newCursor, err = w.fetchGoogleEvents(conn)
	case domain.CalendarProviderMicrosoft:
		events, newCursor, err = w.fetchMicrosoftEvents(conn)
	default:
		return fmt.Errorf("unknown provider: %s", conn.Provider)
	}
	if err != nil {
		return fmt.Errorf("fetch events: %w", err)
	}

	// Upsert activities for each event.
	for _, evt := range events {
		if err := w.upsertActivity(ctx, conn, evt); err != nil {
			w.log.Warn("calendar sync: failed to upsert activity",
				"event_id", evt.ExternalID,
				"err", err,
			)
		}
	}

	// Persist the updated sync cursor.
	if newCursor != "" {
		patch := domain.CalendarConnectionPatch{SyncCursor: &newCursor}
		if _, err := w.connections.Update(ctx, conn.ID, patch); err != nil {
			w.log.Warn("calendar sync: failed to update cursor", "connection_id", conn.ID, "err", err)
		}
	}

	return nil
}

// upsertActivity creates or updates an activity for a calendar event.
func (w *CalendarSyncWorker) upsertActivity(ctx context.Context, conn *domain.CalendarConnection, evt domain.CalendarEvent) error {
	activityType := domain.ActivityTypeMeeting
	activity := &domain.Activity{
		OrgID:           conn.OrgID,
		Type:            activityType,
		Subject:         evt.Title,
		OwnerID:         conn.UserID,
		CalendarEventID: &evt.ExternalID,
		DueDate:         &evt.StartAt,
		StartAt:         &evt.StartAt,
		EndAt:           &evt.EndAt,
	}
	if evt.Description != "" {
		activity.Description = &evt.Description
	}

	// Context without org scoping (worker runs cross-org).
	_, err := w.activities.Create(ctx, activity)
	return err
}

// ─── Google Calendar ──────────────────────────────────────────────────────────

type googleEventsResponse struct {
	Items         []googleEvent `json:"items"`
	NextSyncToken string        `json:"nextSyncToken"`
	NextPageToken string        `json:"nextPageToken"`
}

type googleEvent struct {
	ID          string          `json:"id"`
	Summary     string          `json:"summary"`
	Description string          `json:"description"`
	Status      string          `json:"status"`
	Start       googleEventTime `json:"start"`
	End         googleEventTime `json:"end"`
}

type googleEventTime struct {
	DateTime string `json:"dateTime"` // RFC3339
	Date     string `json:"date"`     // all-day: YYYY-MM-DD
}

func (w *CalendarSyncWorker) fetchGoogleEvents(conn *domain.CalendarConnection) ([]domain.CalendarEvent, string, error) {
	calID := "primary"
	if conn.CalendarID != nil && *conn.CalendarID != "" {
		calID = *conn.CalendarID
	}

	params := url.Values{
		"singleEvents": {"true"},
		"orderBy":      {"startTime"},
		"timeMin":      {time.Now().Add(-24 * time.Hour).Format(time.RFC3339)},
		"timeMax":      {time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)},
		"maxResults":   {"250"},
	}
	if conn.SyncCursor != nil && *conn.SyncCursor != "" {
		params = url.Values{"syncToken": {*conn.SyncCursor}}
	}

	reqURL := fmt.Sprintf(googleEventsURL, url.PathEscape(calID)) + "?" + params.Encode()
	body, err := doCalendarRequest(reqURL, conn.AccessToken)
	if err != nil {
		return nil, "", err
	}

	var resp googleEventsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, "", fmt.Errorf("decode google events: %w", err)
	}

	var events []domain.CalendarEvent
	for _, item := range resp.Items {
		if item.Status == "cancelled" {
			continue
		}
		startAt, err := parseGoogleTime(item.Start)
		if err != nil {
			continue
		}
		endAt, err := parseGoogleTime(item.End)
		if err != nil {
			endAt = startAt.Add(time.Hour)
		}
		events = append(events, domain.CalendarEvent{
			ExternalID:  item.ID,
			Title:       item.Summary,
			Description: item.Description,
			StartAt:     startAt,
			EndAt:       endAt,
			IsAllDay:    item.Start.Date != "",
		})
	}

	return events, resp.NextSyncToken, nil
}

func parseGoogleTime(t googleEventTime) (time.Time, error) {
	if t.DateTime != "" {
		return time.Parse(time.RFC3339, t.DateTime)
	}
	if t.Date != "" {
		return time.Parse("2006-01-02", t.Date)
	}
	return time.Time{}, fmt.Errorf("no time value")
}

// ─── Microsoft Graph ──────────────────────────────────────────────────────────

type msEventsResponse struct {
	Value    []msEvent `json:"value"`
	NextLink string    `json:"@odata.nextLink"`
}

type msEvent struct {
	ID          string `json:"id"`
	Subject     string `json:"subject"`
	BodyPreview string `json:"bodyPreview"`
	Start       msTime `json:"start"`
	End         msTime `json:"end"`
	IsAllDay    bool   `json:"isAllDay"`
	IsCancelled bool   `json:"isCancelled"`
}

type msTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

func (w *CalendarSyncWorker) fetchMicrosoftEvents(conn *domain.CalendarConnection) ([]domain.CalendarEvent, string, error) { //nolint:unparam // cursor always "" — pagination not yet implemented for Microsoft
	now := time.Now().UTC()
	params := url.Values{
		"startDateTime": {now.Add(-24 * time.Hour).Format(time.RFC3339)},
		"endDateTime":   {now.Add(30 * 24 * time.Hour).Format(time.RFC3339)},
		"$top":          {"250"},
		"$select":       {"id,subject,bodyPreview,start,end,isAllDay,isCancelled"},
	}

	reqURL := microsoftEventsURL + "?" + params.Encode()
	body, err := doCalendarRequest(reqURL, conn.AccessToken)
	if err != nil {
		return nil, "", err
	}

	var resp msEventsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, "", fmt.Errorf("decode microsoft events: %w", err)
	}

	var events []domain.CalendarEvent
	for _, item := range resp.Value {
		if item.IsCancelled {
			continue
		}
		startAt, err := time.Parse("2006-01-02T15:04:05.0000000", item.Start.DateTime)
		if err != nil {
			startAt, err = time.Parse(time.RFC3339, item.Start.DateTime)
			if err != nil {
				continue
			}
		}
		endAt, err := time.Parse("2006-01-02T15:04:05.0000000", item.End.DateTime)
		if err != nil {
			endAt = startAt.Add(time.Hour)
		}
		events = append(events, domain.CalendarEvent{
			ExternalID:  item.ID,
			Title:       item.Subject,
			Description: item.BodyPreview,
			StartAt:     startAt.UTC(),
			EndAt:       endAt.UTC(),
			IsAllDay:    item.IsAllDay,
		})
	}

	// nextLink carries paging token but we don't store it as sync cursor for MS.
	return events, "", nil
}

// ─── Token refresh ────────────────────────────────────────────────────────────

func (w *CalendarSyncWorker) refreshToken(ctx context.Context, conn *domain.CalendarConnection) error {
	if conn.RefreshToken == nil {
		return fmt.Errorf("no refresh token available for connection %s", conn.ID)
	}

	var (
		tokenURL string
		formBody url.Values
	)

	switch conn.Provider {
	case domain.CalendarProviderGoogle:
		tokenURL = googleRefreshURL
		formBody = url.Values{
			"refresh_token": {*conn.RefreshToken},
			"client_id":     {w.googleClientID},
			"client_secret": {w.googleClientSecret},
			"grant_type":    {"refresh_token"},
		}
	case domain.CalendarProviderMicrosoft:
		tokenURL = fmt.Sprintf(microsoftRefreshURL, w.microsoftTenantID)
		formBody = url.Values{
			"refresh_token": {*conn.RefreshToken},
			"client_id":     {w.microsoftClientID},
			"client_secret": {w.microsoftClientSecret},
			"grant_type":    {"refresh_token"},
			"scope":         {"https://graph.microsoft.com/Calendars.ReadWrite offline_access"},
		}
	default:
		return fmt.Errorf("unknown provider: %s", conn.Provider)
	}

	resp, err := http.PostForm(tokenURL, formBody) //nolint:noctx,gosec // tokenURL is a known OAuth provider endpoint
	if err != nil {
		return fmt.Errorf("refresh request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(raw, &tok); err != nil || tok.AccessToken == "" {
		return fmt.Errorf("invalid refresh response")
	}

	expiry := time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
	patch := domain.CalendarConnectionPatch{
		AccessToken: &tok.AccessToken,
		TokenExpiry: &expiry,
	}
	updated, err := w.connections.Update(ctx, conn.ID, patch)
	if err != nil {
		return fmt.Errorf("persist refreshed token: %w", err)
	}
	conn.AccessToken = updated.AccessToken
	conn.TokenExpiry = updated.TokenExpiry
	return nil
}

// ─── HTTP helper ──────────────────────────────────────────────────────────────

func doCalendarRequest(reqURL, accessToken string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calendar API request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("calendar API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}
