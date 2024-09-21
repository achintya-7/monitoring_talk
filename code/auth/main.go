package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	// Alloy
	ALLOY_ENDPOINT = "localhost:9001"
	ALLOY_TOKEN    = "G7kLw9xYtQpZrV3mN2sHjXcB8aWfUe5K"

	// Services
	AUTH_SERVICE_NAME = "AUTH_SERVICE"
)

var (
	authOtelContext *OtelContext
	authLogger *CustomLogger
)

type InsertUserRequest struct {
	Username string `json:"username" binding:"required"`
}

func init() {
	var err error

	time.Sleep(1 * time.Second)

	// init otel
	authOtelContext, err = NewOtelContext(ALLOY_ENDPOINT, ALLOY_TOKEN, AUTH_SERVICE_NAME)
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
	}

	// init logger
	authLogger = NewOtelLogger(authOtelContext)
	authLogger.Info(context.Background(), "Starting Auth Server")
}

func main() {
	// disable gin logging
	gin.DefaultWriter = io.Discard

	router := gin.Default()

	// attach otel middleware
	router.Use(authOtelContext.GetGinMiddleware())
	// attach custom middleware
	router.Use(CorrelationIdMiddleware())

	router.POST("/auth", func(ctx *gin.Context) {
		authLogger.Info(ctx, "Got req for Auth Service")

		var req InsertUserRequest
		correlationId := ctx.GetString(CORRELATION_ID)

		if err := ctx.BindJSON(&req); err != nil {
			authLogger.Error(ctx, "Error binding request: "+err.Error())
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Call mongo_data API
		reqBody := gin.H{
			"username": req.Username,
		}

		reqJsonBody, err := json.Marshal(reqBody)
		if err != nil {
			authLogger.Error(ctx, "Error marshalling request body: "+err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request body"})
			return
		}

		authLogger.Info(ctx, "Calling mongo_data API")
		apiReq, _ := http.NewRequestWithContext(ctx.Request.Context(), "POST", "http://localhost:8081/user", strings.NewReader(string(reqJsonBody)))

		apiReq.Header.Set("Content-Type", "application/json")
		apiReq.Header.Set("x-correlation-id", correlationId)

		// make a http client with otel transport
		httpClient := http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		}

		mongoResp, err := httpClient.Do(apiReq)
		if err != nil {
			authLogger.Error(ctx, "Error calling mongo_data API: "+err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call mongo_data API"})
			return
		}

		if mongoResp.StatusCode != http.StatusOK {
			authLogger.Error(ctx, "Error calling mongo_data API: "+mongoResp.Status)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call mongo_data API"})
			return
		}

		defer mongoResp.Body.Close()

		authLogger.Info(ctx, "Data added successfully")

		ctx.String(http.StatusOK, "OK")
	})

	router.Run(":8080")
}
