package config

import (
	"crypto/md5"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	S3Bucket    string
	AuthorEmail string
	Port        string
}

func Load() (*Config, error) {
	cfg := &Config{
		S3Bucket:    os.Getenv("S3_BUCKET"),
		AuthorEmail: os.Getenv("BLOG_AUTHOR_EMAIL"),
		Port:        os.Getenv("PORT"),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}

func (c *Config) GravatarURL() string {
	if c.AuthorEmail == "" {
		return ""
	}
	email := strings.ToLower(strings.TrimSpace(c.AuthorEmail))
	hash := fmt.Sprintf("%x", md5.Sum([]byte(email)))
	return fmt.Sprintf("https://www.gravatar.com/avatar/%s?s=80", hash)
}
