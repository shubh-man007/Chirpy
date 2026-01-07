package server

import (
	"log"
	"net/http"

	"github.com/shubh-man007/Chirpy/cmd/internal/handler"
	"github.com/shubh-man007/Chirpy/cmd/internal/middleware"
)

type Server struct {
	Port       string
	apiCfg     *middleware.ApiConfig
	httpServer *http.Server
}

func New(port string) *Server {
	cfg := middleware.NewApiCfg()
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
	mux.Handle("/app/", http.StripPrefix("/app/", s.apiCfg.HitCounterMiddleware(fileserver)))

	mux.HandleFunc("GET /api/healthz", handler.Health)
	mux.HandleFunc("POST /api/validate_chirp", handler.ValidateChirpLen)

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
