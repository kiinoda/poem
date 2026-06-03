package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	awslambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/kiinoda/poem/internal/config"
	"github.com/kiinoda/poem/internal/handlers"
	"github.com/kiinoda/poem/internal/lambda"
	"github.com/kiinoda/poem/internal/services"
	"github.com/kiinoda/poem/internal/storage"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		if cfg.S3Bucket == "" {
			log.Fatal("S3_BUCKET environment variable is required in Lambda mode")
		}
		store, err := storage.NewS3Store(ctx, cfg.S3Bucket)
		if err != nil {
			log.Fatalf("Failed to initialize storage: %v", err)
		}
		blogService := services.NewBlogService(store)
		handler := handlers.New(blogService, cfg)
		routes := handler.Routes()
		fmt.Println("Starting Lambda function...")
		adapter := lambda.NewAdapter(routes)
		awslambda.Start(adapter.Handle)
	} else {
		store, err := storage.NewLocalStore(".")
		if err != nil {
			log.Fatalf("Failed to initialize local storage: %v", err)
		}
		blogService := services.NewBlogService(store)
		handler := handlers.New(blogService, cfg)
		routes := handler.Routes()
		fmt.Printf("Starting local server on http://localhost:%s\n", cfg.Port)
		fmt.Println("Serving posts from ./posts (changes picked up on next request)")
		log.Fatal(http.ListenAndServe(":"+cfg.Port, routes))
	}
}
