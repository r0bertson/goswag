package main

import (
	"encoding/json"
	"net/http"

	"github.com/r0bertson/goswag"
	"github.com/r0bertson/goswag/models"
)

// User represents a user in the system
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Phone    *string `json:"phone"`               // Optional field (pointer)
	Age      *int    `json:"age,omitempty"`       // Optional field with omitempty
	IsActive *bool   `json:"is_active,omitempty"` // Optional boolean field
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func main() {
	// Create a new HTTP mux
	mux := http.NewServeMux()

	// Create goswag HTTP wrapper with default error responses
	defaultResponses := []models.ReturnType{
		{
			StatusCode: http.StatusBadRequest,
			Body:       ErrorResponse{},
		},
		{
			StatusCode: http.StatusInternalServerError,
			Body:       ErrorResponse{},
		},
	}

	// Initialize goswag HTTP wrapper
	swagger := goswag.NewHTTP(mux, defaultResponses...)

	// Define routes with swagger annotations
	swagger.GET("/users", handleGetUsers).
		Summary("Get all users").
		Description("Retrieve a list of all users in the system").
		Tags("users").
		Produces("application/json").
		Returns([]models.ReturnType{
			{
				StatusCode: http.StatusOK,
				Body:       []User{},
			},
		})

	swagger.POST("/users", handleCreateUser).
		Summary("Create a new user").
		Description("Create a new user with the provided information").
		Tags("users").
		Accepts("application/json").
		Produces("application/json").
		Read(CreateUserRequest{}).
		// Add field descriptions for request body properties
		ReadFieldDescriptions(map[string]string{
			"name":  "The full name of the user",
			"email": "The user's email address, must be a valid email format",
		}).
		Returns([]models.ReturnType{
			{
				StatusCode: http.StatusCreated,
				Body:       User{},
				// Add field descriptions for response body properties
				FieldDescriptions: map[string]string{
					"id":    "Unique identifier for the user",
					"name":  "The full name of the user",
					"email": "The user's email address",
				},
			},
		})

	// Create a group for API routes
	apiGroup := swagger.Group("/api/v1")

	apiGroup.GET("/users/{user_id}", handleGetUser).
		Summary("Get user by ID").
		Description("Retrieve a specific user by their ID").
		Tags("users").
		Produces("application/json").
		PathParam("user_id", "User ID", goswag.StringType, true).
		Returns([]models.ReturnType{
			{
				StatusCode: http.StatusOK,
				Body:       User{},
			},
			{
				StatusCode: http.StatusNotFound,
				Body:       ErrorResponse{},
			},
		})

	apiGroup.PUT("/users/{user_id}", handleUpdateUser).
		Summary("Update user").
		Description("Update an existing user's information").
		Tags("users").
		Accepts("application/json").
		Produces("application/json").
		Read(CreateUserRequest{}).
		PathParam("user_id", "User ID", goswag.StringType, true).
		HeaderParam("Authorization", "Bearer token", goswag.StringType, true).
		Returns([]models.ReturnType{
			{
				StatusCode: http.StatusOK,
				Body:       User{},
			},
		})

	apiGroup.DELETE("/users/{user_id}", handleDeleteUser).
		Summary("Delete user").
		Description("Delete a user from the system").
		Tags("users").
		PathParam("user_id", "User ID", goswag.StringType, true).
		HeaderParam("Authorization", "Bearer token", goswag.StringType, true).
		Returns([]models.ReturnType{
			{
				StatusCode: http.StatusNoContent,
			},
		})

	// Generate swagger documentation
	swagger.GenerateSwagger()

	// Start the server
	println("Server starting on :8080")
	println("Swagger documentation generated in goswag.go")
	http.ListenAndServe(":8080", mux)
}

// Handler functions
func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	users := []User{
		{ID: "1", Name: "John Doe", Email: "john@example.com"},
		{ID: "2", Name: "Jane Smith", Email: "jane@example.com"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user := User{
		ID:    "3", // In a real app, this would be generated
		Name:  req.Name,
		Email: req.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	// In a real app, you would extract user_id from the URL path
	user := User{ID: "1", Name: "John Doe", Email: "john@example.com"}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user := User{
		ID:    "1", // In a real app, this would come from the path
		Name:  req.Name,
		Email: req.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	// In a real app, you would delete the user from the database
	w.WriteHeader(http.StatusNoContent)
}
