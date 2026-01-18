package main

import (
	_ "main"
)

// Wrappermain_CreateUserRequestRequest is a wrapper struct with field descriptions
type Wrappermain_CreateUserRequestRequest struct {
	// User's full name (required, 2-100 characters)
	Name string `json:"name" binding:"required,min=2,max=100"`
	// User's email address (required, must be valid email format)
	Email string `json:"email" binding:"required,email"`
	// User's age (optional, must be between 18 and 120)
	Age *int `json:"age,omitempty" binding:"omitempty,min=18,max=120"`
}

// Wrappermain_CreateUserRequestRequest is a wrapper struct with field descriptions
type Wrappermain_CreateUserRequestRequest struct {
	// User's full name (optional, 2-100 characters)
	Name string `json:"name" binding:"required,min=2,max=100"`
	// User's email address (optional, must be valid email format)
	Email string `json:"email" binding:"required,email"`
	// User's age (optional, must be between 18 and 120)
	Age *int `json:"age,omitempty" binding:"omitempty,min=18,max=120"`
}

// @host api.example.com
// @BasePath /v1
// @schemes https wss
// @contact.name API Support
// @contact.email support@example.com
// @contact.url https://example.com/contact
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @termsOfService https://example.com/terms
// @externalDocs.description Complete API Documentation
// @externalDocs.url https://docs.example.com
// @security BearerAuth

// @Summary List users
// @Description Retrieve a paginated list of users with optional filtering
// @Tags users
// @Produce json
// @Param page query int false "Page number (default: 1) (example: 1) (min: 1)"
// @Param limit query int false "Items per page (default: 10) (example: 10) (min: 1) (max: 100)"
// @Param status query string false "Filter by status (enum: active,inactive,pending) (default: active)"
// @Param tags query string false "Filter by tags (example: tag1,tag2,tag3)"
// @Security BearerAuth
// @Success 200 {object} main.UserListResponse "Successfully retrieved users"// @Header X-Total-Count {integer} "Total number of users"
// @Header X-Page {integer} "Current page number"
// @Header X-Request-ID {string} "Request identifier (format: uuid)"
// @Example application/json "map[page:1 total:2 users:[{1 John Doe john@example.com <nil>} {2 Jane Smith jane@example.com <nil>}]]"

// @Failure 400 {object} main.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} main.ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} main.ErrorResponse "Internal server error"
// @Router /users [get]
func handler() {} //nolint:unused 

// @Summary Get user by ID
// @Description Retrieve a specific user by their unique identifier
// @Tags users
// @Produce json
// @Param id path string true "User ID (format: uuid) (example: 550e8400-e29b-41d4-a716-446655440000) (pattern: ^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$)"
// @Security BearerAuth
// @OperationID getUserById
// @Schemes https
// @ExternalDocs https://docs.example.com/users "User API documentation"
// @Success 200 {object} main.User "User found successfully"
// @Failure 404 {object} main.ErrorResponse "User not found"
// @Failure 400 {object} main.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} main.ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} main.ErrorResponse "Internal server error"
// @Router /users/{id} [get]
func handler() {} //nolint:unused 

// @Summary Create user
// @Description Create a new user account
// @Tags users
// @Accept application/json
// @Produce json
// @Param request body Wrappermain_CreateUserRequestRequest true "Request"
// @Security BearerAuth
// @OperationID createUser
// @Success 201 {object} main.User "User created successfully"// @Header Location {string} "URL of the created user (format: uri)"

// @Failure 400 {object} main.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} main.ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} main.ErrorResponse "Internal server error"
// @Router /users [post]
func handler() {} //nolint:unused 

// @Summary Upload user avatar
// @Description Upload an avatar image for a user
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "User ID"
// @Param avatar formData file true "Avatar image file"
// @Security BearerAuth
// @Success 200 {object} main.UploadResponse "Avatar uploaded successfully"
// @Failure 400 {object} main.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} main.ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} main.ErrorResponse "Internal server error"
// @Router /users/{id}/avatar [post]
func handler() {} //nolint:unused 

// @Summary Partially update user
// @Description Update specific fields of a user (partial update)
// @Tags users
// @Produce application/json
// @Param request body Wrappermain_CreateUserRequestRequest true "Request"
// @Param id path string true "User ID (format: uuid) (example: 550e8400-e29b-41d4-a716-446655440000)"
// @Security BearerAuth
// @OperationID patchUser
// @Success 200 {object} main.User "User updated successfully"
// @Failure 404 {object} main.ErrorResponse "User not found"
// @Failure 400 {object} main.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} main.ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} main.ErrorResponse "Internal server error"
// @Router /users/{id} [patch]
func handler() {} //nolint:unused 

// @Summary Get current user
// @Description Retrieve the currently authenticated user's profile
// @Tags users
// @Produce json
// @Param X-Request-ID header string false "Request identifier (format: uuid) (example: 550e8400-e29b-41d4-a716-446655440000)"
// @Security BearerAuth
// @Success 200 {object} main.User "Current user profile"
// @Failure 400 {object} main.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} main.ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} main.ErrorResponse "Internal server error"
// @Router /users/me [get]
func handler() {} //nolint:unused 

// @Summary Get users (legacy)
// @Description Legacy endpoint - use /users instead
// @Tags users
// @Produce json
// @Security BearerAuth
// @Deprecated
// @Success 200 {object} main.UserListResponse
// @Failure 400 {object} main.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} main.ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} main.ErrorResponse "Internal server error"
// @Router /users/legacy [get]
func handler() {} //nolint:unused 

// @Summary Health check
// @Description Check API health status
// @Tags health
// @Produce json
// @Security BearerAuth
// @Success 200 {object} main.HealthResponse "API is healthy"
// @Failure 400 {object} main.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} main.ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} main.ErrorResponse "Internal server error"
// @Router /health [get]
func handler() {} //nolint:unused 

// @Summary List all users (admin)
// @Description Retrieve all users including deleted ones (admin only)
// @Tags admin,users
// @Produce json
// @Security BearerAuth
// @Security AdminAuth
// @Success 200 {object} main.UserListResponse "All users retrieved successfully"
// @Failure 400 {object} main.ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} main.ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} main.ErrorResponse "Internal server error"
// @Router /admin/users [get]
func handler() {} //nolint:unused 

