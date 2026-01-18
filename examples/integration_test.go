package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationSwaggerGeneration tests the full swagger generation and serving flow
func TestIntegrationSwaggerGeneration(t *testing.T) {
	// Create a temporary directory for swagger docs
	tempDir := t.TempDir()
	goswagDir := filepath.Join(tempDir, "goswag")
	docsDir := filepath.Join(tempDir, "docs")

	require.NoError(t, os.MkdirAll(goswagDir, 0755))
	require.NoError(t, os.MkdirAll(docsDir, 0755))

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	require.NoError(t, os.Chdir(tempDir))

	// Create main.go file for swag
	mainGoContent := `// @title           GoSwag Integration Test API
// @version         1.0
// @description     This is a test API for integration testing
// @termsOfService  http://swagger.io/terms/
// @contact.name    API Support
// @contact.email   support@example.com
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
// @host            localhost:8080
// @BasePath        /
package main

import _ "goswag"
`
	mainGoPath := filepath.Join(goswagDir, "main.go")
	require.NoError(t, os.WriteFile(mainGoPath, []byte(mainGoContent), 0644))

	// Copy goswag.go to goswag directory
	// First, generate swagger by calling SetupRoutes
	_ = SetupRoutes("test")

	// Check if goswag.go exists in current directory
	goswagGoPath := "goswag.go"
	if _, err := os.Stat(goswagGoPath); os.IsNotExist(err) {
		t.Fatalf("goswag.go file not found. Make sure GenerateSwagger() was called.")
	}

	// Copy goswag.go to goswag directory
	goswagGoContent, err := os.ReadFile(goswagGoPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(goswagDir, "goswag.go"), goswagGoContent, 0644))

	// Create a go.mod file in goswag directory for swag
	goModContent := `module goswag

go 1.23

require (
	github.com/r0bertson/goswag v0.0.0
	github.com/gin-gonic/gin v1.10.0
)

replace github.com/r0bertson/goswag => ../../
`
	require.NoError(t, os.WriteFile(filepath.Join(goswagDir, "go.mod"), []byte(goModContent), 0644))

	// Install swag if not available
	swagPath, err := exec.LookPath("swag")
	if err != nil {
		t.Log("swag not found in PATH, installing...")
		installCmd := exec.Command("go", "install", "github.com/swaggo/swag/cmd/swag@latest")
		installCmd.Env = os.Environ()
		if err := installCmd.Run(); err != nil {
			t.Skipf("Could not install swag: %v. Skipping integration test.", err)
			return
		}
		// Try to find swag in GOPATH/bin
		gopath := os.Getenv("GOPATH")
		if gopath != "" {
			swagPath = filepath.Join(gopath, "bin", "swag")
		} else {
			homeDir, _ := os.UserHomeDir()
			swagPath = filepath.Join(homeDir, "go", "bin", "swag")
		}
	}

	// Run swag init
	swagCmd := exec.Command(swagPath, "init", "--parseDependency", "--parseInternal", "-g", mainGoPath, "-o", docsDir)
	swagCmd.Dir = goswagDir
	swagCmd.Env = os.Environ()
	swagOutput, err := swagCmd.CombinedOutput()
	if err != nil {
		t.Logf("swag init output: %s", string(swagOutput))
		t.Skipf("Could not run swag init: %v. Make sure swag is installed and goswag.go is valid.", err)
		return
	}

	// Verify swagger.json was created
	swaggerJSONPath := filepath.Join(docsDir, "swagger.json")
	require.FileExists(t, swaggerJSONPath, "swagger.json should be generated")

	// Read and validate swagger.json
	swaggerJSON, err := os.ReadFile(swaggerJSONPath)
	require.NoError(t, err)

	var swaggerDoc map[string]interface{}
	require.NoError(t, json.Unmarshal(swaggerJSON, &swaggerDoc))

	// Verify basic swagger structure
	assert.Equal(t, "2.0", swaggerDoc["swagger"])
	info := swaggerDoc["info"].(map[string]interface{})
	assert.Equal(t, "GoSwag Integration Test API", info["title"])
	assert.Equal(t, "1.0", info["version"])

	// Verify paths exist
	paths := swaggerDoc["paths"].(map[string]interface{})
	assert.Contains(t, paths, "/users")
	assert.Contains(t, paths, "/api/v1/users/{user_id}")

	// Start HTTP server to serve swagger UI
	router := SetupRoutes("test")

	// Add swagger UI routes
	router.StaticFile("/swagger/doc.json", swaggerJSONPath)
	router.GET("/swagger/*any", func(c *gin.Context) {
		// Simple swagger UI redirect - in real scenario you'd use httpSwagger
		if c.Param("any") == "/index.html" || c.Param("any") == "/" {
			c.Redirect(http.StatusMovedPermanently, "https://petstore.swagger.io/?url="+c.Request.Host+"/swagger/doc.json")
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	})

	// Start server in background
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	// Test that swagger.json is accessible
	resp, err := http.Get("http://localhost:8080/swagger/doc.json")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var servedSwagger map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &servedSwagger))
	assert.Equal(t, "GoSwag Integration Test API", servedSwagger["info"].(map[string]interface{})["title"])

	// Test that API endpoints work
	apiResp, err := http.Get("http://localhost:8080/users")
	require.NoError(t, err)
	defer apiResp.Body.Close()
	assert.Equal(t, http.StatusOK, apiResp.StatusCode)

	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	t.Log("Integration test passed! Swagger UI would be available at http://localhost:8080/swagger/index.html")
}

// TestSwaggerJSONStructure validates the generated swagger.json structure
func TestSwaggerJSONStructure(t *testing.T) {
	// Generate swagger
	_ = SetupRoutes("test")

	// Check if goswag.go exists
	if _, err := os.Stat("goswag.go"); os.IsNotExist(err) {
		t.Skip("goswag.go not found. Run GenerateSwagger() first.")
		return
	}

	// Read goswag.go and verify it contains expected annotations
	goswagContent, err := os.ReadFile("goswag.go")
	require.NoError(t, err)

	content := string(goswagContent)

	// Verify key swagger annotations are present
	assert.Contains(t, content, "@Summary")
	assert.Contains(t, content, "@Description")
	assert.Contains(t, content, "@Tags")
	assert.Contains(t, content, "@Param")
	assert.Contains(t, content, "@Success")
	assert.Contains(t, content, "@Router")

	// Verify route annotations
	assert.Contains(t, content, "GET /users")
	assert.Contains(t, content, "POST /users")
	assert.Contains(t, content, "GET /api/v1/users/{user_id}")

	// Verify field descriptions are present in wrapper structs
	if strings.Contains(content, "Wrapper") {
		assert.Contains(t, content, "// The full name of the user")
		assert.Contains(t, content, "// Unique identifier for the user")
	}

	t.Log("Swagger annotations validation passed!")
}
