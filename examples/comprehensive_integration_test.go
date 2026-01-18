package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/r0bertson/goswag"
	"github.com/r0bertson/goswag/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAllHTTPMethods tests that all HTTP methods are properly documented
func TestAllHTTPMethods(t *testing.T) {
	g := gin.New()
	swagger := goswag.NewGin(g)

	// Test all HTTP methods
	swagger.GET("/test/get").
		Summary("GET test").
		Tags("test").
		Returns([]models.ReturnType{{StatusCode: http.StatusOK}})

	swagger.POST("/test/post").
		Summary("POST test").
		Tags("test").
		Returns([]models.ReturnType{{StatusCode: http.StatusCreated}})

	swagger.PUT("/test/put").
		Summary("PUT test").
		Tags("test").
		Returns([]models.ReturnType{{StatusCode: http.StatusOK}})

	swagger.DELETE("/test/delete").
		Summary("DELETE test").
		Tags("test").
		Returns([]models.ReturnType{{StatusCode: http.StatusNoContent}})

	swagger.PATCH("/test/patch").
		Summary("PATCH test").
		Tags("test").
		Returns([]models.ReturnType{{StatusCode: http.StatusOK}})

	swagger.OPTIONS("/test/options").
		Summary("OPTIONS test").
		Tags("test").
		Returns([]models.ReturnType{{StatusCode: http.StatusOK}})

	swagger.HEAD("/test/head").
		Summary("HEAD test").
		Tags("test").
		Returns([]models.ReturnType{{StatusCode: http.StatusOK}})

	swagger.GenerateSwagger()

	goswagContent, err := os.ReadFile("goswag.go")
	require.NoError(t, err)
	content := string(goswagContent)

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	for _, method := range methods {
		assert.Contains(t, content, method+" /test/", "Should contain %s route", method)
	}
}

// TestAllParameterTypes tests all parameter types and locations
func TestAllParameterTypes(t *testing.T) {
	g := gin.New()
	swagger := goswag.NewGin(g)

	swagger.GET("/test/params/{path_id}").
		Summary("Test all parameter types").
		Tags("test").
		PathParam("path_id", "Path parameter", goswag.StringType, true).
		QueryParam("query_str", "Query string", goswag.StringType, false).
		QueryParam("query_int", "Query int", goswag.IntType, false).
		QueryParam("query_num", "Query number", goswag.NumberType, false).
		QueryParam("query_bool", "Query bool", goswag.BoolType, false).
		HeaderParam("X-Custom-Header", "Custom header", goswag.StringType, false).
		FormDataParam("form_field", "Form data field", goswag.StringType, false, nil).
		Returns([]models.ReturnType{{StatusCode: http.StatusOK}})

	swagger.GenerateSwagger()

	goswagContent, err := os.ReadFile("goswag.go")
	require.NoError(t, err)
	content := string(goswagContent)

	// Check for parameter types
	assert.Contains(t, content, "path_id", "Should contain path parameter")
	assert.Contains(t, content, "query_str", "Should contain query parameter")
	assert.Contains(t, content, "X-Custom-Header", "Should contain header parameter")
	assert.Contains(t, content, "form_field", "Should contain form data parameter")
}

// TestNestedGroups tests nested group structure
func TestNestedGroups(t *testing.T) {
	g := gin.New()
	swagger := goswag.NewGin(g)

	apiV1 := swagger.Group("/api/v1")
	usersGroup := apiV1.Group("/users")
	adminGroup := usersGroup.Group("/admin")

	adminGroup.GET("/list").
		Summary("Admin list users").
		Tags("admin", "users").
		Returns([]models.ReturnType{{StatusCode: http.StatusOK}})

	swagger.GenerateSwagger()

	goswagContent, err := os.ReadFile("goswag.go")
	require.NoError(t, err)
	content := string(goswagContent)

	// Check nested path
	assert.Contains(t, content, "/api/v1/users/admin/list", "Should contain nested group path")
}

