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

const maxUploadSize = 6 * 1024 * 1024 // 6 mb

func UploadFileHandler(w http.ResponseWriter, r *http.Request, assetPath string) (string, error) {
	workDir, _ := os.Getwd()
	uploadPath := filepath.Join(workDir, assetPath)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		fmt.Printf("Could not parse multipart form: %v\n", err)
		return uploadPath, err
	}

	// parse and validate file and post parameters
	file, fileHeader, err := r.FormFile("photo")
	if err != nil {
		return uploadPath, err
	}
	defer file.Close()
	// Get and print out file size
	fileSize := fileHeader.Size
	fmt.Printf("File size (bytes): %v\n", fileSize)
	// validate file size
	if fileSize > maxUploadSize {
		return uploadPath, errors.New("FILE_TOO_BIG")
	}
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return uploadPath, err
	}

	// check file type, detectcontenttype only needs the first 512 bytes
	detectedFileType := http.DetectContentType(fileBytes)
	switch detectedFileType {
	case "image/jpeg", "image/jpg":
	case "image/gif", "image/png":
	case "application/pdf":
		break
	default:
		return uploadPath, errors.New("INVALID_FILE_TYPE")
	}
	fileEndings, err := mime.ExtensionsByType(detectedFileType)
	fileName := randToken(12)

	if err != nil {
		return uploadPath, err
	}

	newFileName := fileName + fileEndings[len(fileEndings)-1]
	newPath := filepath.Join(uploadPath, newFileName)
	fmt.Printf("FileType: %s, File: %s\n", detectedFileType, newPath)

	// write file
	newFile, err := os.Create(newPath)
	defer newFile.Close() // idempotent, okay to call twice
	if err != nil {
		return newPath, err
	}
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		return newPath, err
	}
	return newPath, err
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
