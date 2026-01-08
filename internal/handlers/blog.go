package handlers

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/kiinoda/poem/internal/config"
	"github.com/kiinoda/poem/internal/domain"
	"github.com/kiinoda/poem/internal/services"
)

//go:embed templates/*
var templatesFS embed.FS

var blogTemplates *template.Template

func init() {
	var err error
	blogTemplates, err = template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		panic("failed to parse templates: " + err.Error())
	}
}

type BlogListData struct {
	Title       string
	Author      string
	GravatarURL string
	Posts       []*domain.Post
}

type BlogPostData struct {
	Title       string
	Author      string
	GravatarURL string
	Post        *domain.Post
}

type Handler struct {
	blogService *services.BlogService
	config      *config.Config
}

func New(blogService *services.BlogService, cfg *config.Config) *Handler {
	return &Handler{
		blogService: blogService,
		config:      cfg,
	}
}

func (h *Handler) BlogList(w http.ResponseWriter, r *http.Request) {
	if h.blogService == nil {
		http.Error(w, "Blog service not configured", http.StatusServiceUnavailable)
		return
	}

	posts, err := h.blogService.ListPosts(r.Context())
	if err != nil {
		log.Printf("Error listing blog posts: %v", err)
		http.Error(w, "Failed to load blog posts", http.StatusInternalServerError)
		return
	}

	author := "Blog"
	if len(posts) > 0 {
		author = posts[0].Author
	}

	data := BlogListData{
		Title:       "",
		Author:      author,
		GravatarURL: h.config.GravatarURL(),
		Posts:       posts,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := blogTemplates.ExecuteTemplate(w, "blog_list.html", data); err != nil {
		log.Printf("Error rendering blog list template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Asset(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	content, contentType, err := h.blogService.GetAsset(r.Context(), path)
	if err != nil {
		log.Printf("Error getting asset %s: %v", path, err)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(content)
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", h.BlogList)
	mux.HandleFunc("GET /assets/{path...}", h.Asset)
	mux.HandleFunc("GET /{slug}", h.BlogPost)
	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) BlogPost(w http.ResponseWriter, r *http.Request) {
	if h.blogService == nil {
		http.Error(w, "Blog service not configured", http.StatusServiceUnavailable)
		return
	}

	slug := r.PathValue("slug")
	if slug == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	post, err := h.blogService.GetPost(r.Context(), slug)
	if err != nil {
		log.Printf("Error getting blog post %s: %v", slug, err)
		http.Error(w, "Failed to load blog post", http.StatusInternalServerError)
		return
	}

	if post == nil {
		http.NotFound(w, r)
		return
	}

	data := BlogPostData{
		Title:       post.Title,
		Author:      post.Author,
		GravatarURL: h.config.GravatarURL(),
		Post:        post,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := blogTemplates.ExecuteTemplate(w, "blog_post.html", data); err != nil {
		log.Printf("Error rendering blog post template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}