// TestComplexPathParameters tests complex path parameter scenarios
func TestComplexPathParameters(t *testing.T) {
	g := gin.New()
	swagger := goswag.NewGin(g)

	swagger.GET("/users/{user_id}/posts/{post_id}/comments/{comment_id}").
		Summary("Get comment").
		Tags("comments").
		PathParam("user_id", "User ID", goswag.StringType, true).
		PathParam("post_id", "Post ID", goswag.StringType, true).
		PathParam("comment_id", "Comment ID", goswag.StringType, true).
		Returns([]models.ReturnType{{StatusCode: http.StatusOK}})

	swagger.GenerateSwagger()

	goswagContent, err := os.ReadFile("goswag.go")
	require.NoError(t, err)
	content := string(goswagContent)

	// Check all path parameters are present
	assert.Contains(t, content, "user_id", "Should contain user_id parameter")
	assert.Contains(t, content, "post_id", "Should contain post_id parameter")
	assert.Contains(t, content, "comment_id", "Should contain comment_id parameter")
}

// TestMultipleResponses tests routes with multiple response types
func TestMultipleResponses(t *testing.T) {
	g := gin.New()
	swagger := goswag.NewGin(g)

	swagger.GET("/test/responses").
		Summary("Test multiple responses").
		Tags("test").
		Returns([]models.ReturnType{
			{StatusCode: http.StatusOK, Description: "Success"},
			{StatusCode: http.StatusBadRequest, Description: "Bad request"},
			{StatusCode: http.StatusUnauthorized, Description: "Unauthorized"},
			{StatusCode: http.StatusNotFound, Description: "Not found"},
			{StatusCode: http.StatusInternalServerError, Description: "Server error"},
		})

	swagger.GenerateSwagger()

	goswagContent, err := os.ReadFile("goswag.go")
	require.NoError(t, err)
	content := string(goswagContent)

	// Check multiple response codes
	statusCodes := []string{"200", "400", "401", "404", "500"}
	for _, code := range statusCodes {
		assert.Contains(t, content, code, "Should contain status code %s", code)
	}
}

// TestDefaultResponses tests that default responses are applied
func TestDefaultResponses(t *testing.T) {
	g := gin.New()
	defaultResponses := []models.ReturnType{
		{StatusCode: http.StatusBadRequest, Description: "Default bad request"},
		{StatusCode: http.StatusInternalServerError, Description: "Default server error"},
	}
	swagger := goswag.NewGin(g, defaultResponses...)

	swagger.GET("/test/defaults").
		Summary("Test default responses").
		Tags("test").
		Returns([]models.ReturnType{
			{StatusCode: http.StatusOK, Description: "Success"},
		})

	swagger.GenerateSwagger()

	goswagContent, err := os.ReadFile("goswag.go")
	require.NoError(t, err)
	content := string(goswagContent)

	// Default responses should be added to all routes
	assert.Contains(t, content, "400", "Should contain default 400 response")
	assert.Contains(t, content, "500", "Should contain default 500 response")
}

// TestSecuritySchemes tests security scheme application
func TestSecuritySchemes(t *testing.T) {
	g := gin.New()
	swagger := goswag.NewGin(g)

	swagger.GET("/test/secure").
		Summary("Secure endpoint").
		Tags("test").
		Security("BearerAuth", "ApiKeyAuth").
		Returns([]models.ReturnType{{StatusCode: http.StatusOK}})

	swagger.GenerateSwagger()

	goswagContent, err := os.ReadFile("goswag.go")
	require.NoError(t, err)
	content := string(goswagContent)

	// Check security annotations
	assert.Contains(t, content, "@Security", "Should contain @Security annotation")
	assert.Contains(t, content, "BearerAuth", "Should contain BearerAuth")
	assert.Contains(t, content, "ApiKeyAuth", "Should contain ApiKeyAuth")
}

