package internal

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

func RespondWithError(w http.ResponseWriter, code int, message string) {
	type errorBody struct {
		Error string `json:"error"`
	}

	errBody := errorBody{
		Error: message,
	}
	dat, err := json.Marshal(errBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	var dat []byte
	var err error
	if payload != nil {
		dat, err = json.Marshal(payload)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func GetHeaderApiKey(_ http.ResponseWriter, r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	parts := strings.Split(auth, " ")
	var key string
	var err error
	if len(parts) < 2 {
		err = errors.New("missing Authorization header")
	} else {
		key = parts[1]
	}
	return key, err
}
