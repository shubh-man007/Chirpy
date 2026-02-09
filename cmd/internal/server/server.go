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

func New(port string, db *database.Queries, platform string, jwtSecret string, polkaAPI string) *Server {
	cfg := config.NewApiCfg(db, platform, jwtSecret, polkaAPI)
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
	mux.HandleFunc("POST /api/login", apiHandler.LoginUser)
	mux.HandleFunc("POST /api/users", apiHandler.CreateUser)
	mux.HandleFunc("PUT /api/users", apiHandler.UpdateUserCred)
	mux.HandleFunc("DELETE /api/users/{userID}", apiHandler.DeleteUser)

	//auth:
	mux.HandleFunc("POST /api/refresh", apiHandler.RefreshToken)
	mux.HandleFunc("POST /api/revoke", apiHandler.RevokeToken)

	//chirps:
	mux.HandleFunc("POST /api/chirps", apiHandler.CreateChirp)
	mux.HandleFunc("GET /api/chirps", apiHandler.GetAllChirps)
	mux.HandleFunc("GET /api/me/chirps", apiHandler.GetMyChirps)
	mux.HandleFunc("GET /api/users/{userID}/chirps", apiHandler.GetChirpsByUser)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiHandler.GetChirpByChirpID)
	mux.HandleFunc("PATCH /api/chirps/{chirpID}", apiHandler.UpdateChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiHandler.DeleteChirp)

	// friends:
	mux.HandleFunc("POST /api/friends/request", apiHandler.SendFriendRequest)
	mux.HandleFunc("POST /api/friends/{userID}/accept", apiHandler.AcceptFriendRequest)
	mux.HandleFunc("POST /api/friends/{userID}/reject", apiHandler.RejectFriendRequest)
	mux.HandleFunc("DELETE /api/friends/{userID}", apiHandler.RemoveFriend)
	mux.HandleFunc("GET /api/friends", apiHandler.GetFriends)
	mux.HandleFunc("GET /api/friends/requests", apiHandler.GetPendingFriendRequests)
	mux.HandleFunc("GET /api/friends/sent", apiHandler.GetSentFriendRequests)

	// feed:
	mux.HandleFunc("GET /api/feed", apiHandler.GetFeed)

	// membership:
	mux.HandleFunc("POST /api/polka/webhooks", apiHandler.UpdateUserMembership)

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
