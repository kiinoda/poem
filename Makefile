.PHONY: help build deploy clean posts run invalidate

help:
	@echo "Poem - Dynamic Blog Engine"
	@echo ""
	@echo "Available targets:"
	@echo "  help       - Show this help message"
	@echo "  build      - Build Lambda function"
	@echo "  run        - Run locally"
	@echo "  deploy     - Build and deploy to AWS"
	@echo "  posts      - Sync posts to S3"
	@echo "  invalidate - Invalidate CloudFront cache"
	@echo "  clean      - Remove build artifacts"

build:
	@echo "Building Lambda function..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap .
	@echo "Build complete: bootstrap"

run:
	go run main.go

deploy: build
	@echo "Deploying to AWS Lambda..."
	npx serverless deploy
	@rm -f bootstrap
	@echo "Deployment complete"

clean:
	@echo "Cleaning build artifacts..."
	@rm -f bootstrap
	@rm -rf .serverless
	@echo "Clean complete"

posts:
	@echo "Syncing posts to S3..."
	aws s3 sync posts s3://$$S3_BUCKET/posts --delete
	@echo "Posts synced"

invalidate:
	@echo "Invalidating CloudFront cache..."
	aws cloudfront create-invalidation --distribution-id $$CLOUDFRONT_DISTRIBUTION --paths "/*"
	@echo "Invalidation requested"
