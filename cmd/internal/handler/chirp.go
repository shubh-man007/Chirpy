package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"
)

var profane = []string{"kerfuffle", "sharbert", "fornax"}
var astrix = "****"

type ChirpBody struct {
	Body string `json:"body"`
}

type ChirpLenValid struct {
	Body    string `json:"cleaned_body"`
	Message bool   `json:"valid"`
}

type ChirpLenErr struct {
	Message string `json:"error"`
}

func errJSON(w http.ResponseWriter) {
	errMess := ChirpLenErr{Message: "Something went wrong"}
	data, _ := json.Marshal(errMess)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)
	w.Write(data)
}

func ValidateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	chirp := ChirpBody{}
	err := decoder.Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding requested JSON: %s", err)
		errJSON(w)
		return
	}

	chirpData := chirp.Body

	for _, prof := range profane {
		if strings.Contains(chirpData, prof) {
			chirpData = strings.ReplaceAll(chirpData, prof, astrix)
		}
	}

	switch {
	case utf8.RuneCountInString(chirpData) <= 140:
		resBody := ChirpLenValid{Body: chirpData, Message: true}

		dat, err := json.Marshal(resBody)
		if err != nil {
			log.Printf("Error marshalling Valid JSON: %s", err)
			errJSON(w)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(dat)

	case utf8.RuneCountInString(chirpData) > 140:
		resBody := ChirpLenErr{Message: "Chirp is too long"}

		dat, err := json.Marshal(resBody)
		if err != nil {
			log.Printf("Error marshalling Err JSON: %s", err)
			errJSON(w)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(dat)
	}
}
