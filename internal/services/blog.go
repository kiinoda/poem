package services

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type BlogPost struct {
	Slug    string
	Title   string
	Author  string
	Date    time.Time
	Draft   bool
	Content template.HTML
	Excerpt string
}

type objectMeta struct {
	LastModified time.Time
	Size         int64
}

type cachedPost struct {
	post *BlogPost
	meta objectMeta
}

type BlogService struct {
	s3Client    *s3.Client
	bucketName  string
	authorEmail string

	cache      map[string]*cachedPost
	listCache  []*BlogPost
	listMeta   map[string]objectMeta
	cacheMutex sync.RWMutex
}

type postFrontmatter struct {
	Title  string `toml:"title"`
	Date   string `toml:"date"`
	Author string `toml:"author"`
	Draft  bool   `toml:"draft"`
}

func NewBlogService(ctx context.Context) (*BlogService, error) {
	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		return nil, fmt.Errorf("S3_BUCKET environment variable is required")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	return &BlogService{
		s3Client:    s3.NewFromConfig(cfg),
		bucketName:  bucketName,
		authorEmail: os.Getenv("BLOG_AUTHOR_EMAIL"),
		cache:       make(map[string]*cachedPost),
		listMeta:    make(map[string]objectMeta),
	}, nil
}

func (b *BlogService) GravatarURL() string {
	if b.authorEmail == "" {
		return ""
	}
	email := strings.ToLower(strings.TrimSpace(b.authorEmail))
	hash := fmt.Sprintf("%x", md5.Sum([]byte(email)))
	return fmt.Sprintf("https://www.gravatar.com/avatar/%s?s=80", hash)
}

func (b *BlogService) ListPosts(ctx context.Context) ([]*BlogPost, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(b.bucketName),
		Prefix: aws.String("posts/"),
	}

	result, err := b.s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts from S3: %w", err)
	}

	// Build current metadata map and check for changes
	currentMeta := make(map[string]objectMeta)
	for _, obj := range result.Contents {
		key := aws.ToString(obj.Key)
		if !strings.HasSuffix(key, ".md") {
			continue
		}
		currentMeta[key] = objectMeta{
			LastModified: aws.ToTime(obj.LastModified),
			Size:         aws.ToInt64(obj.Size),
		}
	}

	b.cacheMutex.RLock()
	hasChanges := !b.metadataMatches(currentMeta)
	cachedList := b.listCache
	b.cacheMutex.RUnlock()

	if !hasChanges && cachedList != nil {
		return cachedList, nil
	}

	// Reload posts that changed or are new
	var posts []*BlogPost
	for key, meta := range currentMeta {
		slug := strings.TrimPrefix(key, "posts/")
		slug = strings.TrimSuffix(slug, ".md")

		b.cacheMutex.RLock()
		cached, exists := b.cache[slug]
		needsReload := !exists || cached.meta != meta
		b.cacheMutex.RUnlock()

		var post *BlogPost
		if needsReload {
			post, err = b.fetchAndParsePost(ctx, key, slug)
			if err != nil {
				fmt.Printf("Warning: failed to parse post %s: %v\n", key, err)
				continue
			}
			b.cacheMutex.Lock()
			b.cache[slug] = &cachedPost{post: post, meta: meta}
			b.cacheMutex.Unlock()
		} else {
			post = cached.post
		}

		if post.Draft {
			continue
		}

		posts = append(posts, post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	b.cacheMutex.Lock()
	b.listCache = posts
	b.listMeta = currentMeta
	b.cacheMutex.Unlock()

	return posts, nil
}

func (b *BlogService) metadataMatches(current map[string]objectMeta) bool {
	if len(b.listMeta) != len(current) {
		return false
	}
	for key, meta := range current {
		if stored, exists := b.listMeta[key]; !exists || stored != meta {
			return false
		}
	}
	return true
}

func (b *BlogService) GetPost(ctx context.Context, slug string) (*BlogPost, error) {
	key := fmt.Sprintf("posts/%s.md", slug)

	// Check current metadata via HeadObject
	headInput := &s3.HeadObjectInput{
		Bucket: aws.String(b.bucketName),
		Key:    aws.String(key),
	}
	headResult, err := b.s3Client.HeadObject(ctx, headInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get post metadata from S3: %w", err)
	}

	currentMeta := objectMeta{
		LastModified: aws.ToTime(headResult.LastModified),
		Size:         aws.ToInt64(headResult.ContentLength),
	}

	b.cacheMutex.RLock()
	cached, exists := b.cache[slug]
	if exists && cached.meta == currentMeta {
		post := cached.post
		b.cacheMutex.RUnlock()
		if post.Draft {
			return nil, nil
		}
		return post, nil
	}
	b.cacheMutex.RUnlock()

	post, err := b.fetchAndParsePost(ctx, key, slug)
	if err != nil {
		return nil, err
	}

	b.cacheMutex.Lock()
	b.cache[slug] = &cachedPost{post: post, meta: currentMeta}
	b.cacheMutex.Unlock()

	if post.Draft {
		return nil, nil
	}

	return post, nil
}

func (b *BlogService) fetchAndParsePost(ctx context.Context, key, slug string) (*BlogPost, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(b.bucketName),
		Key:    aws.String(key),
	}

	result, err := b.s3Client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get post from S3: %w", err)
	}
	defer result.Body.Close()

	content, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read post content: %w", err)
	}

	return b.parsePost(content, slug)
}

