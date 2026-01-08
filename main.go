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

	store, err := storage.NewS3Store(ctx, cfg.S3Bucket)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	blogService := services.NewBlogService(store)
	handler := handlers.New(blogService, cfg)
	routes := handler.Routes()

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		fmt.Println("Starting Lambda function...")
		adapter := lambda.NewAdapter(routes)
		awslambda.Start(adapter.Handle)
	} else {
		fmt.Printf("Starting local server on port %s...\n", cfg.Port)
		log.Fatal(http.ListenAndServe(":"+cfg.Port, routes))
	}
}
