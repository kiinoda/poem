package content

import (
	"sync"

	"github.com/kiinoda/poem/internal/domain"
	"github.com/kiinoda/poem/internal/storage"
)

type CachedPost struct {
	Post *domain.Post
	Meta storage.ObjectMeta
}

type Cache struct {
	posts    map[string]*CachedPost
	listMeta map[string]storage.ObjectMeta
	list     []*domain.Post
	mu       sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		posts:    make(map[string]*CachedPost),
		listMeta: make(map[string]storage.ObjectMeta),
	}
}

func (c *Cache) GetPost(slug string) (*CachedPost, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cached, exists := c.posts[slug]
	return cached, exists
}

func (c *Cache) SetPost(slug string, post *domain.Post, meta storage.ObjectMeta) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.posts[slug] = &CachedPost{Post: post, Meta: meta}
}

func (c *Cache) GetList() ([]*domain.Post, map[string]storage.ObjectMeta) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.list, c.listMeta
}

func (c *Cache) SetList(posts []*domain.Post, meta map[string]storage.ObjectMeta) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.list = posts
	c.listMeta = meta
}

func (c *Cache) IsListStale(currentMeta map[string]storage.ObjectMeta) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.list == nil {
		return true
	}

	if len(c.listMeta) != len(currentMeta) {
		return true
	}

	for key, meta := range currentMeta {
		stored, exists := c.listMeta[key]
		if !exists || stored.LastModified != meta.LastModified || stored.Size != meta.Size {
			return true
		}
	}

	return false
}

func (c *Cache) IsPostStale(slug string, currentMeta storage.ObjectMeta) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.posts[slug]
	if !exists {
		return true
	}

	return cached.Meta.LastModified != currentMeta.LastModified || cached.Meta.Size != currentMeta.Size
}
