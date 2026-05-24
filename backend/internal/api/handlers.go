package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"skinemsia/internal/store"
)

// ── Events ─────────────────────────────────────────────────────────────────────

func (s *Server) listEvents(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	events, err := s.store.ListUserEvents(r.Context(), u.ID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if events == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}
	jsonOK(w, events)
}

func (s *Server) createEvent(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)

	var in struct {
		Title          string `json:"title"`
		CollectorName  string `json:"collectorName"`
		CollectorPhone string `json:"collectorPhone"`
		Currency       string `json:"currency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if in.Title == "" {
		jsonError(w, "title is required", http.StatusBadRequest)
		return
	}

	e, err := s.store.CreateEvent(r.Context(), u.ID, store.CreateEventInput{
		Title:          in.Title,
		CollectorName:  in.CollectorName,
		CollectorPhone: in.CollectorPhone,
		Currency:       in.Currency,
	})
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, e)
}

func (s *Server) getEvent(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	eventID := parseID(r, "eventID")

	e, err := s.store.GetEvent(r.Context(), eventID, u.ID)
	if err != nil {
		handleStoreError(w, err)
		return
	}
	jsonOK(w, e)
}

func (s *Server) updateEvent(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	eventID := parseID(r, "eventID")

	var in struct {
		Title                   *string `json:"title"`
		CollectorName           *string `json:"collectorName"`
		CollectorPhone          *string `json:"collectorPhone"`
		Currency                *string `json:"currency"`
		AllowMembersAddExpenses *bool   `json:"allowMembersAddExpenses"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}

	e, err := s.store.UpdateEvent(r.Context(), eventID, u.ID, store.UpdateEventInput{
		Title:                   in.Title,
		CollectorName:           in.CollectorName,
		CollectorPhone:          in.CollectorPhone,
		Currency:                in.Currency,
		AllowMembersAddExpenses: in.AllowMembersAddExpenses,
	})
	if err != nil {
		handleStoreError(w, err)
		return
	}
	jsonOK(w, e)
}

func (s *Server) archiveEvent(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	eventID := parseID(r, "eventID")

	if err := s.store.ArchiveEvent(r.Context(), eventID, u.ID); err != nil {
		handleStoreError(w, err)
		return
	}
	jsonOK(w, map[string]string{"status": "archived"})
}

func (s *Server) deleteEvent(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	eventID := parseID(r, "eventID")

	if err := s.store.DeleteEvent(r.Context(), eventID, u.ID); err != nil {
		handleStoreError(w, err)
		return
	}
	jsonOK(w, map[string]string{"status": "deleted"})
}

func (s *Server) previewEvent(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	e, err := s.store.GetEventByCode(r.Context(), code)
	if err != nil {
		handleStoreError(w, err)
		return
	}
	jsonOK(w, map[string]interface{}{
		"id":    e.ID,
		"title": e.Title,
		"code":  e.JoinCode,
	})
}

func (s *Server) joinEvent(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)

	var in struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if in.Code == "" {
		jsonError(w, "code is required", http.StatusBadRequest)
		return
	}

	e, err := s.store.JoinEvent(r.Context(), in.Code, u.ID)
	if err != nil {
		handleStoreError(w, err)
		return
	}
	jsonOK(w, e)
}

func (s *Server) getSummary(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	eventID := parseID(r, "eventID")

	summary, err := s.store.GetSummary(r.Context(), eventID, u.ID)
	if err != nil {
		handleStoreError(w, err)
		return
	}
	jsonOK(w, summary)
}

// ── Members ────────────────────────────────────────────────────────────────────

func (s *Server) listMembers(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	eventID := parseID(r, "eventID")

	// verify membership
	if _, err := s.store.GetEvent(r.Context(), eventID, u.ID); err != nil {
		handleStoreError(w, err)
		return
	}

	members, err := s.store.GetMembers(r.Context(), eventID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, members)
}

func (s *Server) removeMember(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	eventID := parseID(r, "eventID")
	targetID := parseID(r, "userID")

	if err := s.store.RemoveMember(r.Context(), eventID, targetID, u.ID); err != nil {
		handleStoreError(w, err)
		return
	}
	jsonOK(w, map[string]string{"status": "removed"})
}

func (s *Server) updateMemberEmoji(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	eventID := parseID(r, "eventID")
	targetID := parseID(r, "userID")

	var in struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Emoji == "" {
		jsonError(w, "emoji is required", http.StatusBadRequest)
		return
	}

	if err := s.store.UpdateMemberEmoji(r.Context(), eventID, targetID, u.ID, in.Emoji); err != nil {
		handleStoreError(w, err)
		return
	}
	jsonOK(w, map[string]string{"status": "updated"})
}

