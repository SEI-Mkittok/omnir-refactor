package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/omnir/crm-api/internal/middleware"
	"github.com/omnir/crm-api/internal/repository"
)

type PreferenceSettingsHandler struct {
	repo repository.UserPreferenceRepository
}

func NewPreferenceSettingsHandler(repo repository.UserPreferenceRepository) *PreferenceSettingsHandler {
	return &PreferenceSettingsHandler{repo: repo}
}

func (h *PreferenceSettingsHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetPreferences)
	r.Patch("/", h.UpdatePreferences)
	return r
}

func (h *PreferenceSettingsHandler) CalendarRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetCalendarPreferences)
	r.Patch("/", h.UpdateCalendarPreferences)
	return r
}

func (h *PreferenceSettingsHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	pref, err := h.repo.GetByUser(r.Context(), claims.UserID, claims.OrgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, pref)
}

func validatePreferencesPatch(patch map[string]json.RawMessage) (int, int, bool) {
	var dayStart, duration int
	hasDayStart := false
	if raw, ok := patch["calendar_day_start_hour"]; ok {
		hasDayStart = true
		_ = json.Unmarshal(raw, &dayStart)
	}
	if raw, ok := patch["calendar_default_duration_minutes"]; ok {
		_ = json.Unmarshal(raw, &duration)
	}
	return dayStart, duration, hasDayStart
}

func (h *PreferenceSettingsHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var patchMap map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&patchMap); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var patch repositoryPatch
	if err := decodePatchMapIntoPreferences(patchMap, &patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dayStart, duration, hasDayStart := validatePreferencesPatch(patchMap)
	if hasDayStart && (dayStart < 0 || dayStart > 23) {
		writeError(w, http.StatusUnprocessableEntity, "calendar_day_start_hour must be between 0 and 23")
		return
	}
	if _, ok := patchMap["calendar_default_duration_minutes"]; ok && (duration < 5 || duration > 480) {
		writeError(w, http.StatusUnprocessableEntity, "calendar_default_duration_minutes must be between 5 and 480")
		return
	}
	if patch.CalendarDefaultView != nil {
		v := strings.TrimSpace(strings.ToLower(*patch.CalendarDefaultView))
		if v != "" && v != "month" && v != "week" {
			writeError(w, http.StatusUnprocessableEntity, "calendar_default_view must be one of: month, week")
			return
		}
		patch.CalendarDefaultView = &v
	}
	if patch.CalendarHourFormat != nil {
		v := strings.TrimSpace(strings.ToLower(*patch.CalendarHourFormat))
		if v != "" && v != "12h" && v != "24h" {
			writeError(w, http.StatusUnprocessableEntity, "calendar_hour_format must be one of: 12h, 24h")
			return
		}
		patch.CalendarHourFormat = &v
	}

	updated, err := h.repo.Upsert(r.Context(), claims.UserID, claims.OrgID, patch.toDomainPatch())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *PreferenceSettingsHandler) GetCalendarPreferences(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	pref, err := h.repo.GetByUser(r.Context(), claims.UserID, claims.OrgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"calendar_start_day":                pref.CalendarStartDay,
		"calendar_date_format":              pref.CalendarDateFormat,
		"calendar_time_zone":                pref.CalendarTimeZone,
		"calendar_default_activity_status":  pref.CalendarDefaultActivityStatus,
		"calendar_default_duration_minutes": pref.CalendarDefaultDurationMins,
		"calendar_reminder_interval_minutes": pref.CalendarReminderIntervalMins,
		"calendar_default_view":             pref.CalendarDefaultView,
		"calendar_day_start_hour":           pref.CalendarDayStartHour,
		"calendar_hour_format":              pref.CalendarHourFormat,
		"calendar_default_activity_type":    pref.CalendarDefaultActivityType,
		"calendar_show_completed_events":    pref.CalendarShowCompletedEvents,
	})
}

func (h *PreferenceSettingsHandler) UpdateCalendarPreferences(w http.ResponseWriter, r *http.Request) {
	h.UpdatePreferences(w, r)
}
