package lambda

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

type Adapter struct {
	handler http.Handler
}

func NewAdapter(handler http.Handler) *Adapter {
	return &Adapter{handler: handler}
}

func (a *Adapter) Handle(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
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

	a.handler.ServeHTTP(recorder, req)

	headers := make(map[string]string)
	for key, values := range recorder.Header() {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

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