// ── Expenses ───────────────────────────────────────────────────────────────────

func (s *Server) listExpenses(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	eventID := parseID(r, "eventID")

	if _, err := s.store.GetEvent(r.Context(), eventID, u.ID); err != nil {
		handleStoreError(w, err)
		return
	}

	expenses, err := s.store.ListExpenses(r.Context(), eventID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if expenses == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}
	jsonOK(w, expenses)
}

func (s *Server) getExpense(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	eventID := parseID(r, "eventID")
	expenseID := parseID(r, "expenseID")

	if _, err := s.store.GetEvent(r.Context(), eventID, u.ID); err != nil {
		handleStoreError(w, err)
		return
	}

	e, err := s.store.GetExpense(r.Context(), expenseID, eventID)
	if err != nil {
		handleStoreError(w, err)
		return
	}
	jsonOK(w, e)
}

func (s *Server) createExpense(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	eventID := parseID(r, "eventID")

	event, err := s.store.GetEvent(r.Context(), eventID, u.ID)
	if err != nil {
		handleStoreError(w, err)
		return
	}

	if !event.AllowMembersAddExpenses && event.CreatorUserID != u.ID {
		jsonError(w, "only creator can add expenses", http.StatusForbidden)
		return
	}

	var in struct {
		Title          string  `json:"title"`
		AmountMinor    int64   `json:"amountMinor"`
		PaidByUserID   int64   `json:"paidByUserId"`
		ParticipantIDs []int64 `json:"participantIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if in.Title == "" {
		jsonError(w, "title is required", http.StatusBadRequest)
		return
	}
	if in.AmountMinor <= 0 {
		jsonError(w, "amount must be positive", http.StatusBadRequest)
		return
	}
	if len(in.ParticipantIDs) == 0 {
		jsonError(w, "at least one participant required", http.StatusBadRequest)
		return
	}

	e, err := s.store.CreateExpense(r.Context(), eventID, u.ID, store.CreateExpenseInput{
		Title:          in.Title,
		AmountMinor:    in.AmountMinor,
		PaidByUserID:   in.PaidByUserID,
		ParticipantIDs: in.ParticipantIDs,
	})
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, e)
}

func (s *Server) updateExpense(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	eventID := parseID(r, "eventID")
	expenseID := parseID(r, "expenseID")

	var in struct {
		Title          string  `json:"title"`
		AmountMinor    int64   `json:"amountMinor"`
		PaidByUserID   int64   `json:"paidByUserId"`
		ParticipantIDs []int64 `json:"participantIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		jsonError(w, "invalid body", http.StatusBadRequest)
		return
	}
	if in.Title == "" || in.AmountMinor <= 0 || len(in.ParticipantIDs) == 0 {
		jsonError(w, "invalid input", http.StatusBadRequest)
		return
	}

	e, err := s.store.UpdateExpense(r.Context(), expenseID, eventID, u.ID, store.CreateExpenseInput{
		Title:          in.Title,
		AmountMinor:    in.AmountMinor,
		PaidByUserID:   in.PaidByUserID,
		ParticipantIDs: in.ParticipantIDs,
	})
	if err != nil {
		handleStoreError(w, err)
		return
	}
	jsonOK(w, e)
}

func (s *Server) deleteExpense(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	eventID := parseID(r, "eventID")
	expenseID := parseID(r, "expenseID")

	if err := s.store.DeleteExpense(r.Context(), expenseID, eventID, u.ID); err != nil {
		handleStoreError(w, err)
		return
	}
	jsonOK(w, map[string]string{"status": "deleted"})
}

// ── Payment ────────────────────────────────────────────────────────────────────

func (s *Server) markPaid(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	eventID := parseID(r, "eventID")

	if err := s.store.SetPaymentStatus(r.Context(), eventID, u.ID, "paid"); err != nil {
		handleStoreError(w, err)
		return
	}
	jsonOK(w, map[string]string{"paymentStatus": "paid"})
}

func (s *Server) unmarkPaid(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r)
	eventID := parseID(r, "eventID")

	if err := s.store.SetPaymentStatus(r.Context(), eventID, u.ID, "unpaid"); err != nil {
		handleStoreError(w, err)
		return
	}
	jsonOK(w, map[string]string{"paymentStatus": "unpaid"})
}

// ── util ───────────────────────────────────────────────────────────────────────

func parseID(r *http.Request, param string) int64 {
	id, _ := strconv.ParseInt(chi.URLParam(r, param), 10, 64)
	return id
}