// TestContentTypes tests Accepts and Produces content types
func TestContentTypes(t *testing.T) {
	g := gin.New()
	swagger := goswag.NewGin(g)

	swagger.POST("/test/content").
		Summary("Test content types").
		Tags("test").
		Accepts("application/json", "application/xml").
		Produces("application/json", "application/xml", "text/plain").
		Returns([]models.ReturnType{{StatusCode: http.StatusOK}})

	swagger.GenerateSwagger()

	goswagContent, err := os.ReadFile("goswag.go")
	require.NoError(t, err)
	content := string(goswagContent)

	// Check content type annotations
	assert.Contains(t, content, "@Accept", "Should contain @Accept annotation")
	assert.Contains(t, content, "@Produce", "Should contain @Produce annotation")
}

// TestEmptyRoutes tests behavior with no routes
func TestEmptyRoutes(t *testing.T) {
	g := gin.New()
	swagger := goswag.NewGin(g)

	// No routes added
	swagger.GenerateSwagger()

	goswagContent, err := os.ReadFile("goswag.go")
	require.NoError(t, err)
	content := string(goswagContent)

	// Should still generate valid swagger file
	assert.NotEmpty(t, content, "Should generate swagger file even with no routes")
}

// TestTagGeneration tests automatic tag generation from groups
func TestTagGeneration(t *testing.T) {
	g := gin.New()
	swagger := goswag.NewGin(g)

	usersGroup := swagger.Group("/users")
	usersGroup.GET("/list").
		Summary("List users").
		Returns([]models.ReturnType{{StatusCode: http.StatusOK}})

	postsGroup := swagger.Group("/posts")
	postsGroup.GET("/list").
		Summary("List posts").
		Returns([]models.ReturnType{{StatusCode: http.StatusOK}})

	swagger.GenerateSwagger()

	goswagContent, err := os.ReadFile("goswag.go")
	require.NoError(t, err)
	content := string(goswagContent)

	// Groups should generate tags
	assert.Contains(t, content, "@Tags", "Should contain @Tags annotation")
}

// TestFieldDescriptions tests field descriptions in request and response bodies
func TestFieldDescriptions(t *testing.T) {
	type TestRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type TestResponse struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	g := gin.New()
	swagger := goswag.NewGin(g)

	swagger.POST("/test/fields").
		Summary("Test field descriptions").
		Tags("test").
		Read(TestRequest{}).
		ReadFieldDescriptions(map[string]string{
			"name":  "User's full name",
			"email": "User's email address",
		}).
		Returns([]models.ReturnType{
			{
				StatusCode: http.StatusOK,
				Body:       TestResponse{},
				FieldDescriptions: map[string]string{
					"id":    "Unique identifier",
					"name":  "User's full name",
					"email": "User's email address",
				},
			},
		})

	swagger.GenerateSwagger()

	goswagContent, err := os.ReadFile("goswag.go")
	require.NoError(t, err)
	content := string(goswagContent)

	// Check for field descriptions in wrapper structs
	assert.Contains(t, content, "User's full name", "Should contain request field description")
	assert.Contains(t, content, "Unique identifier", "Should contain response field description")
}

