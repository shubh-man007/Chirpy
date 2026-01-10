package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/shubh-man007/Chirpy/cmd/internal/config"
)

const metricBody = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <link rel="icon" href="/static/favicon.ico" type="image/x-icon">
  <title>Metrics - Chirpy Admin</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&family=Inter:wght@400;500;600&display=swap" rel="stylesheet">
  <link rel="stylesheet" href="/static/styles.css" />
</head>
<body>
  <div class="container wide">
    <h1>Welcome, <strong>Admin</strong></h1>
    <p class="label">Total Visits</p>
    <div class="metric-value">%d</div>
    <p class="info-text">Chirpy has been visited this many times!</p>
  </div>
</body>
</html>
`

const resetBody = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <link rel="icon" href="/static/favicon.ico" type="image/x-icon">
  <title>Reset - Chirpy Admin</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&family=Inter:wght@400;500;600&display=swap" rel="stylesheet">
  <link rel="stylesheet" href="/static/styles.css" />
</head>
<body>
  <div class="container compact">
    <h2>Metrics <span class="accent">Reset</span></h2>
    <div class="status-badge success">Hits: 0</div>
    <p class="info-text">All visitor metrics have been reset successfully.</p>
  </div>
</body>
</html>
`

type AdminHandler struct {
	apiCfg *config.ApiConfig
}

func NewAdminHandler(cfg *config.ApiConfig) *AdminHandler {
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
	_, err := w.Write([]byte(resetBody))
	if err != nil {
		log.Printf("could not write to reset endpoint: %s", err)
	}
}