func (b *BlogService) parsePost(content []byte, slug string) (*BlogPost, error) {
	parts := bytes.SplitN(content, []byte("+++"), 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid post format: missing frontmatter delimiters")
	}

	var meta postFrontmatter
	if _, err := toml.Decode(string(parts[1]), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	date, err := time.Parse("2006-01-02", meta.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	markdownContent := bytes.TrimSpace(parts[2])
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)

	var htmlBuf bytes.Buffer
	if err := md.Convert(markdownContent, &htmlBuf); err != nil {
		return nil, fmt.Errorf("failed to convert markdown: %w", err)
	}

	excerpt := generateExcerpt(markdownContent, 200)

	return &BlogPost{
		Slug:    slug,
		Title:   meta.Title,
		Author:  meta.Author,
		Date:    date,
		Draft:   meta.Draft,
		Content: template.HTML(htmlBuf.String()),
		Excerpt: excerpt,
	}, nil
}

func (b *BlogService) GetAsset(ctx context.Context, path string) ([]byte, string, error) {
	key := fmt.Sprintf("posts/assets/%s", path)

	input := &s3.GetObjectInput{
		Bucket: aws.String(b.bucketName),
		Key:    aws.String(key),
	}

	result, err := b.s3Client.GetObject(ctx, input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get asset from S3: %w", err)
	}
	defer result.Body.Close()

	content, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read asset content: %w", err)
	}

	contentType := "application/octet-stream"
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	return content, contentType, nil
}

func generateExcerpt(markdown []byte, maxLen int) string {
	text := string(markdown)
	text = strings.ReplaceAll(text, "#", "")
	text = strings.ReplaceAll(text, "*", "")
	text = strings.ReplaceAll(text, "_", "")
	text = strings.ReplaceAll(text, "`", "")
	text = strings.ReplaceAll(text, "[", "")
	text = strings.ReplaceAll(text, "]", "")
	text = strings.ReplaceAll(text, "(", "")
	text = strings.ReplaceAll(text, ")", "")

	lines := strings.Fields(text)
	text = strings.Join(lines, " ")

	if len(text) > maxLen {
		text = text[:maxLen]
		if lastSpace := strings.LastIndex(text, " "); lastSpace > maxLen/2 {
			text = text[:lastSpace]
		}
		text += "..."
	}

	return text
}
