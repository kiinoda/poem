package domain

import (
	"html/template"
	"time"
)

type Post struct {
	Slug    string
	Title   string
	Author  string
	Date    time.Time
	Draft   bool
	Content template.HTML
	Excerpt string
}
