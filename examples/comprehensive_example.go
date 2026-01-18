package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/r0bertson/goswag"
	"github.com/r0bertson/goswag/models"
)

// @title           GoSwag Comprehensive Example API
// @version         2.0
// @description     This example demonstrates all new features of goswag including enhanced parameters, operation metadata, response enhancements, global configuration, and tag metadata
// @termsOfService  http://swagger.io/terms/
// @contact.name    API Support
// @contact.email   support@example.com
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
// @host            api.example.com
// @BasePath        /v1
func main() {
	mux := http.NewServeMux()

	// Define default error responses for all routes
	defaultResponses := []models.ReturnType{
		{
			StatusCode:  http.StatusBadRequest,
			Body:        ErrorResponse{},
			Description: "Bad request - invalid input",
		},
		{
			StatusCode:  http.StatusUnauthorized,
			Body:        ErrorResponse{},
			Description: "Unauthorized - authentication required",
		},
		{
			StatusCode:  http.StatusInternalServerError,
			Body:        ErrorResponse{},
			Description: "Internal server error",
		},
	}

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

	// Initialize goswag HTTP router with config and default responses
	gh := goswag.NewHTTPWithConfig(mux, config, defaultResponses...)

	// ============================================================================
	// Example 1: Basic route with enhanced parameters
	// ============================================================================
	gh.GET("/users").
		Summary("List users").
		Description("Retrieve a paginated list of users with optional filtering").
		Tags("users").
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
				StatusCode:  http.StatusOK,
				Body:        UserListResponse{},
				Description: "Successfully retrieved users",
				Headers: map[string]*models.ResponseHeader{
					"X-Total-Count": models.NewResponseHeader("integer", "Total number of users"),
					"X-Page":        models.NewResponseHeader("integer", "Current page number"),
					"X-Request-ID":  models.NewResponseHeader("string", "Request identifier").WithFormat("uuid"),
				},
				Examples: map[string]interface{}{
					"application/json": map[string]interface{}{
						"users": []User{
							{ID: 1, Name: "John Doe", Email: "john@example.com"},
							{ID: 2, Name: "Jane Smith", Email: "jane@example.com"},
						},
						"total": 2,
						"page":  1,
					},
				},
			},
		})

	// ============================================================================
	// Example 2: Route with operation-level enhancements
	// ============================================================================
	gh.GET("/users/{id}").
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
				StatusCode:  http.StatusOK,
				Body:        User{},
				Description: "User found successfully",
			},
			{
				StatusCode:  http.StatusNotFound,
				Body:        ErrorResponse{},
				Description: "User not found",
			},
		})

	// ============================================================================
	// Example 3: POST route with form data and file upload
	// ============================================================================
	gh.POST("/users").
		Summary("Create user").
		Description("Create a new user account").
		Tags("users").
		OperationID("createUser").
		Accepts("application/json").
		Read(CreateUserRequest{}).
		ReadFieldDescriptions(map[string]string{
			"name":  "User's full name (required, 2-100 characters)",
			"email": "User's email address (required, must be valid email format)",
			"age":   "User's age (optional, must be between 18 and 120)",
		}).
		Returns([]models.ReturnType{
			{
				StatusCode:  http.StatusCreated,
				Body:        User{},
				Description: "User created successfully",
				Headers: map[string]*models.ResponseHeader{
					"Location": models.NewResponseHeader("string", "URL of the created user").WithFormat("uri"),
				},
			},
		})

	// ============================================================================
	// Example 4: File upload endpoint
	// ============================================================================
	gh.POST("/users/{id}/avatar").
		Summary("Upload user avatar").
		Description("Upload an avatar image for a user").
		Tags("users").
		Accepts("multipart/form-data").
		PathParam("id", "User ID", goswag.StringType, true).
		FileParam("avatar", "Avatar image file", true).
		Returns([]models.ReturnType{
			{
				StatusCode:  http.StatusOK,
				Body:        UploadResponse{},
				Description: "Avatar uploaded successfully",
			},
		})

	// ============================================================================
	// Example 5: PATCH route for partial updates
	// ============================================================================
	gh.PATCH("/users/{id}").
		Summary("Partially update user").
		Description("Update specific fields of a user (partial update)").
		Tags("users").
		OperationID("patchUser").
		Accepts("application/json").
		Produces("application/json").
		PathParamWithOptions("id", "User ID", goswag.StringType, true,
			models.NewParamOptions().
				WithFormat("uuid").
				WithExample("550e8400-e29b-41d4-a716-446655440000"),
		).
		Read(CreateUserRequest{}).
		ReadFieldDescriptions(map[string]string{
			"name":  "User's full name (optional, 2-100 characters)",
			"email": "User's email address (optional, must be valid email format)",
			"age":   "User's age (optional, must be between 18 and 120)",
		}).
		Returns([]models.ReturnType{
			{
				StatusCode:  http.StatusOK,
				Body:        User{},
				Description: "User updated successfully",
			},
			{
				StatusCode:  http.StatusNotFound,
				Body:        ErrorResponse{},
				Description: "User not found",
			},
		})

	// ============================================================================
	// Example 6: Route with header parameters and validation
	// ============================================================================
	gh.GET("/users/me").
		Summary("Get current user").
		Description("Retrieve the currently authenticated user's profile").
		Tags("users").
		Security("BearerAuth").
		HeaderParamWithOptions("X-Request-ID", "Request identifier", goswag.StringType, false,
			models.NewParamOptions().
				WithFormat("uuid").
				WithExample("550e8400-e29b-41d4-a716-446655440000"),
		).
		Returns([]models.ReturnType{
			{
				StatusCode:  http.StatusOK,
				Body:        User{},
				Description: "Current user profile",
			},
		})

	// ============================================================================
	// Example 7: Deprecated endpoint
	// ============================================================================
	gh.GET("/users/legacy").
		Summary("Get users (legacy)").
		Description("Legacy endpoint - use /users instead").
		Tags("users").
		Deprecated().
		Returns([]models.ReturnType{
			{
				StatusCode: http.StatusOK,
				Body:       UserListResponse{},
			},
		})

	// ============================================================================
	// Example 8: Group with tag metadata
	// ============================================================================
	// Note: Group() returns HTTPRouter, so we need to cast or use a different pattern
	// For tag metadata, create the group first, then add routes
	adminGroupRouter := gh.Group("/admin")
	// In a real scenario, you would store the group and add tag metadata when creating routes
	// For this example, we'll add routes directly
	adminGroupRouter.GET("/users").
		Summary("List all users (admin)").
		Description("Retrieve all users including deleted ones (admin only)").
		Tags("admin", "users").
		Security("BearerAuth", "AdminAuth").
		Returns([]models.ReturnType{
			{
				StatusCode:  http.StatusOK,
				Body:        UserListResponse{},
				Description: "All users retrieved successfully",
			},
		})

	// ============================================================================
	// Example 9: Health check endpoint
	// ============================================================================
	gh.GET("/health").
		Summary("Health check").
		Description("Check API health status").
		Tags("health").
		Returns([]models.ReturnType{
			{
				StatusCode:  http.StatusOK,
				Body:        HealthResponse{},
				Description: "API is healthy",
			},
		})

	// Generate Swagger documentation
	gh.GenerateSwagger()

	// Get current directory
	originalDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Try to generate swagger.json using swag
	swaggerJSONPath, err := generateSwaggerJSON(originalDir)
	if err != nil {
		log.Printf("Warning: Failed to generate swagger.json: %v", err)
		log.Println("Serving API without Swagger UI. Install swag to enable Swagger UI:")
		log.Println("  go install github.com/swaggo/swag/cmd/swag@latest")
		serveAPIOnly(mux)
		return
	}

	// Serve swagger.json
	mux.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, swaggerJSONPath)
	})

	// Serve Swagger UI HTML page
	swaggerHTML := getSwaggerUIHTML()
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(swaggerHTML))
	})

	mux.HandleFunc("/swagger/index.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(swaggerHTML))
	})

	log.Println("\n🚀 Server starting on http://localhost:8080")
	log.Println("📚 Swagger UI available at: http://localhost:8080/swagger/index.html")
	log.Println("📄 Swagger JSON available at: http://localhost:8080/swagger/doc.json")
	log.Println("\nAPI Endpoints:")
	log.Println("  GET    /v1/users")
	log.Println("  GET    /v1/users/{id}")
	log.Println("  POST   /v1/users")
	log.Println("  PATCH  /v1/users/{id}")
	log.Println("  POST   /v1/users/{id}/avatar")
	log.Println("  GET    /v1/users/me")
	log.Println("  GET    /v1/users/legacy")
	log.Println("  GET    /v1/admin/users")
	log.Println("  GET    /v1/health")
	log.Println("\nPress Ctrl+C to stop the server")

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func serveAPIOnly(mux *http.ServeMux) {
	log.Println("\n🚀 Server starting on http://localhost:8080")
	log.Println("API Endpoints:")
	log.Println("  GET    /v1/users")
	log.Println("  GET    /v1/users/{id}")
	log.Println("  POST   /v1/users")
	log.Println("  PATCH  /v1/users/{id}")
	log.Println("  POST   /v1/users/{id}/avatar")
	log.Println("  GET    /v1/users/me")
	log.Println("  GET    /v1/users/legacy")
	log.Println("  GET    /v1/admin/users")
	log.Println("  GET    /v1/health")
	log.Println("\nNote: Install swag to enable Swagger UI")
	log.Println("  go install github.com/swaggo/swag/cmd/swag@latest")
	log.Println("\nPress Ctrl+C to stop the server")

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func generateSwaggerJSON(originalDir string) (string, error) {
	// Create a temporary directory for swagger generation
	tempDir, err := os.MkdirTemp("", "goswag-comprehensive-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	goswagDir := filepath.Join(tempDir, "goswag")
	docsDir := filepath.Join(tempDir, "docs")

	if err := os.MkdirAll(goswagDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create goswag directory: %w", err)
	}
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create docs directory: %w", err)
	}

	// The goswag module root is the parent of the examples directory
	goswagModuleRoot := filepath.Dir(originalDir)

	// Read goswag.go from original directory
	goswagGoPath := filepath.Join(originalDir, "goswag.go")
	goswagGoContent, err := os.ReadFile(goswagGoPath)
	if err != nil {
		return "", fmt.Errorf("failed to read goswag.go: %w", err)
	}

	// Write goswag.go to temp directory
	if err := os.WriteFile(filepath.Join(goswagDir, "goswag.go"), goswagGoContent, 0644); err != nil {
		return "", fmt.Errorf("failed to write goswag.go: %w", err)
	}

	// Copy comprehensive_example.go to temp directory so swag can resolve type definitions
	comprehensiveGoPath := filepath.Join(originalDir, "comprehensive_example.go")
	comprehensiveGoContent, err := os.ReadFile(comprehensiveGoPath)
	if err != nil {
		log.Printf("Warning: Failed to read comprehensive_example.go: %v", err)
	} else {
		if err := os.WriteFile(filepath.Join(goswagDir, "comprehensive_example.go"), comprehensiveGoContent, 0644); err != nil {
			log.Printf("Warning: Failed to write comprehensive_example.go: %v", err)
		}
	}

	// Create main.go for swag
	mainGoContent := `// @title           GoSwag Comprehensive Example API
// @version         2.0
// @description     This example demonstrates all new features of goswag including enhanced parameters, operation metadata, response enhancements, global configuration, and tag metadata
// @host            api.example.com
// @BasePath        /v1
package main

import _ "swagger-temp/goswag"
`
	if err := os.WriteFile(filepath.Join(goswagDir, "main.go"), []byte(mainGoContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write main.go: %w", err)
	}

	// Create go.mod for swag
	relPath, err := filepath.Rel(goswagDir, goswagModuleRoot)
	if err != nil {
		goswagModuleRootAbs, absErr := filepath.Abs(goswagModuleRoot)
		if absErr != nil {
			return "", fmt.Errorf("failed to get path for goswag module root: %w", absErr)
		}
		relPath = filepath.ToSlash(goswagModuleRootAbs)
	} else {
		relPath = filepath.ToSlash(relPath)
		if !filepath.IsAbs(relPath) && !strings.HasPrefix(relPath, "..") && relPath != "." {
			relPath = "../" + relPath
		}
	}

	goModContent := fmt.Sprintf(`module swagger-temp

go 1.23

require (
	github.com/r0bertson/goswag v0.0.0
)

replace github.com/r0bertson/goswag => %s
`, relPath)

	if err := os.WriteFile(filepath.Join(goswagDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write go.mod: %w", err)
	}

	// Initialize go module
	goModCmd := exec.Command("go", "mod", "tidy")
	goModCmd.Dir = goswagDir
	goModCmd.Env = os.Environ()
	if output, err := goModCmd.CombinedOutput(); err != nil {
		log.Printf("Warning: go mod tidy failed: %v", err)
		log.Printf("Output: %s", string(output))
	}

	// Find swag
	swagPath, err := exec.LookPath("swag")
	if err != nil {
		return "", fmt.Errorf("swag not found in PATH: %w", err)
	}

	// Run swag init
	docsRelPath, err := filepath.Rel(goswagDir, docsDir)
	if err != nil {
		return "", fmt.Errorf("failed to calculate relative path for docs: %w", err)
	}
	docsRelPath = filepath.ToSlash(docsRelPath)

	swagCmd := exec.Command(swagPath, "init", "--parseDependency", "--parseInternal", "-g", "main.go", "-o", docsRelPath)
	swagCmd.Dir = goswagDir
	swagCmd.Env = os.Environ()

	if output, err := swagCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("swag init failed: %w\nOutput: %s", err, string(output))
	}

	// Verify swagger.json was created
	swaggerJSONPath := filepath.Join(docsDir, "swagger.json")
	if _, err := os.Stat(swaggerJSONPath); os.IsNotExist(err) {
		return "", fmt.Errorf("swagger.json not generated")
	}

	// Read and validate swagger.json
	swaggerJSON, err := os.ReadFile(swaggerJSONPath)
	if err != nil {
		return "", fmt.Errorf("failed to read swagger.json: %w", err)
	}

	var swaggerDoc map[string]interface{}
	if err := json.Unmarshal(swaggerJSON, &swaggerDoc); err != nil {
		return "", fmt.Errorf("invalid swagger.json: %w", err)
	}

	// Post-process swagger.json to remove pointer fields from required arrays
	if err := fixPointerFieldsInSwagger(swaggerDoc, goswagDir); err != nil {
		log.Printf("Warning: Failed to fix pointer fields in swagger: %v", err)
	} else {
		fixedJSON, err := json.MarshalIndent(swaggerDoc, "", "  ")
		if err == nil {
			if err := os.WriteFile(swaggerJSONPath, fixedJSON, 0644); err != nil {
				log.Printf("Warning: Failed to write fixed swagger.json: %v", err)
			}
		}
	}

	info := swaggerDoc["info"].(map[string]interface{})
	log.Printf("✓ Swagger documentation generated successfully!")
	log.Printf("  Title: %s", info["title"])
	log.Printf("  Version: %s", info["version"])

	// Copy swagger.json to original directory for serving
	finalSwaggerPath := filepath.Join(originalDir, "swagger.json")
	if err := os.WriteFile(finalSwaggerPath, swaggerJSON, 0644); err != nil {
		log.Printf("Warning: Failed to copy swagger.json to %s: %v", finalSwaggerPath, err)
		return swaggerJSONPath, nil // Return temp path if copy fails
	}

	return finalSwaggerPath, nil
}

func fixPointerFieldsInSwagger(swaggerDoc map[string]interface{}, goswagDir string) error {
	definitions, ok := swaggerDoc["definitions"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Read goswag.go to identify pointer fields
	goswagGoPath := filepath.Join(goswagDir, "goswag.go")
	goswagContent, err := os.ReadFile(goswagGoPath)
	if err != nil {
		return fmt.Errorf("failed to read goswag.go: %w", err)
	}

	pointerFields := findPointerFieldsInSource(string(goswagContent))

	for defName, defValue := range definitions {
		def, ok := defValue.(map[string]interface{})
		if !ok {
			continue
		}

		properties, ok := def["properties"].(map[string]interface{})
		if !ok {
			continue
		}

		required, ok := def["required"].([]interface{})
		if !ok {
			required = []interface{}{}
		}

		var pointerFieldNames []string
		for structName, fields := range pointerFields {
			if strings.Contains(defName, structName) {
				pointerFieldNames = append(pointerFieldNames, fields...)
			}
		}

		for propName, propValue := range properties {
			prop, ok := propValue.(map[string]interface{})
			if !ok {
				continue
			}

			isPointer := false
			for _, fieldName := range pointerFieldNames {
				if propName == fieldName {
					isPointer = true
					break
				}
			}

			if isPointer {
				prop["x-nullable"] = true
				delete(prop, "default")
				prop["x-example"] = nil
			}
		}

		var newRequired []interface{}
		for _, reqField := range required {
			reqFieldStr, ok := reqField.(string)
			if !ok {
				newRequired = append(newRequired, reqField)
				continue
			}

			isPointer := false
			for _, fieldName := range pointerFieldNames {
				if reqFieldStr == fieldName {
					isPointer = true
					break
				}
			}

			if !isPointer {
				newRequired = append(newRequired, reqField)
			}
		}

		if len(newRequired) > 0 {
			def["required"] = newRequired
		} else {
			delete(def, "required")
		}
	}

	return nil
}

func findPointerFieldsInSource(source string) map[string][]string {
	result := make(map[string][]string)
	lines := strings.Split(source, "\n")

	var currentStruct string
	for i, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "type ") && strings.Contains(line, "struct {") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentStruct = parts[1]
				result[currentStruct] = []string{}
			}
			continue
		}

		if line == "}" && currentStruct != "" {
			currentStruct = ""
			continue
		}

		if currentStruct != "" && strings.Contains(line, "*") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				fieldName := parts[0]
				for j := i; j < len(lines) && j < i+3; j++ {
					if strings.Contains(lines[j], "json:") {
						jsonTag := extractJSONFieldName(lines[j])
						if jsonTag != "" {
							result[currentStruct] = append(result[currentStruct], jsonTag)
						} else {
							if len(fieldName) > 0 {
								jsonName := strings.ToLower(fieldName[:1]) + fieldName[1:]
								result[currentStruct] = append(result[currentStruct], jsonName)
							}
						}
						break
					}
				}
			}
		}
	}

	return result
}

