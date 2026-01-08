# poem

A minimalist blog engine written in Go that serves Markdown posts from S3. Runs locally as an HTTP server or deploys as an AWS Lambda function.

## Features

- Markdown posts with TOML frontmatter
- S3-backed content storage
- In-memory caching with metadata-based invalidation
- Dual-mode: local development server and AWS Lambda
- Gravatar support for author avatars
- GitHub Flavored Markdown rendering

## Installation

```bash
go install github.com/kiinoda/poem@latest
```

Or build from source:

```bash
git clone https://github.com/kiinoda/poem.git
cd poem
go build -o poem .
```

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `S3_BUCKET` | Yes | - | S3 bucket name containing blog posts |
| `BLOG_AUTHOR_EMAIL` | No | - | Author email for Gravatar avatar |
| `PORT` | No | `8080` | HTTP server port (local mode only) |

AWS credentials are loaded via the standard AWS SDK credential chain (environment variables, shared credentials file, IAM role, etc.).

## Usage

### Local Development

```bash
export S3_BUCKET=my-blog-bucket
export BLOG_AUTHOR_EMAIL=author@example.com
./poem
```

The server starts at `http://localhost:8080`.

### AWS Lambda

Deploy the binary as a Lambda function. The application automatically detects Lambda execution via `AWS_LAMBDA_FUNCTION_NAME` and switches to Lambda mode. Configure a Lambda Function URL for HTTP access.

## Content Structure

Posts are stored in S3 under the `posts/` prefix as Markdown files with TOML frontmatter:

```
s3://my-bucket/
  posts/
    my-first-post.md
    another-post.md
    assets/
      image.png
```

### Post Format

```markdown
+++
title = "My Post Title"
date = "2024-01-15"
author = "Author Name"
draft = false
+++

Post content in Markdown...
```

### Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `title` | Yes | Post title |
| `date` | Yes | Publication date (YYYY-MM-DD) |
| `author` | Yes | Author name |
| `draft` | No | Set to `true` to hide from listings |

## Routes

| Path | Description |
|------|-------------|
| `/` | Blog post listing |
| `/{slug}` | Individual post (slug matches filename without `.md`) |
| `/assets/{path}` | Static assets from `posts/assets/` |

## License

MIT
