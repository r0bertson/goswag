package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/r0bertson/goswag"
	"github.com/r0bertson/goswag/models"
)

// EnhancedErrorResponse represents an error response for enhanced tests
type EnhancedErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// EnhancedUser represents a user for enhanced tests
type EnhancedUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// SetupEnhancedRoutes creates routes with all v2.0 enhanced features for integration testing
func SetupEnhancedRoutes() *gin.Engine {
	g := gin.New()

	// Create global Swagger configuration
	config := models.NewSwaggerConfig().
		WithHost("api.example.com").
		WithBasePath("/v1").
		WithSchemes("https", "wss").
		WithContact(models.NewContactInfo("API Support", "support@example.com", "https://example.com/contact")).
		WithLicense(models.NewLicenseInfo("MIT", "https://opensource.org/licenses/MIT")).
		WithTermsOfService("https://example.com/terms").
		WithExternalDocs(models.NewExternalDocs("https://docs.example.com", "Complete API Documentation")).
		WithGlobalSecurity("BearerAuth")

	defaultResponses := []models.ReturnType{
		{
			StatusCode:  http.StatusBadRequest,
			Body:        EnhancedErrorResponse{},
			Description: "Bad request - invalid input",
		},
		{
			StatusCode:  http.StatusUnauthorized,
			Body:        EnhancedErrorResponse{},
			Description: "Unauthorized - authentication required",
		},
		{
			StatusCode:  http.StatusInternalServerError,
			Body:        EnhancedErrorResponse{},
			Description: "Internal server error",
		},
	}

	swagger := goswag.NewGinWithConfig(g, config, defaultResponses...)

	// ============================================================================
	// Test Enhanced Parameters (Sprint 1)
	// ============================================================================
	swagger.GET("/users").
		Summary("List users").
		Description("Retrieve a paginated list of users with optional filtering").
		Tags("users").
		OperationID("listUsers").
		// Enhanced query parameters with validation
		QueryParamWithOptions("page", "Page number", goswag.IntType, false,
			models.NewParamOptions().
				WithDefault(1).
				WithMinimum(1).
				WithExample(1),
		).
		QueryParamWithOptions("limit", "Items per page", goswag.IntType, false,
			models.NewParamOptions().
				WithDefault(10).
				WithMinimum(1).
				WithMaximum(100).
				WithExample(10),
		).
		QueryParamWithOptions("status", "Filter by status", goswag.StringType, false,
			models.NewParamOptions().
				WithEnum("active", "inactive", "pending").
				WithDefault("active"),
		).
		QueryParamWithOptions("tags", "Filter by tags", goswag.StringType, false,
			models.NewParamOptions().
				WithCollectionFormat("csv").
				WithExample("tag1,tag2,tag3"),
		).
		// Response with headers and examples
		Returns([]models.ReturnType{
			{
				StatusCode: http.StatusOK,
				Body:       []EnhancedUser{},
				Description: "Successfully retrieved users",
				Headers: map[string]*models.ResponseHeader{
					"X-Total-Count": models.NewResponseHeader("integer", "Total number of users"),
					"X-Page":         models.NewResponseHeader("integer", "Current page number"),
					"X-Request-ID":   models.NewResponseHeader("string", "Request identifier").WithFormat("uuid"),
				},
				Examples: map[string]interface{}{
					"application/json": []EnhancedUser{
						{ID: "1", Name: "John Doe", Email: "john@example.com"},
						{ID: "2", Name: "Jane Smith", Email: "jane@example.com"},
					},
				},
			},
		})

	// ============================================================================
	// Test Operation Enhancements (Sprint 2)
	// ============================================================================
	swagger.GET("/users/{id}").
		Summary("Get user by ID").
		Description("Retrieve a specific user by their unique identifier").
		Tags("users").
		OperationID("getUserById").
		Schemes("https").
		ExternalDocs("https://docs.example.com/users", "User API documentation").
		PathParamWithOptions("id", "User ID", goswag.StringType, true,
			models.NewParamOptions().
				WithFormat("uuid").
				WithPattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$").
				WithExample("550e8400-e29b-41d4-a716-446655440000"),
		).
		Returns([]models.ReturnType{
			{
				StatusCode: http.StatusOK,
				Body:       EnhancedUser{},
				Description: "User found successfully",
			},
			{
				StatusCode: http.StatusNotFound,
				Body:       EnhancedErrorResponse{},
				Description: "User not found",
			},
		})

	// ============================================================================
	// Test Deprecated Endpoint
	// ============================================================================
	swagger.GET("/users/legacy").
		Summary("Get users (legacy)").
		Description("Legacy endpoint - use /users instead").
		Tags("users").
		Deprecated().
		Returns([]models.ReturnType{
			{
				StatusCode: http.StatusOK,
				Body:       []EnhancedUser{},
			},
		})

	// ============================================================================
	// Test File Upload
	// ============================================================================
	swagger.POST("/users/{id}/avatar").
		Summary("Upload user avatar").
		Description("Upload an avatar image for a user").
		Tags("users").
		Accepts("multipart/form-data").
		PathParam("id", "User ID", goswag.StringType, true).
		FileParam("avatar", "Avatar image file", true).
		Returns([]models.ReturnType{
			{
				StatusCode: http.StatusOK,
				Body:       EnhancedUploadResponse{},
				Description: "Avatar uploaded successfully",
			},
		})

	// ============================================================================
	// Test Tag Metadata (Sprint 4)
	// ============================================================================
	adminGroup := swagger.Group("/admin").
		TagDescription("Admin operations require elevated privileges").
		TagExternalDocs("https://docs.example.com/admin", "Admin API documentation")
	
	adminGroup.GET("/users").
		Summary("List all users (admin)").
		Description("Retrieve all users including deleted ones (admin only)").
		Tags("admin", "users").
		Security("BearerAuth", "AdminAuth").
		Returns([]models.ReturnType{
			{
				StatusCode: http.StatusOK,
				Body:       []EnhancedUser{},
				Description: "All users retrieved successfully",
			},
		})

	swagger.GenerateSwagger()

	return g
}

// EnhancedUploadResponse represents a file upload response for enhanced tests
type EnhancedUploadResponse struct {
	URL      string `json:"url" example:"https://example.com/avatars/user123.jpg"`
	Filename string `json:"filename" example:"avatar.jpg"`
	Size     int64  `json:"size" example:"102400"`
}
