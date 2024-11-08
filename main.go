package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/donseba/go-htmx"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
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
	UserID string
}

type (
	App struct {
		htmx         *htmx.HTMX
		DB           *database.Queries
		DBConn       *pgx.Conn
		Router       chi.Router
		SessionStore sessions.Store
	}
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	port := os.Getenv("PORT")
	conn, err := pgx.Connect(context.Background(), os.Getenv("GOOSE_DBSTRING"))
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	if err != nil {
		panic(err)
	}

	mux := chi.NewRouter()
	// new app with htmx instance
	app := &App{
		htmx:         htmx.New(),
		DB:           database.New(conn),
		DBConn:       conn,
		Router:       mux,
		SessionStore: store,
	}

	htmx.UseTemplateCache = false
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "assets", "uploads"))
	cssDir := http.Dir(filepath.Join(workDir, "assets", "static"))

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
	mux.Get("/v1/users", app.middlewareAuth(app.usersGet))
	mux.Post("/session/new", app.sessionNew)
	FileServer(mux, "/assets/uploads", filesDir)
	FileServer(mux, "/static", cssDir)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,  // set a custom read timeout
		WriteTimeout: 10 * time.Second, // set a custom write timeout
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	h := a.htmx.NewHandler(w, r)

	data := map[string]any{
		"Text": "Welcome home",
	}

	component := htmx.NewComponent("views/home.html").SetData(data)
	page := mainContentWithNavbar("Phamily Photos Home", navbarWithoutUser())
	page.With(component, "Content")

	_, err := h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func (a *App) GetPhotoNew(w http.ResponseWriter, r *http.Request, user database.User) {
	h := a.htmx.NewHandler(w, r)
	data := map[string]any{
		"Title": "Phamily Photos Photo",
	}
	component := htmx.NewComponent("views/photo-new.html")
	page := htmx.NewComponent("views/index.html").SetData(data).With(navbarWithUser(user), "Navbar")
	page.With(component, "Content")

	_, err := h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

type Photo struct {
	PhotoID       string `json:"photo_id"`
	PhotoName     string `json:"photo_name"`
	PhotoUrl      string `json:"photo_url"`
	PhotoThumbUrl string `json:"photo_thumb_url"`
}

func (a *App) GetPhotosIndex(w http.ResponseWriter, r *http.Request, user database.User) {
	h := a.htmx.NewHandler(w, r)
	posts, _ := a.DB.GetPostsByUserFamilyAggregated(r.Context(), database.GetPostsByUserFamilyAggregatedParams{
		FamilyID: user.FamilyID,
		Limit:    10,
	})

	data := map[string]any{
		"Title": "Posts Title",
		"Posts": posts,
	}
	component := htmx.NewComponent("views/posts-index.html", "views/photo.html").SetData(data)
	component.AddTemplateFunction("formatDate", formatDate)
	page := mainContentWithNavbar("Phamily Photos", navbarWithUser(user))
	page.With(component, "Content")
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
		"Photo":      photo,
		"formatDate": formatDate,
	}
	component := htmx.NewComponent("views/photo.html").SetData(data)
	page := mainContentWithNavbar("Phamily Photos Photo", navbarWithUser(user))
	page.With(component, "Content")

	_, err = h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func (a *App) usersGet(w http.ResponseWriter, r *http.Request, user database.User) {
	internal.RespondWithJSON(w, http.StatusOK, UserResponse{
		BaseParams: BaseParams{
			Id:        uuid.MustParse(user.ID),
			CreatedAt: user.CreatedAt.Time.Format(time.DateTime),
			UpdatedAt: user.UpdatedAt.Time.Format(time.DateTime),
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
	session, err := a.SessionStore.Get(r, "session_id")
	if err != nil {
		http.Error(w, "Unable to create session", http.StatusInternalServerError)
		return
	}

	// Store API key and set session expiration (e.g., 60 minutes)
	session.Values["UserID"] = user.ID

	// Save the session
	err = session.Save(r, w)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Unable to save session", http.StatusInternalServerError)
		return
	}
	if contentType == "application/json" {
		internal.RespondWithJSON(w, http.StatusOK, UserResponse{
			BaseParams: BaseParams{
				Id:        uuid.MustParse(user.ID),
				CreatedAt: user.CreatedAt.Time.Format(time.DateTime),
				UpdatedAt: user.UpdatedAt.Time.Format(time.DateTime),
			},
			Name: user.Name,
		})
	} else {
		http.Redirect(w, r, "/photos", http.StatusSeeOther)
	}
}

func (a *App) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		var user database.User
		var uerr error
		if contentType == "" || contentType == "application/x-www-form-urlencoded" || contentType[:19] == "multipart/form-data" {
			session, err := a.SessionStore.Get(r, "session_id")
			if err != nil {
				http.Redirect(w, r, "/login?error=redirected", http.StatusSeeOther)
				return
			}

			err = session.Save(r, w)
			if err != nil {
				http.Error(w, "unable to save session", http.StatusInternalServerError)
				return
			}

			userID, ok := session.Values["UserID"].(string)
			if !ok || userID == "" {
				http.Error(w, "session invalid", http.StatusUnauthorized)
				return
			}

			user, uerr = a.DB.GetUserByID(r.Context(), userID)
			if uerr != nil {
				http.Redirect(w, r, "/login?error=redirected", http.StatusSeeOther)
				return
			}
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
	var form htmx.RenderableComponent
	if err := r.ParseMultipartForm(internal.MaxUploadSize); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["photo"]

	pageData := map[string]any{
		"Title": "Phamily Photos Photo",
	}
	tx, err := a.DBConn.Begin(r.Context())
	if err != nil {
		panic(err)
	}
	defer tx.Rollback(r.Context())
	txq := a.DB.WithTx(tx)

	post, perr := txq.CreatePost(r.Context(), database.CreatePostParams{
		UserID:    user.ID,
		UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		FamilyID:  user.FamilyID.Int64,
	})
	if perr != nil {
		form = uploadFormWithError(w, r, user, perr)
		_, err := h.Render(r.Context(), form)
		if err != nil {
			fmt.Printf("error rendering page: %v", err.Error())
		}
		return
	}
	for fileHeader := range internal.FileGenerator(files) {
		filePath, uploadErr := internal.SaveFile(fileHeader)

		if uploadErr != nil {
			form = uploadFormWithError(w, r, user, uploadErr)
			_, err := h.Render(r.Context(), form)
			if err != nil {
				fmt.Printf("error rendering page: %v", err.Error())
			}
			return
		}
		info, err := os.Stat(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newPath := filepath.Join("/", "assets", "uploads", info.Name())

		_, perr := txq.CreatePhoto(r.Context(), database.CreatePhotoParams{
			ID:       uuid.NewString(),
			UserID:   user.ID,
			Url:      newPath,
			ThumbUrl: newPath,
			ModifiedAt: pgtype.Timestamp{
				Time:             info.ModTime(),
				InfinityModifier: pgtype.Finite,
				Valid:            true,
			},
			Name:    info.Name(),
			AltText: info.Name(),
			PostID:  pgtype.Int8{Int64: post.ID, Valid: true},
		})
		if perr != nil {
			http.Error(w, perr.Error(), http.StatusInternalServerError)
			return
		}
	}
	if err := tx.Commit(r.Context()); err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	posts, _ := a.DB.GetPostsByUserFamilyAggregated(r.Context(), database.GetPostsByUserFamilyAggregatedParams{
		FamilyID: user.FamilyID,
		Limit:    10,
	})
	pageData["Posts"] = posts

	page := htmx.NewComponent("views/index.html").SetData(pageData).
		With(navbarWithUser(user), "Navbar").
		With(htmx.NewComponent("views/posts-index.html").SetData(pageData), "Content")
	if _, herr := h.Render(r.Context(), page); herr != nil {
		fmt.Println(herr.Error())
		http.Error(w, herr.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *App) FamiliesGet(w http.ResponseWriter, r *http.Request, user database.User) {
	h := a.htmx.NewHandler(w, r)
	family, err := a.DB.GetUserFamily(r.Context(), user.ID)
	if err != nil {
		internal.RespondWithErrorHtmx(h, w, http.StatusNotFound, "no family")
	}
	users, _ := a.DB.GetUsersByFamily(r.Context(), user.FamilyID)
	data := map[string]any{
		"Family":        family,
		"FamilyMembers": users,
	}

	component := htmx.NewComponent("views/family.html").SetData(data)
	page := mainContentWithNavbar("Phamily Photos Families", navbarWithUser(user))
	page.With(component, "Content")

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
	session, err := a.SessionStore.Get(r, "session_id")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Invalidate the session by clearing its values
	session.Options.MaxAge = -1 // MaxAge of -1 tells the browser to delete the cookie
	err = session.Save(r, w)    // Save the session changes
	if err != nil {
		http.Error(w, "Unable to logout", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	h := a.htmx.NewHandler(w, r)
	data := map[string]any{
		"RedirectedMessage": r.URL.Query().Get("error") == "redirected",
		"ErrorMessage":      r.URL.Query().Get("error") == "invalid_credentials",
	}
	component := htmx.NewComponent("views/login.html").SetData(data)
	page := mainContentWithNavbar("Phamily Photos Login", navbarWithoutUser())
	page.With(component, "Content")

	_, err := h.Render(r.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func uploadFormWithError(w http.ResponseWriter, _ *http.Request, user database.User, err error) htmx.RenderableComponent {
	formData := map[string]any{
		"Errors": []error{err},
	}
	pageData := map[string]any{
		"Title": "Phamily Photos Photo",
	}
	component := htmx.NewComponent("views/photo-new.html").SetData(formData)
	page := htmx.NewComponent("views/index.html").SetData(pageData).With(navbarWithUser(user), "Navbar")
	page.With(component, "Content")
	w.WriteHeader(http.StatusUnprocessableEntity)
	return page
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

	navbar := htmx.NewComponent("views/navbar.html")
	return navbar.SetData(data)
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

	navbar := htmx.NewComponent("views/navbar.html")
	return navbar.SetData(data)
}

func mainContentWithNavbar(title string, navbar htmx.RenderableComponent) htmx.RenderableComponent {
	data := map[string]any{
		"Title": title,
	}
	return htmx.NewComponent("views/index.html").SetData(data).With(navbar, "Navbar")
}

func formatDate(t time.Time) string {
	return t.Format(time.DateOnly)
}