// TestSwaggerJSONComprehensive validates comprehensive Swagger JSON structure
func TestSwaggerJSONComprehensive(t *testing.T) {
	// Generate enhanced routes
	_ = SetupEnhancedRoutes()

	swaggerJSONPath := filepath.Join("docs", "swagger.json")
	if _, err := os.Stat(swaggerJSONPath); os.IsNotExist(err) {
		t.Skip("swagger.json not found. Run 'swag init' first.")
		return
	}

	swaggerJSON, err := os.ReadFile(swaggerJSONPath)
	require.NoError(t, err)

	var swaggerDoc map[string]interface{}
	require.NoError(t, json.Unmarshal(swaggerJSON, &swaggerDoc))

	t.Run("Swagger Version", func(t *testing.T) {
		assert.Equal(t, "2.0", swaggerDoc["swagger"], "Should be Swagger 2.0")
	})

	t.Run("Info Section", func(t *testing.T) {
		info, ok := swaggerDoc["info"].(map[string]interface{})
		require.True(t, ok, "Should have info section")
		assert.Contains(t, info, "title", "Should have title")
		assert.Contains(t, info, "version", "Should have version")
	})

	t.Run("Paths Section", func(t *testing.T) {
		paths, ok := swaggerDoc["paths"].(map[string]interface{})
		require.True(t, ok, "Should have paths section")
		assert.NotEmpty(t, paths, "Should have at least one path")
	})

	t.Run("Path Operations", func(t *testing.T) {
		paths := swaggerDoc["paths"].(map[string]interface{})
		for pathKey, pathValue := range paths {
			path, ok := pathValue.(map[string]interface{})
			require.True(t, ok, "Path %s should be an object", pathKey)

			hasOperation := false
			for method := range path {
				if method == "get" || method == "post" || method == "put" ||
					method == "delete" || method == "patch" || method == "options" || method == "head" {
					hasOperation = true
					break
				}
			}
			assert.True(t, hasOperation, "Path %s should have at least one operation", pathKey)
		}
	})

	t.Run("Operation Structure", func(t *testing.T) {
		paths := swaggerDoc["paths"].(map[string]interface{})
		for pathKey, pathValue := range paths {
			path := pathValue.(map[string]interface{})
			for method, operation := range path {
				if method == "get" || method == "post" || method == "put" || method == "delete" {
					op, ok := operation.(map[string]interface{})
					require.True(t, ok, "Operation %s %s should be an object", method, pathKey)

					// Check required fields
					assert.Contains(t, op, "responses", "Operation should have responses")
					assert.Contains(t, op, "tags", "Operation should have tags")

					// Check responses structure
					responses, ok := op["responses"].(map[string]interface{})
					require.True(t, ok, "Responses should be an object")
					assert.NotEmpty(t, responses, "Should have at least one response")
				}
			}
		}
	})

	t.Run("Parameter Structure", func(t *testing.T) {
		paths := swaggerDoc["paths"].(map[string]interface{})
		for pathKey, pathValue := range paths {
			path := pathValue.(map[string]interface{})
			for method, operation := range path {
				if method == "get" || method == "post" || method == "put" || method == "delete" {
					op := operation.(map[string]interface{})
					if parameters, ok := op["parameters"].([]interface{}); ok {
						for i, param := range parameters {
							p, ok := param.(map[string]interface{})
							require.True(t, ok, "Parameter %d in %s %s should be an object", i, method, pathKey)

							// Required fields
							assert.Contains(t, p, "name", "Parameter should have name")
							assert.Contains(t, p, "in", "Parameter should have 'in' field")
							assert.Contains(t, p, "required", "Parameter should have required field")

							// Validate 'in' field
							in, ok := p["in"].(string)
							require.True(t, ok, "Parameter 'in' should be a string")
							assert.Contains(t, []string{"query", "path", "header", "formData", "body"}, in,
								"Parameter 'in' should be valid location")
						}
					}
				}
			}
		}
	})

	t.Run("Response Structure", func(t *testing.T) {
		paths := swaggerDoc["paths"].(map[string]interface{})
		for pathKey, pathValue := range paths {
			path := pathValue.(map[string]interface{})
			for method, operation := range path {
				if method == "get" || method == "post" || method == "put" || method == "delete" {
					op := operation.(map[string]interface{})
					responses := op["responses"].(map[string]interface{})
					for statusCode, response := range responses {
						resp, ok := response.(map[string]interface{})
						require.True(t, ok, "Response %s in %s %s should be an object", statusCode, method, pathKey)

						// Response should have description or schema
						hasDescription := false
						if _, ok := resp["description"]; ok {
							hasDescription = true
						}
						if _, ok := resp["schema"]; ok {
							hasDescription = true
						}
						assert.True(t, hasDescription, "Response %s should have description or schema", statusCode)
					}
				}
			}
		}
	})

	t.Run("Definitions Section", func(t *testing.T) {
		if definitions, ok := swaggerDoc["definitions"].(map[string]interface{}); ok {
			for defName, defValue := range definitions {
				def, ok := defValue.(map[string]interface{})
				require.True(t, ok, "Definition %s should be an object", defName)
				assert.Contains(t, def, "type", "Definition should have type")
			}
		}
	})
}

