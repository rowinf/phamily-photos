package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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
		htmx   *htmx.HTMX
		DB     *database.Queries
		Router chi.Router
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

	mux := chi.NewRouter()
	// new app with htmx instance
	app := &App{
		htmx:   htmx.New(),
		DB:     database.New(db),
		Router: mux,
	}

	htmx.UseTemplateCache = false
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "assets", "uploads"))

	mux.Use(middleware.MiddleWare)
	mux.Use(chimiddleware.Logger)
	mux.Get("/", app.Home)
	mux.Get("/child", app.Child)
	mux.Get("/photos", app.middlewareAuth(app.GetPhotosIndex))
	mux.Get("/photos/new", app.GetPhotoNew)
	mux.Get("/photos/{photoID}", app.GetPhoto)
	mux.Post("/photos", app.middlewareAuth(app.PhotoCreate))
	mux.Post("/v1/users", app.usersCreate)
	mux.Get("/v1/users", app.middlewareAuth(app.usersGet))
	FileServer(mux, "/assets/uploads", filesDir)
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

func (a *App) GetPhotoNew(w http.ResponseWriter, r *http.Request) {
	h := a.htmx.NewHandler(w, r)
	page := htmx.NewComponent("photo-new.html").Wrap(mainContent(), "Content")

	_, err := h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func (a *App) GetPhotosIndex(w http.ResponseWriter, r *http.Request, user database.User) {
	h := a.htmx.NewHandler(w, r)
	photos, _ := a.DB.GetPhotosByUser(r.Context(), database.GetPhotosByUserParams{
		ID:    user.ID,
		Limit: 10,
	})
	data := map[string]any{
		"Title":  "Photos Title",
		"Url":    "/assets/uploads/Lake-Sherwood1.jpg",
		"Photos": photos,
	}
	page := htmx.NewComponent("views/photos-index.html").SetData(data).Wrap(mainContent(), "Content")

	_, err := h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func (a *App) GetPhoto(w http.ResponseWriter, r *http.Request) {
	h := a.htmx.NewHandler(w, r)
	data := map[string]any{
		"Title": "Photo Title",
		"Url":   "/assets/uploads/Lake-Sherwood1.jpg",
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
			internal.RespondWithError(w, http.StatusUnauthorized, "no api key")
			return
		}
		user, uerr := a.DB.GetUserByApiKey(r.Context(), apiKey)
		if uerr != nil {
			internal.RespondWithError(w, http.StatusUnauthorized, "invalid api key")
			return
		}
		handler(w, r, user)
	}
}

func (a *App) PhotoCreate(w http.ResponseWriter, r *http.Request, user database.User) {
	h := a.htmx.NewHandler(w, r)
	assetPath := filepath.Join("assets", "uploads")
	sysPath := internal.UploadFileHandler(w, r, assetPath)
	info, err := os.Stat(sysPath)
	if err != nil || info == nil {
		internal.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
	paths := strings.Split(sysPath, "/")
	fileName := paths[len(paths)-1]
	newPath := filepath.Join("/", assetPath, fileName)

	photo, perr := a.DB.CreatePhoto(r.Context(), database.CreatePhotoParams{
		ID:         uuid.NewString(),
		UserID:     user.ID,
		Url:        newPath,
		ThumbUrl:   newPath,
		ModifiedAt: info.ModTime(),
		Name:       info.Name(),
		AltText:    info.Name(),
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
		{"Photos", "/photos"},
		{"New", "/photos/new"},
	}

	data := map[string]any{
		"Title":     "Home",
		"MenuItems": menuItems,
	}

	sidebar := htmx.NewComponent("sidebar.html")
	return htmx.NewComponent("index.html").SetData(data).With(sidebar, "Sidebar")
}
