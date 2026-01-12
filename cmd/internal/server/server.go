package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shubh-man007/Chirpy/cmd/internal/config"
	"github.com/shubh-man007/Chirpy/cmd/internal/database"
	"github.com/shubh-man007/Chirpy/cmd/internal/handler"
	"github.com/shubh-man007/Chirpy/cmd/internal/middleware"
)

type Server struct {
	Port       string
	apiCfg     *config.ApiConfig
	httpServer *http.Server
}

func New(port string, db *database.Queries) *Server {
	cfg := config.NewApiCfg(db)
	cfg.FileserverHits.Store(0)

	return &Server{
		Port:   port,
		apiCfg: cfg,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../static"))))
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("../assets"))))

	fileserver := http.FileServer(http.Dir("../static"))
	mux.Handle("/app/", http.StripPrefix("/app/", middleware.HitCounterMiddleware(s.apiCfg, fileserver)))

	mux.HandleFunc("GET /api/healthz", handler.Health)
	mux.HandleFunc("POST /api/validate_chirp", handler.ValidateChirp)
	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		type userEmail struct {
			Email string `json:"email"`
		}

		decoder := json.NewDecoder(r.Body)
		use := userEmail{}
		err := decoder.Decode(&use)
		if err != nil {
			log.Printf("Error decoding request JSON: %s", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			w.Write([]byte("Email not found"))
			return
		}

		email := use.Email
		user, err := s.apiCfg.DB.CreateUser(r.Context(), email)
		if err != nil {
			log.Printf("Error decoding request JSON: %s", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			w.Write([]byte("could no create user"))
			return
		}

		type Users struct {
			ID        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Email     string    `json:"email"`
		}

		resBody := Users{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		}

		data, err := json.Marshal(resBody)
		if err != nil {
			log.Printf("Error encoding response JSON: %s", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			w.Write([]byte("could no marshal user"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	adminHandler := handler.NewAdminHandler(s.apiCfg)
	mux.HandleFunc("GET /admin/metrics", adminHandler.Metrics)
	mux.HandleFunc("POST /admin/reset", adminHandler.Reset)

	return middleware.LogMiddleware(mux)
}

func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:    ":" + s.Port,
		Handler: s.Routes(),
	}

	log.Printf("Running server at port:%s", s.Port)
	return s.httpServer.ListenAndServe()
}
