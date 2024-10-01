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
	"golang.org/x/crypto/bcrypt"
)

// Create a struct that models the structure of a user, both in the request body, and in the DB
type Credentials struct {
	Password string `json:"password" db:"password"`
	Name     string `json:"name" db:"name"`
}

type BaseParams struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type UserParams struct {
	BaseParams
	Name string `json:"name"`
}

type UserCreateParams struct {
	Credentials
	BaseParams
	FamilyId int64 `json:"family_id"`
}

type UserLoginParams struct {
	Credentials
	BaseParams
	ApiKey string `json:"apikey"`
}

type UserResponse struct {
	BaseParams
	Name string `json:"name" db:"name"`
}

type PhotoParams struct {
	BaseParams
	ModifiedAt string `json:"modified_at"`
	Name       string `json:"name"`
}

type Session struct {
	ApiKey    string
	ExpiresAt time.Time
}

type (
	App struct {
		htmx   *htmx.HTMX
		DB     *database.Queries
		Router chi.Router
	}
)

var sessionStore = map[string]Session{} // In-memory session store

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
	mux.Get("/login", app.Login)
	mux.Get("/logout", app.Logout)
	mux.Get("/photos", app.middlewareAuth(app.GetPhotosIndex))
	mux.Get("/photos/new", app.middlewareAuth(app.GetPhotoNew))
	mux.Get("/family", app.middlewareAuth(app.FamiliesGet))
	mux.Delete("/photos/{photoID}", app.middlewareAuth(app.DeletePhoto))
	mux.Get("/photos/{photoID}", app.middlewareAuth(app.GetPhoto))
	mux.Post("/photos", app.middlewareAuth(app.PhotoCreate))
	mux.Post("/v1/users", app.usersCreate)
	mux.Get("/v1/users", app.middlewareAuth(app.usersGet))
	mux.Post("/session/new", app.sessionNew)
	FileServer(mux, "/assets/uploads", filesDir)
	err = http.ListenAndServe(":"+port, mux)
	log.Fatal(err)
}

