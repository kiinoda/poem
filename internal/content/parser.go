package content

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/kiinoda/poem/internal/domain"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type frontmatter struct {
	Title  string `toml:"title"`
	Date   string `toml:"date"`
	Author string `toml:"author"`
	Draft  bool   `toml:"draft"`
}

type Parser interface {
	Parse(raw []byte, slug string) (*domain.Post, error)
}

type MarkdownParser struct {
	md goldmark.Markdown
}

func NewParser() *MarkdownParser {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)
	return &MarkdownParser{md: md}
}

func (p *MarkdownParser) Parse(raw []byte, slug string) (*domain.Post, error) {
	parts := bytes.SplitN(raw, []byte("+++"), 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid post format: missing frontmatter delimiters")
	}

	var meta frontmatter
	if _, err := toml.Decode(string(parts[1]), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	date, err := time.Parse("2006-01-02", meta.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %w", err)
	}

	markdownContent := bytes.TrimSpace(parts[2])

	var htmlBuf bytes.Buffer
	if err := p.md.Convert(markdownContent, &htmlBuf); err != nil {
		return nil, fmt.Errorf("failed to convert markdown: %w", err)
	}

	excerpt := generateExcerpt(markdownContent, 200)

	return &domain.Post{
		Slug:    slug,
		Title:   meta.Title,
		Author:  meta.Author,
		Date:    date,
		Draft:   meta.Draft,
		Content: template.HTML(htmlBuf.String()),
		Excerpt: excerpt,
	}, nil
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
