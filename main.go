package main

import (
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kiinoda/poem/internal/handlers"
	"github.com/kiinoda/poem/internal/services"
)

//go:embed templates/*
var templatesFS embed.FS

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func setupRoutes(handler *handlers.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", handler.BlogList)
	mux.HandleFunc("GET /assets/{path...}", handler.Asset)
	mux.HandleFunc("GET /{slug}", handler.BlogPost)

	return corsMiddleware(mux)
}

var globalHandler http.Handler

func lambdaHandler(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	path := request.RequestContext.HTTP.Path
	if path == "" {
		path = "/"
	}

	log.Printf("Lambda request - Path: %s, Method: %s", path, request.RequestContext.HTTP.Method)

	req, err := http.NewRequestWithContext(ctx, request.RequestContext.HTTP.Method, path, strings.NewReader(request.Body))
	if err != nil {
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Headers:    map[string]string{"Content-Type": "text/html"},
			Body:       "Internal server error",
		}, nil
	}

	for key, value := range request.Headers {
		req.Header.Set(key, value)
	}

	if request.RawQueryString != "" {
		req.URL.RawQuery = request.RawQueryString
	}

	recorder := &responseRecorder{
		statusCode: 200,
		body:       make([]byte, 0),
	}

	globalHandler.ServeHTTP(recorder, req)

	headers := make(map[string]string)
	for key, values := range recorder.Header() {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Check if response is binary based on Content-Type
	contentType := headers["Content-Type"]
	isBinary := !strings.HasPrefix(contentType, "text/") &&
		!strings.HasPrefix(contentType, "application/json")

	var body string
	if isBinary {
		body = base64.StdEncoding.EncodeToString(recorder.body)
	} else {
		body = string(recorder.body)
	}

	return events.LambdaFunctionURLResponse{
		StatusCode:      recorder.statusCode,
		Headers:         headers,
		Body:            body,
		IsBase64Encoded: isBinary,
	}, nil
}

type responseRecorder struct {
	statusCode int
	header     http.Header
	body       []byte
}

func (r *responseRecorder) Header() http.Header {
	if r.header == nil {
		r.header = make(http.Header)
	}
	return r.header
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	r.body = append(r.body, data...)
	return len(data), nil
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func main() {
	ctx := context.Background()

	// Initialize templates
	if err := handlers.InitBlogTemplates(templatesFS); err != nil {
		log.Fatalf("Failed to initialize templates: %v", err)
	}

	// Initialize blog service
	blogService, err := services.NewBlogService(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize blog service: %v", err)
	}

	handler := handlers.New(blogService)

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		fmt.Println("Starting Lambda function...")
		globalHandler = setupRoutes(handler)
		lambda.Start(lambdaHandler)
	} else {
		fmt.Println("Starting local server...")
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		httpHandler := setupRoutes(handler)
		fmt.Printf("Server starting on port %s\n", port)
		log.Fatal(http.ListenAndServe(":"+port, httpHandler))
	}
}