// TestParameterValidation tests parameter validation attributes
func TestParameterValidation(t *testing.T) {
	g := gin.New()
	swagger := goswag.NewGin(g)

	swagger.GET("/test/validation").
		Summary("Test parameter validation").
		Tags("test").
		QueryParamWithOptions("required_param", "Required parameter", goswag.StringType, true,
			models.NewParamOptions().
				WithMinLength(3).
				WithMaxLength(50).
				WithPattern("^[a-zA-Z0-9]+$").
				WithExample("example123"),
		).
		QueryParamWithOptions("optional_param", "Optional parameter", goswag.IntType, false,
			models.NewParamOptions().
				WithDefault(10).
				WithMinimum(1).
				WithMaximum(100),
		).
		QueryParamWithOptions("enum_param", "Enum parameter", goswag.StringType, false,
			models.NewParamOptions().
				WithEnum("option1", "option2", "option3").
				WithDefault("option1"),
		).
		Returns([]models.ReturnType{{StatusCode: http.StatusOK}})

	swagger.GenerateSwagger()

	goswagContent, err := os.ReadFile("goswag.go")
	require.NoError(t, err)
	content := string(goswagContent)

	// Check validation attributes in parameter descriptions
	assert.Contains(t, content, "required_param", "Should contain required parameter")
	assert.Contains(t, content, "optional_param", "Should contain optional parameter")
	assert.Contains(t, content, "enum_param", "Should contain enum parameter")
}

// TestGlobalConfigInJSON tests global configuration in Swagger JSON
func TestGlobalConfigInJSON(t *testing.T) {
	_ = SetupEnhancedRoutes()

	swaggerJSONPath := filepath.Join("docs", "swagger.json")
	if _, err := os.Stat(swaggerJSONPath); os.IsNotExist(err) {
		t.Skip("swagger.json not found. Run 'swag init' first.")
		return
	}

	swaggerJSON, err := os.ReadFile(swaggerJSONPath)
	require.NoError(t, err)

	var swaggerDoc map[string]interface{}
	require.NoError(t, json.Unmarshal(swaggerJSON, &swaggerDoc))

	t.Run("Host", func(t *testing.T) {
		if host, ok := swaggerDoc["host"].(string); ok {
			assert.NotEmpty(t, host, "Host should not be empty if present")
		}
	})

	t.Run("BasePath", func(t *testing.T) {
		if basePath, ok := swaggerDoc["basePath"].(string); ok {
			assert.NotEmpty(t, basePath, "BasePath should not be empty if present")
		}
	})

	t.Run("Schemes", func(t *testing.T) {
		if schemes, ok := swaggerDoc["schemes"].([]interface{}); ok {
			assert.NotEmpty(t, schemes, "Schemes should not be empty if present")
			for _, scheme := range schemes {
				s, ok := scheme.(string)
				require.True(t, ok, "Scheme should be a string")
				assert.Contains(t, []string{"http", "https", "ws", "wss"}, s,
					"Scheme should be valid")
			}
		}
	})

	t.Run("Info Contact", func(t *testing.T) {
		info := swaggerDoc["info"].(map[string]interface{})
		if contact, ok := info["contact"].(map[string]interface{}); ok {
			if name, ok := contact["name"].(string); ok {
				assert.NotEmpty(t, name, "Contact name should not be empty")
			}
			if email, ok := contact["email"].(string); ok {
				assert.NotEmpty(t, email, "Contact email should not be empty")
			}
		}
	})

	t.Run("Info License", func(t *testing.T) {
		info := swaggerDoc["info"].(map[string]interface{})
		if license, ok := info["license"].(map[string]interface{}); ok {
			if name, ok := license["name"].(string); ok {
				assert.NotEmpty(t, name, "License name should not be empty")
			}
			if url, ok := license["url"].(string); ok {
				assert.NotEmpty(t, url, "License URL should not be empty")
			}
		}
	})
}

