package handler

import (
	"log"
	"net/http"
)

// const readinessMessage = `
// <html>
// 	<body>
// 		<h2>Chirpin'</h2>
// 		<p>OK<p>
// 	</body>
// </html>
// `

const readinessMessage = "OK"

func Health(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(readinessMessage))
		if err != nil {
			log.Printf("could not write to health endpoint: %s", err.Error())
		}
	}
}
