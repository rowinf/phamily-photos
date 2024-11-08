package internal

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/donseba/go-htmx"
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
	if _, err := w.Write(dat); err != nil {
		panic(err)
	}
}

func RespondWithErrorHtmx(h *htmx.Handler, w http.ResponseWriter, code int, message string) {
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
	if h.IsHxRequest() {
		h.TriggerError(message)
	}
	w.WriteHeader(code)
	if _, err := w.Write(dat); err != nil {
		panic(err)
	}
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
	if _, err := w.Write(dat); err != nil {
		panic(err)
	}
}

func RespondWithOk(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
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

// FileGenerator handles uploaded files and sends them over a channel
func FileGenerator(files []*multipart.FileHeader) <-chan *multipart.FileHeader {
	ch := make(chan *multipart.FileHeader)

	go func() {
		defer close(ch)
		for _, file := range files {
			// Send each file over the channel
			ch <- file
		}
	}()

	return ch
}

func SaveFile(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()
	// Get and print out file size
	fileSize := fileHeader.Size
	// validate file size
	if fileSize > MaxUploadSize {
		return "", errors.New("FILE_TOO_BIG")
	}
	fileBytes, err := io.ReadAll(io.LimitReader(file, MaxUploadSize))
	if err != nil {
		return "", err
	}

	// check file type, detectcontenttype only needs the first 512 bytes
	detectedFileType := http.DetectContentType(fileBytes[:512])
	switch detectedFileType {
	case "image/jpeg", "image/jpg", "image/gif", "image/png", "application/pdf":
		break
	default:
		return "", errors.New("COULD_NOT_DETERMINE_FILE_EXTENSION")
	}
	fileEndings, err := mime.ExtensionsByType(detectedFileType)
	if err != nil || len(fileEndings) == 0 {
		return "", errors.New("INVALID_FILE_TYPE")
	}

	fileName := randToken(12)
	newPath := filepath.Join("assets", "uploads", fileName) + fileEndings[len(fileEndings)-1]
	newFile, err := os.OpenFile(filepath.Clean(newPath), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)

	if err != nil {
		return "", err
	}

	defer newFile.Close()

	// Write the file bytes to the new file
	if _, err := newFile.Write(fileBytes); err != nil {
		return "", err
	}
	return newPath, err
}

const MaxUploadSize = 10 << 20 // 10mb

func randToken(len int) string {
	b := make([]byte, len)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}
