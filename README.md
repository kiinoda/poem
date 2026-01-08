# poem

A minimalist blog engine written in Go that serves Markdown posts from S3. Designed for deployment as an AWS Lambda function using the Serverless Framework, which provisions the S3 bucket, custom domain, and DNS records. Also runs locally for development.

## Features

- Markdown posts with TOML frontmatter
- S3-backed content storage
- In-memory caching with metadata-based invalidation
- Custom domain support with automatic DNS configuration
- Gravatar support for author avatars
- GitHub Flavored Markdown rendering

## Deployment (Recommended)

The recommended way to run poem is via AWS Lambda using the Serverless Framework. This provisions:

- Lambda function with Function URL
- S3 bucket for content storage
- Custom domain with Route 53 DNS records
- ACM certificate integration

### Prerequisites

- [Serverless Framework](https://www.serverless.com/framework/docs/getting-started)
- AWS credentials configured
- (Optional) Route 53 hosted zone and ACM certificate for custom domain

### Deploy

```bash
# Build for Lambda
GOOS=linux GOARCH=arm64 go build -o bootstrap .

# Deploy
export S3_BUCKET=my-blog-bucket
export CUSTOM_DOMAIN=blog.example.com        # optional
export HOSTED_ZONE_NAME=example.com          # optional
export CERTIFICATE_ARN=arn:aws:acm:...       # optional
serverless deploy
```

### Environment Variables (Deployment)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `S3_BUCKET` | Yes | - | S3 bucket name for blog content |
| `BLOG_AUTHOR_EMAIL` | No | - | Author email for Gravatar avatar |
| `AWS_REGION` | No | `eu-west-1` | AWS region for deployment |
| `STAGE` | No | `prod` | Deployment stage |
| `CUSTOM_DOMAIN` | No | - | Custom domain for the blog |
| `HOSTED_ZONE_NAME` | No | - | Route 53 hosted zone name |
| `CERTIFICATE_ARN` | No | - | ACM certificate ARN for HTTPS |

## Local Development

For local testing, run the binary directly. Requires AWS credentials with S3 access.

```bash
export S3_BUCKET=my-blog-bucket
export BLOG_AUTHOR_EMAIL=author@example.com
go run .
```

The server starts at `http://localhost:8080`. Override with the `PORT` environment variable.

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
