package handlers

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/kiinoda/poem/internal/services"
)

var blogTemplates *template.Template

func InitBlogTemplates(fs embed.FS) error {
	var err error
	blogTemplates, err = template.ParseFS(fs, "templates/*.html")
	return err
}

type BlogListData struct {
	Title       string
	Author      string
	GravatarURL string
	Posts       interface{}
}

type BlogPostData struct {
	Title       string
	Author      string
	GravatarURL string
	Post        interface{}
}

type Handler struct {
	blogService *services.BlogService
}

func New(blogService *services.BlogService) *Handler {
	return &Handler{
		blogService: blogService,
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
		GravatarURL: h.blogService.GravatarURL(),
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
		GravatarURL: h.blogService.GravatarURL(),
		Post:        post,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := blogTemplates.ExecuteTemplate(w, "blog_post.html", data); err != nil {
		log.Printf("Error rendering blog post template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}
