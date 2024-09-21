package main

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/exaring/otelpgx"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// Alloy
	ALLOY_ENDPOINT = "localhost:9001"
	ALLOY_TOKEN    = "G7kLw9xYtQpZrV3mN2sHjXcB8aWfUe5K"

	// Services
	POSTGRES_SERVICE_NAME = "POSTGRES_SERVICE"

	POSTGRES_URL = "postgres://postgres:postgres@localhost:5432/demo"
)

type InsertUserRequest struct {
	Username string `json:"username" binding:"required"`
}

var (
	pgLogger      *CustomLogger
	pgcConn       *pgxpool.Pool
	pgOtelContext *OtelContext
)

func init() {
	var err error

	// init otel
	pgOtelContext, err = NewOtelContext(ALLOY_ENDPOINT, ALLOY_TOKEN, POSTGRES_SERVICE_NAME)
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
	}

	// init logger
	pgLogger = NewOtelLogger(pgOtelContext)
	pgLogger.Info(context.Background(), "Initializing postgres server")

	pgxCfg, err := pgxpool.ParseConfig(POSTGRES_URL)
	if err != nil {
		log.Fatalf("parse database URL: %v", err)
	}

	pgxCfg.MaxConns = 99

	pgxCfg.ConnConfig.Tracer = otelpgx.NewTracer()

	pgcConn, err = pgxpool.NewWithConfig(context.Background(), pgxCfg)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	err = pgcConn.Ping(context.Background())
	if err != nil {
		log.Fatalf("ping database: %v", err)
	}

	pgLogger.Info(context.Background(), "Connected to Postgres")
}

func main() {
	// disable gin logging
	gin.DefaultWriter = io.Discard

	router := gin.Default()

	// attach otel middleware
	router.Use(pgOtelContext.GetGinMiddleware())
	// attach custom middleware
	router.Use(CorrelationIdMiddleware())

	router.POST("/user", func(ctx *gin.Context) {
		pgLogger.Info(ctx, "Got req for Postgres service")

		reqCtx := ctx.Request.Context()
		var req InsertUserRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Insert the username into the users table
		_, err := pgcConn.Exec(reqCtx, "INSERT INTO users (username) VALUES ($1)", req.Username)
		if err != nil {
			pgLogger.Error(ctx, "Failed to insert user: "+err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert user"})
			return
		}

		pgLogger.Info(ctx, "User added to postgres successfully")

		ctx.String(http.StatusOK, "User added successfully")
	})

	router.Run(":8082")
}
