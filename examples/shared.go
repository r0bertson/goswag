package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	Phone    *string `json:"phone,omitempty"`     // Optional phone number
	Age      *int    `json:"age,omitempty"`       // Optional age
	IsActive *bool   `json:"is_active,omitempty"` // Optional active status
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Handler functions
func handleGetUsers(c *gin.Context) {
	users := []User{
		{ID: "1", Name: "John Doe", Email: "john@example.com"},
		{ID: "2", Name: "Jane Smith", Email: "jane@example.com"},
	}
	c.JSON(http.StatusOK, users)
}

func handleCreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	user := User{
		ID:    "3",
		Name:  req.Name,
		Email: req.Email,
	}
	c.JSON(http.StatusCreated, user)
}

func handleGetUser(c *gin.Context) {
	userID := c.Param("user_id")
	user := User{ID: userID, Name: "John Doe", Email: "john@example.com"}
	c.JSON(http.StatusOK, user)
}

func handleUpdateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	user := User{
		ID:    c.Param("user_id"),
		Name:  req.Name,
		Email: req.Email,
	}
	c.JSON(http.StatusOK, user)
}

func handleDeleteUser(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// SetupRoutes creates routes using goswag
// mode can be "release", "test", or "debug" (defaults to "release")
func SetupRoutes(mode ...string) *gin.Engine {
	ginMode := gin.ReleaseMode
	if len(mode) > 0 {
		switch mode[0] {
		case "test":
			ginMode = gin.TestMode
		case "debug":
			ginMode = gin.DebugMode
		default:
			ginMode = gin.ReleaseMode
		}
	}
	gin.SetMode(ginMode)
	g := gin.New()

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

	swagger := goswag.NewGin(g, defaultResponses...)

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
		ReadFieldDescriptions(map[string]string{
			"name":  "The full name of the user",
			"email": "The user's email address, must be a valid email format",
			"phone": "Optional phone number",
		}).
		Returns([]models.ReturnType{
			{
				StatusCode: http.StatusCreated,
				Body:       User{},
				FieldDescriptions: map[string]string{
					"id":    "Unique identifier for the user",
					"name":  "The full name of the user",
					"email": "The user's email address",
				},
			},
		})

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

	return g
}
