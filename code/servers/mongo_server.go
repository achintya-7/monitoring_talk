package servers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"monitoring-talk/logger"
	"monitoring-talk/telemetry"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	mongoLogger      logger.CustomLogger
	client           *mongo.Client
	usersCollection  *mongo.Collection
	mongoOtelContext *telemetry.OtelContext
)

func init() {
	var err error

	// init otel
	mongoOtelContext, err = telemetry.NewOtelContext(ALLOY_ENDPOINT, ALLOY_TOKEN, MONGO_SERVICE_NAME)
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
	}

	// init logger
	mongoLogger = logger.NewOtelLogger(mongoOtelContext)
	mongoLogger.Info(context.Background(), "Initializing mongo server")

	opts := options.Client()
	opts.ApplyURI(MONGO_URL)
	opts.Monitor = mongoOtelContext.GetMongoDefaultHook()

	// Connect to MongoDB
	client, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	// Get a handle for the users collection
	usersCollection = client.Database(USER_DATABASE).Collection(USERS_COLLECTION)

	mongoLogger.Info(context.Background(), "Connected to MongoDB")
}

func InitMongoServer() {
	// disable gin logging
	gin.DefaultWriter = io.Discard

	router := gin.Default()

	// attach otel middleware
	router.Use(mongoOtelContext.GetGinMiddleware())
	// attach custom middleware
	router.Use(CorrelationIdMiddleware())

	router.POST("/user", func(ctx *gin.Context) {
		var req InsertUserRequest
		reqCtx := ctx.Request.Context()
		correlationId := ctx.GetString(CORRELATION_ID)

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := usersCollection.InsertOne(reqCtx, bson.M{"username": req.Username})
		if err != nil {
			mongoLogger.Error(ctx, "Error inserting user: "+err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert user"})
			return
		}

		mongoLogger.Info(ctx, "User addedv to mongo successfully")

		// Call postgres_data API
		mongoLogger.Info(ctx, "Calling postgres_data API")

		// Call mongo_data API
		reqBody := gin.H{
			"username": req.Username,
		}

		reqJsonBody, err := json.Marshal(reqBody)
		if err != nil {
			mongoLogger.Error(ctx, "Error marshalling request body: "+err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request body"})
			return
		}

		// make a http client with otel transport
		httpClient := http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		}

		apiReq, err := http.NewRequestWithContext(reqCtx, "POST", "http://localhost:8082/user", strings.NewReader(string(reqJsonBody)))
		if err != nil {
			mongoLogger.Error(ctx, "Error creating request for postgres_data API: "+err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request for postgres_data API"})
			return
		}

		apiReq.Header.Set("Content-Type", "application/json")
		apiReq.Header.Set("x-correlation-id", correlationId)

		postgresResp, err := httpClient.Do(apiReq)
		if err != nil {
			mongoLogger.Error(ctx, "Error calling postgres_data API: "+err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call postgres_data API"})
			return
		}

		if postgresResp.StatusCode != http.StatusOK {
			mongoLogger.Error(ctx, "Error calling postgres_data API: "+postgresResp.Status)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call postgres_data API"})
			return
		}

		defer postgresResp.Body.Close()

		ctx.String(http.StatusOK, "User added successfully")
	})

	router.Run(":8081")
}
