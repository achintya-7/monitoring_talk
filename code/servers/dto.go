package servers

const (
	// Alloy
	ALLOY_ENDPOINT = "localhost:9001"
	ALLOY_TOKEN    = "G7kLw9xYtQpZrV3mN2sHjXcB8aWfUe5K"

	// Services
	AUTH_SERVICE_NAME     = "AUTH_SERVICE"
	MONGO_SERVICE_NAME    = "MONGO_USER_SERVICE"
	POSTGRES_SERVICE_NAME = "POSTGRES_USER_SERVICE"

	// Database
	MONGO_URL    = "mongodb://localhost:27017"
	POSTGRES_URL = "postgres://postgres:postgres@localhost:5432/demo"

	// MongoDB Collections
	USER_DATABASE    = "users"
	USERS_COLLECTION = "users"

	CORRELATION_ID = "correlation_id"
)

type InsertUserRequest struct {
	Username string `json:"username" binding:"required"`
}
