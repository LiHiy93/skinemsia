package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"skinemsia/internal/config"
	"skinemsia/internal/store"
)

type Server struct {
	router *chi.Mux
	store  *store.Store
	cfg    *config.Config
}

func NewServer(cfg *config.Config, st *store.Store) http.Handler {
	s := &Server{
		router: chi.NewRouter(),
		store:  st,
		cfg:    cfg,
	}
	s.routes()
	return s.router
}

func (s *Server) routes() {
	r := s.router

	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   strings.Split(s.cfg.AllowedOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Dev-User-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		jsonOK(w, map[string]string{"status": "ok"})
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(s.authMiddleware)

		r.Route("/events", func(r chi.Router) {
			r.Get("/", s.listEvents)
			r.Post("/", s.createEvent)
			r.Post("/join", s.joinEvent)
			r.Get("/preview/{code}", s.previewEvent)

			r.Route("/{eventID}", func(r chi.Router) {
				r.Get("/", s.getEvent)
				r.Patch("/", s.updateEvent)
				r.Delete("/", s.deleteEvent)
				r.Post("/archive", s.archiveEvent)
				r.Get("/summary", s.getSummary)

				r.Route("/members", func(r chi.Router) {
					r.Get("/", s.listMembers)
					r.Delete("/{userID}", s.removeMember)
					r.Patch("/{userID}/emoji", s.updateMemberEmoji)
				})

				r.Route("/expenses", func(r chi.Router) {
					r.Get("/", s.listExpenses)
					r.Post("/", s.createExpense)
					r.Get("/{expenseID}", s.getExpense)
					r.Put("/{expenseID}", s.updateExpense)
					r.Delete("/{expenseID}", s.deleteExpense)
				})

				r.Post("/payment/paid", s.markPaid)
				r.Delete("/payment/paid", s.unmarkPaid)
			})
		})
	})
}