// TestTagMetadataInJSON tests tag metadata in Swagger JSON
func TestTagMetadataInJSON(t *testing.T) {
	_ = SetupEnhancedRoutes()

	swaggerJSONPath := filepath.Join("docs", "swagger.json")
	if _, err := os.Stat(swaggerJSONPath); os.IsNotExist(err) {
		t.Skip("swagger.json not found. Run 'swag init' first.")
		return
	}

	swaggerJSON, err := os.ReadFile(swaggerJSONPath)
	require.NoError(t, err)

	var swaggerDoc map[string]interface{}
	require.NoError(t, json.Unmarshal(swaggerJSON, &swaggerDoc))

	if tags, ok := swaggerDoc["tags"].([]interface{}); ok {
		for _, tag := range tags {
			tagMap, ok := tag.(map[string]interface{})
			require.True(t, ok, "Tag should be an object")

			if name, ok := tagMap["name"].(string); ok {
				assert.NotEmpty(t, name, "Tag name should not be empty")

				// Check for description if present
				if description, ok := tagMap["description"].(string); ok {
					assert.NotEmpty(t, description, "Tag description should not be empty if present")
				}

				// Check for externalDocs if present
				if externalDocs, ok := tagMap["externalDocs"].(map[string]interface{}); ok {
					if url, ok := externalDocs["url"].(string); ok {
						assert.NotEmpty(t, url, "External docs URL should not be empty")
						assert.True(t, strings.HasPrefix(url, "http"), "External docs URL should be valid")
					}
				}
			}
		}
	}
}

// TestComplexScenarios tests complex real-world scenarios
func TestComplexScenarios(t *testing.T) {
	g := gin.New()
	swagger := goswag.NewGin(g)

	// Scenario 1: RESTful API with CRUD operations
	api := swagger.Group("/api/v1")
	users := api.Group("/users")

	users.GET("").
		Summary("List all users").
		Tags("users").
		QueryParamWithOptions("page", "Page number", goswag.IntType, false,
			models.NewParamOptions().WithDefault(1).WithMinimum(1)).
		QueryParamWithOptions("limit", "Items per page", goswag.IntType, false,
			models.NewParamOptions().WithDefault(20).WithMaximum(100)).
		Returns([]models.ReturnType{{StatusCode: http.StatusOK}})

	users.POST("").
		Summary("Create user").
		Tags("users").
		Accepts("application/json").
		Produces("application/json").
		Read(EnhancedUser{}).
		Returns([]models.ReturnType{{StatusCode: http.StatusCreated}})

	users.GET("/{id}").
		Summary("Get user").
		Tags("users").
		PathParamWithOptions("id", "User ID", goswag.StringType, true,
			models.NewParamOptions().WithFormat("uuid")).
		Returns([]models.ReturnType{
			{StatusCode: http.StatusOK},
			{StatusCode: http.StatusNotFound},
		})

	users.PUT("/{id}").
		Summary("Update user").
		Tags("users").
		PathParam("id", "User ID", goswag.StringType, true).
		Read(EnhancedUser{}).
		Returns([]models.ReturnType{{StatusCode: http.StatusOK}})

	users.DELETE("/{id}").
		Summary("Delete user").
		Tags("users").
		PathParam("id", "User ID", goswag.StringType, true).
		Returns([]models.ReturnType{{StatusCode: http.StatusNoContent}})

	swagger.GenerateSwagger()

	goswagContent, err := os.ReadFile("goswag.go")
	require.NoError(t, err)
	content := string(goswagContent)

	// Verify all CRUD operations are present
	assert.Contains(t, content, "GET /api/v1/users", "Should contain list endpoint")
	assert.Contains(t, content, "POST /api/v1/users", "Should contain create endpoint")
	assert.Contains(t, content, "GET /api/v1/users/{id}", "Should contain get endpoint")
	assert.Contains(t, content, "PUT /api/v1/users/{id}", "Should contain update endpoint")
	assert.Contains(t, content, "DELETE /api/v1/users/{id}", "Should contain delete endpoint")
}
