package services

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/kiinoda/poem/internal/content"
	"github.com/kiinoda/poem/internal/domain"
	"github.com/kiinoda/poem/internal/storage"
)

type BlogService struct {
	store  storage.ContentStore
	parser *content.MarkdownParser
	cache  *content.Cache
}

func NewBlogService(store storage.ContentStore) *BlogService {
	return &BlogService{
		store:  store,
		parser: content.NewParser(),
		cache:  content.NewCache(),
	}
}

func (s *BlogService) ListPosts(ctx context.Context) ([]*domain.Post, error) {
	objects, err := s.store.ListObjects(ctx, "posts/")
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	currentMeta := make(map[string]storage.ObjectMeta)
	for _, obj := range objects {
		if !strings.HasSuffix(obj.Key, ".md") {
			continue
		}
		currentMeta[obj.Key] = obj
	}

	if !s.cache.IsListStale(currentMeta) {
		list, _ := s.cache.GetList()
		return list, nil
	}

	var posts []*domain.Post
	for key, meta := range currentMeta {
		slug := strings.TrimPrefix(key, "posts/")
		slug = strings.TrimSuffix(slug, ".md")

		var post *domain.Post

		if s.cache.IsPostStale(slug, meta) {
			post, err = s.fetchAndParsePost(ctx, key, slug)
			if err != nil {
				fmt.Printf("Warning: failed to parse post %s: %v\n", key, err)
				continue
			}
			s.cache.SetPost(slug, post, meta)
		} else {
			cached, _ := s.cache.GetPost(slug)
			post = cached.Post
		}

		if post.Draft {
			continue
		}

		posts = append(posts, post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	s.cache.SetList(posts, currentMeta)

	return posts, nil
}

func (s *BlogService) GetPost(ctx context.Context, slug string) (*domain.Post, error) {
	key := fmt.Sprintf("posts/%s.md", slug)

	meta, err := s.store.GetObjectMeta(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get post metadata: %w", err)
	}

	if !s.cache.IsPostStale(slug, *meta) {
		cached, _ := s.cache.GetPost(slug)
		if cached.Post.Draft {
			return nil, nil
		}
		return cached.Post, nil
	}

	post, err := s.fetchAndParsePost(ctx, key, slug)
	if err != nil {
		return nil, err
	}

	s.cache.SetPost(slug, post, *meta)

	if post.Draft {
		return nil, nil
	}

	return post, nil
}

func (s *BlogService) GetAsset(ctx context.Context, path string) ([]byte, string, error) {
	return s.store.GetAsset(ctx, path)
}

func (s *BlogService) fetchAndParsePost(ctx context.Context, key, slug string) (*domain.Post, error) {
	data, err := s.store.GetObject(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}

	post, err := s.parser.Parse(data, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to parse post: %w", err)
	}

	return post, nil
}
