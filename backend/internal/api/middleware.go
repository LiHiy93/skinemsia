package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"skinemsia/internal/models"
	"skinemsia/internal/store"
)

type contextKey string

const ctxUser contextKey = "user"

func userFromCtx(r *http.Request) *models.User {
	u, _ := r.Context().Value(ctxUser).(*models.User)
	return u
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tgID int64

		// dev mode: X-Dev-User-ID header skips real validation
		if s.cfg.IsDev() {
			if devID := r.Header.Get("X-Dev-User-ID"); devID != "" {
				id, err := strconv.ParseInt(devID, 10, 64)
				if err == nil {
					tgID = id
				}
			}
		}

		if tgID == 0 {
			initData := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if initData == "" {
				jsonError(w, "missing authorization", http.StatusUnauthorized)
				return
			}
			tgUser, err := validateInitData(initData, s.cfg.BotToken)
			if err != nil {
				jsonError(w, "invalid authorization", http.StatusUnauthorized)
				return
			}
			tgID = tgUser.ID

			// upsert user on every request (updates name if changed)
			u, err := s.store.UpsertUser(r.Context(), tgUser.ID, tgUser.Username, tgUser.FirstName, tgUser.LastName)
			if err != nil {
				jsonError(w, "db error", http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), ctxUser, u)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// dev path
		u, err := s.store.UpsertUser(r.Context(), tgID, "devuser", "Dev", "User")
		if err != nil {
			jsonError(w, "db error", http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(r.Context(), ctxUser, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type tgUserData struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

func validateInitData(initData, botToken string) (*tgUserData, error) {
	vals, err := url.ParseQuery(initData)
	if err != nil {
		return nil, err
	}

	hash := vals.Get("hash")
	if hash == "" {
		return nil, errors.New("missing hash")
	}
	vals.Del("hash")

	keys := make([]string, 0, len(vals))
	for k := range vals {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + "=" + vals.Get(k)
	}
	dataCheck := strings.Join(parts, "\n")

	// secret = HMAC-SHA256("WebAppData", botToken)
	mac := hmac.New(sha256.New, []byte("WebAppData"))
	mac.Write([]byte(botToken))
	secret := mac.Sum(nil)

	// verify = HMAC-SHA256(dataCheck, secret)
	mac2 := hmac.New(sha256.New, secret)
	mac2.Write([]byte(dataCheck))
	expected := hex.EncodeToString(mac2.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(hash)) {
		return nil, errors.New("invalid hash")
	}

	var user tgUserData
	if err := json.Unmarshal([]byte(vals.Get("user")), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func handleStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		jsonError(w, "not found", http.StatusNotFound)
	case errors.Is(err, store.ErrForbidden):
		jsonError(w, "forbidden", http.StatusForbidden)
	case errors.Is(err, store.ErrConflict):
		jsonError(w, "conflict", http.StatusConflict)
	default:
		jsonError(w, err.Error(), http.StatusBadRequest)
	}
}