func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	h := a.htmx.NewHandler(w, r)

	data := map[string]any{
		"Text": "Welcome home",
	}

	page := htmx.NewComponent("views/home.html").SetData(data).Wrap(mainContent("Phamily Photos Home", navbarWithoutUser()), "Content")

	_, err := h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func (a *App) GetPhotoNew(w http.ResponseWriter, r *http.Request, user database.User) {
	h := a.htmx.NewHandler(w, r)
	page := htmx.NewComponent("views/photo-new.html").Wrap(mainContent("Phamily Photos Photo", navbarWithUser(user)), "Content")

	_, err := h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func (a *App) GetPhotosIndex(w http.ResponseWriter, r *http.Request, user database.User) {
	h := a.htmx.NewHandler(w, r)
	photos, _ := a.DB.GetPhotosByUserFamily(r.Context(), database.GetPhotosByUserFamilyParams{
		UserID: user.ID,
		Limit:  10,
	})

	data := map[string]any{
		"Title":      "Photos Title",
		"Url":        "/assets/uploads/Lake-Sherwood1.jpg",
		"Photos":     photos,
		"FormatDate": formatDate,
	}
	page := htmx.NewComponent("views/photos-index.html").SetData(data).Wrap(mainContent("Phamily Photos", navbarWithUser(user)), "Content")

	_, err := h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func (a *App) DeletePhoto(w http.ResponseWriter, r *http.Request, user database.User) {
	err := a.DB.DeletePhoto(r.Context(), database.DeletePhotoParams{
		ID:     r.PathValue("photoID"),
		UserID: user.ID,
	})
	if err != nil {
		internal.RespondWithError(w, http.StatusNotFound, "not found")
		return
	}
	internal.RespondWithOk(w)
}

func (a *App) GetPhoto(w http.ResponseWriter, r *http.Request, user database.User) {
	h := a.htmx.NewHandler(w, r)
	photo, err := a.DB.GetPhoto(r.Context(), database.GetPhotoParams{
		ID:     r.PathValue("photoID"),
		UserID: user.ID,
	})
	if err != nil {
		internal.RespondWithError(w, http.StatusNotFound, "not found")
		return
	}
	data := map[string]any{
		"User":       user,
		"Photo":      photo,
		"IsMyPhoto":  photo.IsMyPhoto,
		"FormatDate": formatDate,
	}
	page := htmx.NewComponent("views/photo.html").SetData(data).Wrap(mainContent("Phamily Photos Photo", navbarWithUser(user)), "Content")

	_, err = h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func (a *App) usersCreate(w http.ResponseWriter, r *http.Request) {
	body := UserCreateParams{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	hashedPassword, perr := bcrypt.GenerateFromPassword([]byte(body.Password), 8)
	if perr != nil {
		internal.RespondWithError(w, http.StatusBadRequest, perr.Error())
	}

	if user, uerr := a.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:       uuid.NewString(),
		Name:     body.Name,
		Password: string(hashedPassword),
		FamilyID: sql.NullInt64{1, true},
	}); uerr != nil {
		// If there is any issue with inserting into the database, return a 500 error
		internal.RespondWithError(w, http.StatusInternalServerError, uerr.Error())
		return
	} else {
		internal.RespondWithJSON(w, http.StatusCreated, UserResponse{
			BaseParams: BaseParams{
				Id:        uuid.MustParse(user.ID),
				CreatedAt: user.CreatedAt.Format(time.DateTime),
				UpdatedAt: user.UpdatedAt.Format(time.DateTime),
			},
			Name: user.Name,
		})
	}
}

func (a *App) usersGet(w http.ResponseWriter, r *http.Request, user database.User) {
	internal.RespondWithJSON(w, http.StatusOK, UserResponse{
		BaseParams: BaseParams{
			Id:        uuid.MustParse(user.ID),
			CreatedAt: user.CreatedAt.Format(time.DateTime),
			UpdatedAt: user.UpdatedAt.Format(time.DateTime),
		},
		Name: user.Name,
	})
}

func (a *App) sessionNew(w http.ResponseWriter, r *http.Request) {
	body := UserLoginParams{}
	h := a.htmx.NewHandler(w, r)
	contentType := r.Header.Get("Content-Type")
	if h.IsHxRequest() {
		http.Error(w, "htmx auth not allowed", http.StatusBadRequest)
		return
	}
	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			// If there is something wrong with the request body, return a 400 status
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else if contentType == "application/x-www-form-urlencoded" || contentType == "multipart/form-data" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form data", http.StatusBadRequest)
			return
		}
		body.Name = r.FormValue("username")
		body.Password = r.FormValue("password")
	}
	user, err := a.DB.GetUserByName(r.Context(), body.Name)
	if err != nil {
		// If there is an issue with the database, return a 500 error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		// If the two passwords don't match, return a 401 status
		// http.Error(w, err.Error(), http.StatusUnauthorized)
		http.Redirect(w, r, "/login?error=invalid_credentials", http.StatusSeeOther)
		return
	}
	apikey := user.Apikey

	// Generate a unique session ID
	sessionID := uuid.NewString()

	// Set session expiration (e.g., 30 minutes)
	expiration := time.Now().Add(60 * time.Minute)

	// Store the session in the server (in-memory for this demo)
	sessionStore[sessionID] = Session{
		ApiKey:    apikey,
		ExpiresAt: expiration,
	}

	// Set session ID in a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  expiration,
		HttpOnly: true, // Makes cookie inaccessible to JavaScript for security
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	if contentType == "application/json" {
		internal.RespondWithJSON(w, http.StatusOK, UserResponse{
			BaseParams: BaseParams{
				Id:        uuid.MustParse(user.ID),
				CreatedAt: user.CreatedAt.Format(time.DateTime),
				UpdatedAt: user.UpdatedAt.Format(time.DateTime),
			},
			Name: user.Name,
		})
	} else if contentType == "application/x-www-form-urlencoded" || contentType == "multipart/form-data" {
		http.Redirect(w, r, "/photos", http.StatusSeeOther)
	}
}

