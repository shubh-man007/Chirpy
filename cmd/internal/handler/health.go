package handler

import (
	"log"
	"net/http"
)

const readinessMessage = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <link rel="icon" href="/app/favicon.ico" type="image/x-icon">
  <title>Health Check - Chirpy</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&family=Inter:wght@400;500;600&display=swap" rel="stylesheet">
  <link rel="stylesheet" href="/app/styles.css" />
</head>
<body>
  <div class="container compact">
    <h2><span class="accent">Chirpin'</span></h2>
    <div class="status-badge success">OK</div>
    <p class="info-text">All systems operational</p>
  </div>
</body>
</html>
`

func Health(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(readinessMessage))
		if err != nil {
			log.Printf("could not write to health endpoint: %s", err.Error())
		}
	}
}
