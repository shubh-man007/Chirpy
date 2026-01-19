package server

import (
	"log"
	"net/http"

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

func New(port string, db *database.Queries, platform string, jwt_secret string) *Server {
	cfg := config.NewApiCfg(db, platform, jwt_secret)
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

	//app:
	fileserver := http.FileServer(http.Dir("../static"))
	mux.Handle("/app/", http.StripPrefix("/app/", middleware.HitCounterMiddleware(s.apiCfg, fileserver)))

	//api:
	apiHandler := handler.NewAPIHandler(s.apiCfg)

	//readiness
	mux.HandleFunc("GET /api/healthz", handler.Health)

	//users:
	mux.HandleFunc("POST /api/users", apiHandler.CreateUser)
	mux.HandleFunc("PUT /api/users", apiHandler.UpdateUserCred)

	//auth:
	mux.HandleFunc("POST /api/login", apiHandler.LoginUser)
	mux.HandleFunc("POST /api/refresh", apiHandler.RefreshToken)
	mux.HandleFunc("POST /api/revoke", apiHandler.RevokeToken)

	//chirps:
	mux.HandleFunc("POST /api/chirps", apiHandler.CreateChirp)
	mux.HandleFunc("GET /api/chirps", apiHandler.GetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiHandler.GetChirpsByUser)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiHandler.DeleteChirp)

	//admin:
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
