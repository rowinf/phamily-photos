package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/donseba/go-htmx"
	"github.com/donseba/go-htmx/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rowinf/phamily-photos/internal"
	"github.com/rowinf/phamily-photos/internal/database"
)

type BaseParams struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type UserParams struct {
	BaseParams
	Name   string `json:"name"`
	ApiKey string `json:"apikey"`
}

type PhotoParams struct {
	BaseParams
	ModifiedAt string `json:"modified_at"`
	Name       string `json:"name"`
}

type (
	App struct {
		htmx *htmx.HTMX
		DB   *database.Queries
	}
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	db, err := sql.Open("postgres", os.Getenv("GOOSE_DBSTRING"))
	if err != nil {
		panic(err.Error())
	}
	// new app with htmx instance
	app := &App{
		htmx: htmx.New(),
		DB:   database.New(db),
	}

	mux := chi.NewRouter()

	htmx.UseTemplateCache = false
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "assets", "upload"))

	mux.Use(middleware.MiddleWare)
	mux.Use(chimiddleware.Logger)
	mux.Get("/", app.Home)
	mux.Get("/child", app.Child)
	mux.Get("/photo/new", app.PhotoNew)
	mux.Get("/photo", app.Photo)
	mux.Post("/photo", app.middlewareAuth(app.PhotoCreate))
	mux.Post("/v1/users", app.usersCreate)
	mux.Get("/v1/users", app.middlewareAuth(app.usersGet))
	FileServer(mux, "/assets/upload", filesDir)
	err = http.ListenAndServe(":"+port, mux)
	log.Fatal(err)
}

func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	h := a.htmx.NewHandler(w, r)

	data := map[string]any{
		"Text": "Welcome to the home geiiiii",
	}

	page := htmx.NewComponent("home.html").SetData(data).Wrap(mainContent(), "Content")

	_, err := h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func (a *App) Child(w http.ResponseWriter, r *http.Request) {
	h := a.htmx.NewHandler(w, r)

	data := map[string]any{
		"Text": "Welcome to the child page",
	}

	page := htmx.NewComponent("child.html").SetData(data).Wrap(mainContent(), "Content")

	_, err := h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func (a *App) PhotoNew(w http.ResponseWriter, r *http.Request) {
	h := a.htmx.NewHandler(w, r)
	page := htmx.NewComponent("photo-new.html").Wrap(mainContent(), "Content")

	_, err := h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func (a *App) Photo(w http.ResponseWriter, r *http.Request) {
	h := a.htmx.NewHandler(w, r)
	data := map[string]any{
		"Title": "Photo Title",
		"Url":   "/uploads/Lake-Sherwood1.jpg",
	}
	page := htmx.NewComponent("photo.html").SetData(data).Wrap(mainContent(), "Content")

	_, err := h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func (a *App) usersCreate(w http.ResponseWriter, r *http.Request) {
	body := UserParams{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	payload := database.CreateUserParams{
		Name:      body.Name,
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	user, uerr := a.DB.CreateUser(r.Context(), payload)
	if uerr != nil {
		internal.RespondWithError(w, http.StatusBadRequest, uerr.Error())
		return
	}
	internal.RespondWithJSON(w, http.StatusCreated, UserParams{
		BaseParams: BaseParams{
			Id:        uuid.MustParse(user.ID),
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		},
		Name:   user.Name,
		ApiKey: user.Apikey,
	})
}

func (a *App) usersGet(w http.ResponseWriter, _ *http.Request, user database.User) {
	internal.RespondWithJSON(w, http.StatusOK, UserParams{
		BaseParams: BaseParams{
			Id:        uuid.MustParse(user.ID),
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		},
		ApiKey: user.Apikey,
		Name:   user.Name,
	})
}

func (a *App) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := internal.GetHeaderApiKey(w, r)
		if err != nil {
			internal.RespondWithError(w, http.StatusBadRequest, "no api key")
		} else {
			user, uerr := a.DB.GetUserByApiKey(r.Context(), apiKey)
			if uerr != nil {
				internal.RespondWithError(w, http.StatusBadRequest, "invalid api key")
			} else {
				handler(w, r, user)
			}
		}
	}
}

const maxUploadSize = 2 * 1024 * 1024 // 2 MB
const uploadPath = "./assets/upload"

func (a *App) PhotoCreate(w http.ResponseWriter, r *http.Request, user database.User) {
	h := a.htmx.NewHandler(w, r)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		fmt.Printf("Could not parse multipart form: %v\n", err)
		internal.RespondWithError(w, http.StatusInternalServerError, "CANT_PARSE_FORM")
		return
	}
	file, fileHeader, err := r.FormFile("photo")
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, "INVALID_FILE")
		return
	}
	defer file.Close()
	fileSize := fileHeader.Size
	fmt.Printf("File size (bytes): %v\n", fileSize)
	if fileSize > maxUploadSize {
		internal.RespondWithError(w, http.StatusBadRequest, "FILE_TOO_BIG")
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, "INVALID_FILE")
		return
	}

	detectedFileType := http.DetectContentType(fileBytes)
	switch detectedFileType {
	case "image/jpeg", "image/jpg":
	case "image/gif", "image/png":
	case "application/pdf":
		break
	default:
		internal.RespondWithError(w, http.StatusBadRequest, "INVALID_FILE_TYPE")
		return
	}
	fileName := randToken(12)
	fileEndings, err := mime.ExtensionsByType(detectedFileType)
	if err != nil {
		internal.RespondWithError(w, http.StatusInternalServerError, "CANT_READ_FILE_TYPE")
		return
	}
	newPath := filepath.Join(uploadPath, fileName+fileEndings[0])
	fmt.Printf("FileType: %s, File: %s\n", detectedFileType, newPath)
	// write the file to disk: newPath, fileBytes
	newFile, cerr := os.Create(newPath)
	if cerr != nil {
		internal.RespondWithError(w, http.StatusInternalServerError, "CANT_WRITE_FILE")
		return
	}
	defer newFile.Close()
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		internal.RespondWithError(w, http.StatusInternalServerError, "CANT_WRITE_FILE")
		return
	}
	info, ferr := newFile.Stat()
	if ferr != nil {
		internal.RespondWithError(w, http.StatusInternalServerError, ferr.Error())
	}
	photo, perr := a.DB.CreatePhoto(r.Context(), database.CreatePhotoParams{
		ID:         uuid.NewString(),
		UserID:     user.ID,
		Url:        newPath,
		ThumbUrl:   newPath,
		ModifiedAt: info.ModTime(),
		Name:       newFile.Name(),
		AltText:    newFile.Name(),
	})
	if perr != nil {
		internal.RespondWithError(w, http.StatusInternalServerError, perr.Error())
	}
	data := map[string]any{
		"Photo": photo,
		"Title": "Photo Title",
		"Url":   newPath,
	}
	page := htmx.NewComponent("photo.html").SetData(data).Wrap(mainContent(), "Content")

	_, herr := h.Render(r.Context(), page)
	if herr != nil {
		fmt.Printf("error rendering page: %v", herr.Error())
	}
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusFound).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func mainContent() htmx.RenderableComponent {
	menuItems := []struct {
		Name string
		Link string
	}{
		{"Home", "/"},
		{"Child", "/child"},
	}

	data := map[string]any{
		"Title":     "Home",
		"MenuItems": menuItems,
	}

	sidebar := htmx.NewComponent("sidebar.html")
	return htmx.NewComponent("index.html").SetData(data).With(sidebar, "Sidebar")
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
