package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/shubh-man007/Chirpy/cmd/internal/middleware"
)

const metricBody = `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`

type AdminHandler struct {
	apiCfg *middleware.ApiConfig
}

func NewAdminHandler(cfg *middleware.ApiConfig) *AdminHandler {
	return &AdminHandler{apiCfg: cfg}
}

func (h *AdminHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	hits := h.apiCfg.FileserverHits.Load()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(fmt.Sprintf(metricBody, hits)))
	if err != nil {
		log.Printf("could not write to metric endpoint: %s", err)
	}
}

func (h *AdminHandler) Reset(w http.ResponseWriter, r *http.Request) {
	h.apiCfg.FileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("Hits: 0"))
	if err != nil {
		log.Printf("could not write to reset endpoint: %s", err)
	}
}
