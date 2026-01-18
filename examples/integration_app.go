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

	"github.com/gin-gonic/gin"
)

// @title           GoSwag Integration Test API
// @version         1.0
// @description     Integration test API for goswag library
// @termsOfService  http://swagger.io/terms/
// @contact.name    API Support
// @contact.email   support@example.com
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
// @host            localhost:8080
// @BasePath        /
func main() {
	// Create a temporary directory for this test run
	tempDir, err := os.MkdirTemp("", "goswag-integration-*")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	goswagDir := filepath.Join(tempDir, "goswag")
	docsDir := filepath.Join(tempDir, "docs")

	if err := os.MkdirAll(goswagDir, 0755); err != nil {
		log.Fatalf("Failed to create goswag directory: %v", err)
	}
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		log.Fatalf("Failed to create docs directory: %v", err)
	}

	// Get original directory (examples directory) and calculate goswag module root (parent directory)
	originalDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	// The goswag module root is the parent of the examples directory
	goswagModuleRoot := filepath.Dir(originalDir)

	// Setup routes and generate swagger (generates goswag.go in current directory)
	router := SetupRoutes("release")

	// Read goswag.go from original directory before changing directories
	goswagGoPath := filepath.Join(originalDir, "goswag.go")
	goswagGoContent, err := os.ReadFile(goswagGoPath)
	if err != nil {
		log.Fatalf("Failed to read goswag.go from %s: %v", goswagGoPath, err)
	}

	// Don't change directories - work with absolute paths instead
	// Changing directories can cause path resolution issues with swag

	if err := os.WriteFile(filepath.Join(goswagDir, "goswag.go"), goswagGoContent, 0644); err != nil {
		log.Fatalf("Failed to write goswag.go: %v", err)
	}

	// Copy shared.go to temp directory so swag can resolve type definitions (User, CreateUserRequest, etc.)
	sharedGoPath := filepath.Join(originalDir, "shared.go")
	sharedGoContent, err := os.ReadFile(sharedGoPath)
	if err != nil {
		log.Printf("Warning: Failed to read shared.go from %s: %v", sharedGoPath, err)
		log.Println("Swag may not be able to resolve all type definitions")
	} else {
		if err := os.WriteFile(filepath.Join(goswagDir, "shared.go"), sharedGoContent, 0644); err != nil {
			log.Printf("Warning: Failed to write shared.go: %v", err)
		}
	}

	// Create main.go for swag
	// Note: shared.go will be in the same package, so swag will parse it automatically
	mainGoContent := `// @title           GoSwag Integration Test API
// @version         1.0
// @description     Integration test API for goswag library
// @host            localhost:8080
// @BasePath        /
package main

import _ "swagger-temp/goswag"
`
	if err := os.WriteFile(filepath.Join(goswagDir, "main.go"), []byte(mainGoContent), 0644); err != nil {
		log.Fatalf("Failed to write main.go: %v", err)
	}

	// Create go.mod for swag
	// Calculate relative path from goswagDir (where go.mod is) to goswag module root for replace directive
	// The replace directive must point to the directory containing the actual go.mod file
	relPath, err := filepath.Rel(goswagDir, goswagModuleRoot)
	if err != nil {
		// Fallback to absolute path if relative path calculation fails
		goswagModuleRootAbs, absErr := filepath.Abs(goswagModuleRoot)
		if absErr != nil {
			log.Fatalf("Failed to get path for goswag module root: %v", absErr)
		}
		// Use forward slashes for go.mod (works on Windows too)
		relPath = filepath.ToSlash(goswagModuleRootAbs)
	} else {
		// Normalize relative path separators
		relPath = filepath.ToSlash(relPath)
		// Ensure it starts with .. if it's outside the goswag directory
		if !filepath.IsAbs(relPath) && !strings.HasPrefix(relPath, "..") && relPath != "." {
			relPath = "../" + relPath
		}
	}
	goModContent := fmt.Sprintf(`module swagger-temp

go 1.23

require (
	github.com/r0bertson/goswag v0.0.0
	github.com/gin-gonic/gin v1.10.0
)

replace github.com/r0bertson/goswag => %s
`, relPath)

	if err := os.WriteFile(filepath.Join(goswagDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		log.Fatalf("Failed to write go.mod: %v", err)
	}

	// Initialize go module and download dependencies before running swag
	// This ensures the module is properly resolved
	// First try go mod download to fetch dependencies
	goModDownloadCmd := exec.Command("go", "mod", "download")
	goModDownloadCmd.Dir = goswagDir
	goModDownloadCmd.Env = os.Environ()
	if output, err := goModDownloadCmd.CombinedOutput(); err != nil {
		log.Printf("Warning: go mod download failed: %v", err)
		log.Printf("Output: %s", string(output))
	}

	// Then try go mod tidy to clean up
	goModCmd := exec.Command("go", "mod", "tidy")
	goModCmd.Dir = goswagDir
	goModCmd.Env = os.Environ()
	if output, err := goModCmd.CombinedOutput(); err != nil {
		log.Printf("Warning: go mod tidy failed: %v", err)
		log.Printf("Output: %s", string(output))
		// Continue anyway, swag might still work
	}

	// Try to run swag init
	swagPath, err := exec.LookPath("swag")
	if err != nil {
		log.Printf("Warning: swag not found in PATH. Attempting to install...")
		installCmd := exec.Command("go", "install", "github.com/swaggo/swag/cmd/swag@latest")
		if err := installCmd.Run(); err != nil {
			log.Printf("Could not install swag: %v", err)
			log.Println("Serving API without Swagger UI. Install swag to enable Swagger UI:")
			log.Println("  go install github.com/swaggo/swag/cmd/swag@latest")
			log.Println("  swag init -g ./goswag/main.go -o ./docs")
			serveAPIOnly(router)
			return
		}
		// Try to find swag
		gopath := os.Getenv("GOPATH")
		if gopath != "" {
			swagPath = filepath.Join(gopath, "bin", "swag")
		} else {
			homeDir, _ := os.UserHomeDir()
			swagPath = filepath.Join(homeDir, "go", "bin", "swag")
		}
	}

	// Run swag init
	// Use relative paths since we're setting the working directory to goswagDir
	// Calculate relative path from goswagDir to docsDir
	docsRelPath, err := filepath.Rel(goswagDir, docsDir)
	if err != nil {
		log.Printf("Warning: Failed to calculate relative path for docs: %v", err)
		serveAPIOnly(router)
		return
	}
	// Normalize path separators
	docsRelPath = filepath.ToSlash(docsRelPath)
	swagCmd := exec.Command(swagPath, "init", "--parseDependency", "--parseInternal", "-g", "main.go", "-o", docsRelPath)
	swagCmd.Dir = goswagDir
	swagCmd.Env = os.Environ()

	if output, err := swagCmd.CombinedOutput(); err != nil {
		log.Printf("Warning: swag init failed: %v", err)
		log.Printf("Output: %s", string(output))
		log.Println("Serving API without Swagger UI.")
		serveAPIOnly(router)
		return
	}

	// Verify swagger.json was created
	swaggerJSONPath := filepath.Join(docsDir, "swagger.json")
	if _, err := os.Stat(swaggerJSONPath); os.IsNotExist(err) {
		log.Println("Warning: swagger.json not generated. Serving API without Swagger UI.")
		serveAPIOnly(router)
		return
	}

	// Read and validate swagger.json
	swaggerJSON, err := os.ReadFile(swaggerJSONPath)
	if err != nil {
		log.Printf("Warning: Failed to read swagger.json: %v", err)
		serveAPIOnly(router)
		return
	}

	var swaggerDoc map[string]interface{}
	if err := json.Unmarshal(swaggerJSON, &swaggerDoc); err != nil {
		log.Printf("Warning: Invalid swagger.json: %v", err)
		serveAPIOnly(router)
		return
	}

	// Post-process swagger.json to remove pointer fields from required arrays
	// Swaggo sometimes includes pointer fields in required arrays even with omitempty
	if err := fixPointerFieldsInSwagger(swaggerDoc, goswagDir); err != nil {
		log.Printf("Warning: Failed to fix pointer fields in swagger: %v", err)
		// Continue anyway, swagger might still work
	} else {
		// Write the fixed swagger.json back
		fixedJSON, err := json.MarshalIndent(swaggerDoc, "", "  ")
		if err != nil {
			log.Printf("Warning: Failed to marshal fixed swagger.json: %v", err)
		} else {
			if err := os.WriteFile(swaggerJSONPath, fixedJSON, 0644); err != nil {
				log.Printf("Warning: Failed to write fixed swagger.json: %v", err)
			}
		}
	}

	info := swaggerDoc["info"].(map[string]interface{})
	log.Printf("✓ Swagger documentation generated successfully!")
	log.Printf("  Title: %s", info["title"])
	log.Printf("  Version: %s", info["version"])

	// Serve swagger.json
	router.StaticFile("/swagger/doc.json", swaggerJSONPath)

	// Serve Swagger UI HTML page (embedded to avoid mixed-content issues)
	// Loading HTTPS resources from HTTP page is allowed (opposite of mixed-content)
	swaggerHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Swagger UI</title>
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
          // Enhance display of nullable/optional fields after Swagger UI loads
          setTimeout(function() {
            // Find all nullable properties and mark them
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

	router.GET("/swagger", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, swaggerHTML)
	})

	router.GET("/swagger/index.html", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, swaggerHTML)
	})

	log.Println("\n🚀 Server starting on http://localhost:8080")
	log.Println("📚 Swagger UI available at: http://localhost:8080/swagger/index.html")
	log.Println("📄 Swagger JSON available at: http://localhost:8080/swagger/doc.json")
	log.Println("\nAPI Endpoints:")
	log.Println("  GET    /users")
	log.Println("  POST   /users")
	log.Println("  GET    /api/v1/users/{user_id}")
	log.Println("  PUT    /api/v1/users/{user_id}")
	log.Println("  DELETE /api/v1/users/{user_id}")
	log.Println("\nPress Ctrl+C to stop the server")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func serveAPIOnly(router *gin.Engine) {
	log.Println("\n🚀 Server starting on http://localhost:8080")
	log.Println("API Endpoints:")
	log.Println("  GET    /users")
	log.Println("  POST   /users")
	log.Println("  GET    /api/v1/users/{user_id}")
	log.Println("  PUT    /api/v1/users/{user_id}")
	log.Println("  DELETE /api/v1/users/{user_id}")
	log.Println("\nNote: Install swag to enable Swagger UI")
	log.Println("  go install github.com/swaggo/swag/cmd/swag@latest")
	log.Println("\nPress Ctrl+C to stop the server")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// fixPointerFieldsInSwagger removes pointer fields from required arrays in swagger definitions
// by analyzing the Go source files to identify pointer fields
func fixPointerFieldsInSwagger(swaggerDoc map[string]interface{}, goswagDir string) error {
	definitions, ok := swaggerDoc["definitions"].(map[string]interface{})
	if !ok {
		return nil // No definitions to process
	}

	// Read goswag.go to identify pointer fields
	goswagGoPath := filepath.Join(goswagDir, "goswag.go")
	goswagContent, err := os.ReadFile(goswagGoPath)
	if err != nil {
		return fmt.Errorf("failed to read goswag.go: %w", err)
	}

	// Parse Go source to find pointer fields in structs
	pointerFields := findPointerFieldsInSource(string(goswagContent))

	// Process each definition
	for defName, defValue := range definitions {
		def, ok := defValue.(map[string]interface{})
		if !ok {
			continue
		}

		// Get properties
		properties, ok := def["properties"].(map[string]interface{})
		if !ok {
			continue
		}

		// Get required array
		required, ok := def["required"].([]interface{})
		if !ok {
			required = []interface{}{}
		}

		// Find pointer fields for this definition
		var pointerFieldNames []string
		for structName, fields := range pointerFields {
			// Check if definition name matches or contains struct name
			if strings.Contains(defName, structName) {
				pointerFieldNames = append(pointerFieldNames, fields...)
			}
		}

		// Process each property to mark pointer fields as nullable and remove defaults
		for propName, propValue := range properties {
			prop, ok := propValue.(map[string]interface{})
			if !ok {
				continue
			}

			// Check if this is a pointer field
			isPointer := false
			for _, fieldName := range pointerFieldNames {
				if propName == fieldName {
					isPointer = true
					break
				}
			}

			if isPointer {
				// Mark as nullable (Swagger 2.0 uses x-nullable extension)
				prop["x-nullable"] = true

				// Remove default values that make fields look required
				delete(prop, "default")

				// Set example to null to make it clear it's optional
				prop["x-example"] = nil

				// For integer/number types, remove the default 0
				if propType, ok := prop["type"].(string); ok {
					if propType == "integer" || propType == "number" {
						// Don't show 0 as default
						if prop["default"] == float64(0) || prop["default"] == 0 {
							delete(prop, "default")
						}
					}
					// For boolean, don't show false as default
					if propType == "boolean" {
						if prop["default"] == false {
							delete(prop, "default")
						}
					}
					// For string, don't show empty string as default
					if propType == "string" {
						if prop["default"] == "" {
							delete(prop, "default")
						}
					}
				}
			}
		}

		// Remove pointer fields from required array
		var newRequired []interface{}
		for _, reqField := range required {
			reqFieldStr, ok := reqField.(string)
			if !ok {
				newRequired = append(newRequired, reqField)
				continue
			}

			// Check if this field is a pointer field
			isPointer := false
			for _, fieldName := range pointerFieldNames {
				if reqFieldStr == fieldName {
					isPointer = true
					break
				}
			}

			// Only add to required if it's not a pointer field
			if !isPointer {
				newRequired = append(newRequired, reqField)
			}
		}

		// Update required array
		if len(newRequired) > 0 {
			def["required"] = newRequired
		} else {
			// Remove required field if empty (all fields optional)
			delete(def, "required")
		}
	}

	return nil
}

// findPointerFieldsInSource parses Go source code to find pointer fields in structs
func findPointerFieldsInSource(source string) map[string][]string {
	result := make(map[string][]string)
	lines := strings.Split(source, "\n")

	var currentStruct string
	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Detect struct definition
		if strings.HasPrefix(line, "type ") && strings.Contains(line, "struct {") {
			// Extract struct name
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentStruct = parts[1]
				result[currentStruct] = []string{}
			}
			continue
		}

		// Detect struct end
		if line == "}" && currentStruct != "" {
			currentStruct = ""
			continue
		}

		// Detect pointer fields (lines with *Type)
		if currentStruct != "" && strings.Contains(line, "*") {
			// Extract field name and JSON tag
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				fieldName := parts[0]
				// Look for json tag
				for j := i; j < len(lines) && j < i+3; j++ {
					if strings.Contains(lines[j], "json:") {
						// Extract JSON field name from tag
						jsonTag := extractJSONFieldName(lines[j])
						if jsonTag != "" {
							result[currentStruct] = append(result[currentStruct], jsonTag)
						} else {
							// Fallback to Go field name (lowercase first letter)
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

// extractJSONFieldName extracts the JSON field name from a struct tag line
func extractJSONFieldName(tagLine string) string {
	// Look for json:"fieldname" or json:"fieldname,omitempty"
	start := strings.Index(tagLine, `json:"`)
	if start == -1 {
		return ""
	}
	start += 6 // len(`json:"`)
	end := strings.Index(tagLine[start:], `"`)
	if end == -1 {
		return ""
	}
	jsonTag := tagLine[start : start+end]
	// Remove omitempty and other options
	parts := strings.Split(jsonTag, ",")
	return parts[0]
}