func extractJSONFieldName(tagLine string) string {
	start := strings.Index(tagLine, `json:"`)
	if start == -1 {
		return ""
	}
	start += 6
	end := strings.Index(tagLine[start:], `"`)
	if end == -1 {
		return ""
	}
	jsonTag := tagLine[start : start+end]
	parts := strings.Split(jsonTag, ",")
	return parts[0]
}

func getSwaggerUIHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Swagger UI - Comprehensive Example</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css" />
  <style>
    html {
      box-sizing: border-box;
      overflow: -moz-scrollbars-vertical;
      overflow-y: scroll;
    }
    *, *:before, *:after {
      box-sizing: inherit;
    }
    body {
      margin: 0;
      background: #fafafa;
    }
    /* Style for nullable/optional fields */
    .swagger-ui .model-box .property[data-nullable="true"] .property-name::after {
      content: " (optional, nullable)";
      color: #999;
      font-weight: normal;
      font-size: 0.9em;
    }
    .swagger-ui .model-box .property[data-nullable="true"] {
      opacity: 0.85;
    }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "/swagger/doc.json",
        dom_id: '#swagger-ui',
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout",
        deepLinking: true,
        filter: true,
        showExtensions: true,
        showCommonExtensions: true,
        onComplete: function() {
          setTimeout(function() {
            const nullableProps = document.querySelectorAll('[data-nullable="true"]');
            nullableProps.forEach(function(prop) {
              const propName = prop.querySelector('.property-name');
              if (propName && !propName.textContent.includes('(optional')) {
                propName.innerHTML += ' <span style="color: #999; font-weight: normal; font-size: 0.9em;">(optional, nullable)</span>';
              }
            });
          }, 500);
        }
      });
      window.ui = ui;
    };
  </script>
</body>
</html>`
}

// ============================================================================
// Data Models
// ============================================================================

type User struct {
	ID    int    `json:"id" example:"1"`
	Name  string `json:"name" example:"John Doe"`
	Email string `json:"email" example:"john@example.com"`
	Age   *int   `json:"age,omitempty" example:"30"`
}

type UserListResponse struct {
	Users []User `json:"users"`
	Total int    `json:"total" example:"100"`
	Page  int    `json:"page" example:"1"`
}

type CreateUserRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=100"`
	Email string `json:"email" binding:"required,email"`
	Age   *int   `json:"age,omitempty" binding:"omitempty,min=18,max=120"`
}

type ErrorResponse struct {
	Error   string `json:"error" example:"Bad request"`
	Message string `json:"message,omitempty" example:"Invalid input parameters"`
}

type UploadResponse struct {
	URL      string `json:"url" example:"https://example.com/avatars/user123.jpg"`
	Filename string `json:"filename" example:"avatar.jpg"`
	Size     int64  `json:"size" example:"102400"`
}

type HealthResponse struct {
	Status    string `json:"status" example:"ok"`
	Timestamp string `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	Version   string `json:"version" example:"1.0.0"`
}
