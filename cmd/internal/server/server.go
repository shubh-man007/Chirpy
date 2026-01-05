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

	fileserver := http.FileServer(http.Dir("../static"))
	assetserver := http.FileServer(http.Dir("../assets"))

	mux.Handle("/app/", http.StripPrefix("/app/", s.apiCfg.HitCounterMiddleware(fileserver)))
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", s.apiCfg.HitCounterMiddleware(assetserver)))

	mux.HandleFunc("GET /api/healthz", handler.Health)

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