func (a *App) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		var user database.User
		var uerr error
		if contentType == "application/json" {
			apikey, err := internal.GetHeaderApiKey(w, r)
			user, uerr = a.DB.GetUserByApiKey(r.Context(), apikey)
			if err != nil {
				http.Error(w, "no api key", http.StatusUnauthorized)
				return
			}
		} else if contentType == "" || contentType == "application/x-www-form-urlencoded" || contentType[:19] == "multipart/form-data" {
			cookie, err := r.Cookie("session_id")
			if err != nil {
				http.Redirect(w, r, "/login?error=redirected", http.StatusSeeOther)
				return
			}
			sessionID := cookie.Value
			session, exists := sessionStore[sessionID]

			if !exists || session.ExpiresAt.Before(time.Now()) {
				http.Redirect(w, r, "/login?error=redirected", http.StatusSeeOther)
				return
			}
			session.ExpiresAt = time.Now().Add(30 * time.Minute)
			sessionStore[sessionID] = session
			user, uerr = a.DB.GetUserByApiKey(r.Context(), session.ApiKey)
		}
		if uerr != nil {
			http.Error(w, "session invalid", http.StatusUnauthorized)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, perr.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]any{
		"User":       user,
		"Photo":      photo,
		"FormatDate": formatDate,
	}
	page := htmx.NewComponent("views/photo.html").SetData(data).Wrap(mainContent("Phamily Photos Photo", navbarWithUser(user)), "Content")

	_, err = h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func (a *App) FamiliesGet(w http.ResponseWriter, r *http.Request, user database.User) {
	h := a.htmx.NewHandler(w, r)
	family, err := a.DB.GetUserFamily(r.Context(), user.ID)
	if err != nil {
		internal.RespondWithErrorHtmx(h, w, http.StatusNotFound, "no family")
	}
	nullFamilyId := sql.NullInt64{Int64: user.FamilyID.Int64, Valid: true}
	users, _ := a.DB.GetUsersByFamily(r.Context(), nullFamilyId)
	data := map[string]any{
		"Family":        family,
		"FamilyMembers": users,
	}
	page := htmx.NewComponent("views/family.html").SetData(data).Wrap(mainContent("Phamily Photos Families", navbarWithUser(user)), "Content")

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

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")

	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	sessionID := cookie.Value
	session, exists := sessionStore[sessionID]
	if !exists || session.ExpiresAt.Before(time.Now()) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	session.ExpiresAt = time.Now()
	sessionStore[sessionID] = session
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	h := a.htmx.NewHandler(w, r)
	data := map[string]any{
		"RedirectedMessage": r.URL.Query().Get("error") == "redirected",
		"ErrorMessage":      r.URL.Query().Get("error") == "invalid_credentials",
	}
	page := htmx.NewComponent("views/login.html").SetData(data).Wrap(mainContent("Phamily Photos Login", navbarWithoutUser()), "Content")

	_, err := h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func navbarWithoutUser() htmx.RenderableComponent {
	menuItems := []struct {
		Name      string
		Link      string
		BoostAttr string
	}{
		{"Home", "/", "true"},
		{"Photos", "/photos", "true"},
		{"Login", "/login", "true"},
	}
	data := map[string]any{
		"MenuItems": menuItems,
	}

	sidebar := htmx.NewComponent("views/sidebar.html")
	return sidebar.SetData(data)
}

func navbarWithUser(user database.User) htmx.RenderableComponent {
	menuItems := []struct {
		Name      string
		Link      string
		BoostAttr string
	}{
		{"Home", "/", "true"},
		{"Photos", "/photos", "true"},
		{"New", "/photos/new", "true"},
		{"Family", "/family", "true"},
		{"Logout", "/logout", "false"},
	}
	data := map[string]any{
		"User":      user,
		"MenuItems": menuItems,
	}

	sidebar := htmx.NewComponent("views/sidebar.html")
	return sidebar.SetData(data)
}

func mainContent(title string, navbar htmx.RenderableComponent) htmx.RenderableComponent {
	data := map[string]any{
		"Title": title,
	}
	return htmx.NewComponent("views/index.html").SetData(data).With(navbar, "Sidebar")
}

func formatDate(t time.Time) string {
	return t.Format(time.DateOnly)
}
