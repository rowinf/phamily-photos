package internal

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
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
	w.Write(dat)
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

const maxUploadSize = 2 * 1024 * 1024 // 2 mb

func UploadFileHandler(w http.ResponseWriter, r *http.Request, assetPath string) string {
	var path string
	workDir, _ := os.Getwd()
	uploadPath := filepath.Join(workDir, assetPath)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		fmt.Printf("Could not parse multipart form: %v\n", err)
		RespondWithError(w, http.StatusInternalServerError, "CANT_PARSE_FORM")
		return path
	}

	// parse and validate file and post parameters
	file, fileHeader, err := r.FormFile("photo")
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "INVALID_FILE")
		return path
	}
	defer file.Close()
	// Get and print out file size
	fileSize := fileHeader.Size
	fmt.Printf("File size (bytes): %v\n", fileSize)
	// validate file size
	if fileSize > maxUploadSize {
		RespondWithError(w, http.StatusBadRequest, "FILE_TOO_BIG")
		return path
	}
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "INVALID_FILE")
		return path
	}

	// check file type, detectcontenttype only needs the first 512 bytes
	detectedFileType := http.DetectContentType(fileBytes)
	switch detectedFileType {
	case "image/jpeg", "image/jpg":
	case "image/gif", "image/png":
	case "application/pdf":
		break
	default:
		RespondWithError(w, http.StatusBadRequest, "INVALID_FILE_TYPE")
		return path
	}
	fileEndings, err := mime.ExtensionsByType(detectedFileType)
	fileName := randToken(12)

	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "CANT_READ_FILE_TYPE")
		return path
	}

	newFileName := fileName + fileEndings[len(fileEndings)-1]
	newPath := filepath.Join(uploadPath, newFileName)
	fmt.Printf("FileType: %s, File: %s\n", detectedFileType, newPath)

	// write file
	newFile, err := os.Create(newPath)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "CANT_WRITE_FILE")
		return path
	}
	defer newFile.Close() // idempotent, okay to call twice
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		RespondWithError(w, http.StatusInternalServerError, "CANT_WRITE_FILE")
		return path
	}
	return newPath
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
