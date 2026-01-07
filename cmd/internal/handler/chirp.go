package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"unicode/utf8"
)

type ChirpBody struct {
	Body string `json:"body"`
}

type ChirpLenValid struct {
	Message bool `json:"valid"`
}

type ChirpLenErr struct {
	Message string `json:"error"`
}

func ValidateChirpLen(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	chirp := ChirpBody{}
	err := decoder.Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding requested JSON: %s", err)
		errMess := ChirpLenErr{Message: "Something went wrong"}
		data, _ := json.Marshal(errMess)
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}

	chirpData := chirp.Body

	switch {
	case utf8.RuneCountInString(chirpData) <= 140:
		resBody := ChirpLenValid{Message: true}

		dat, err := json.Marshal(resBody)
		if err != nil {
			log.Printf("Error marshalling Valid JSON: %s", err)
			errMess := ChirpLenErr{Message: "Something went wrong"}
			data, _ := json.Marshal(errMess)
			w.WriteHeader(500)
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(dat)

	case utf8.RuneCountInString(chirpData) > 140:
		resBody := ChirpLenErr{Message: "Chirp is too long"}

		dat, err := json.Marshal(resBody)
		if err != nil {
			log.Printf("Error marshalling Err JSON: %s", err)
			errMess := ChirpLenErr{Message: "Something went wrong"}
			data, _ := json.Marshal(errMess)
			w.WriteHeader(500)
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(dat)
	}
}
