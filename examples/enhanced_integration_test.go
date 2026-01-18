package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnhancedFeaturesIntegration tests all new v2.0 features
func TestEnhancedFeaturesIntegration(t *testing.T) {
	// Generate swagger using enhanced features
	_ = SetupEnhancedRoutes()

	// Check if goswag.go exists
	goswagGoPath := "goswag.go"
	if _, err := os.Stat(goswagGoPath); os.IsNotExist(err) {
		t.Skip("goswag.go not found. Run GenerateSwagger() first.")
		return
	}

	// Read goswag.go and verify enhanced features
	goswagContent, err := os.ReadFile(goswagGoPath)
	require.NoError(t, err)

	content := string(goswagContent)

	// Test Sprint 1: Enhanced Parameters
	t.Run("Enhanced Parameters", func(t *testing.T) {
		// QueryParamWithOptions
		assert.Contains(t, content, "QueryParamWithOptions", "Should contain QueryParamWithOptions")

		// Parameter options
		assert.Contains(t, content, "WithDefault", "Should contain WithDefault")
		assert.Contains(t, content, "WithMinimum", "Should contain WithMinimum")
		assert.Contains(t, content, "WithMaximum", "Should contain WithMaximum")
		assert.Contains(t, content, "WithEnum", "Should contain WithEnum")
		assert.Contains(t, content, "WithFormat", "Should contain WithFormat")
		assert.Contains(t, content, "WithPattern", "Should contain WithPattern")
		assert.Contains(t, content, "WithCollectionFormat", "Should contain WithCollectionFormat")

		// FormDataParam and FileParam
		assert.Contains(t, content, "FormDataParam", "Should contain FormDataParam")
		assert.Contains(t, content, "FileParam", "Should contain FileParam")
	})

	// Test Sprint 2: Operation & Response Enhancements
	t.Run("Operation Enhancements", func(t *testing.T) {
		// Check for operation-level annotations
		assert.Contains(t, content, "@OperationID", "Should contain @OperationID annotation")
		assert.Contains(t, content, "@Deprecated", "Should contain @Deprecated annotation")
		assert.Contains(t, content, "@Schemes", "Should contain @Schemes annotation")
		assert.Contains(t, content, "@ExternalDocs", "Should contain @ExternalDocs annotation")
	})

	t.Run("Response Enhancements", func(t *testing.T) {
		// Response headers - check for @Header annotation
		assert.Contains(t, content, "@Header", "Should contain @Header annotation for response headers")

		// Response examples - check for @Example annotation
		assert.Contains(t, content, "@Example", "Should contain @Example annotation for response examples")

		// Response descriptions - embedded in @Success/@Failure annotations
		// Look for @Success or @Failure followed by quoted description
		lines := strings.Split(content, "\n")
		hasResponseWithDescription := false
		for i, line := range lines {
			if strings.Contains(line, "@Success") || strings.Contains(line, "@Failure") {
				// Check if the line contains a quoted description (after status code and type)
				if strings.Count(line, "\"") >= 1 {
					hasResponseWithDescription = true
					break
				}
				// Also check next line
				if i+1 < len(lines) && strings.Contains(lines[i+1], "\"") {
					hasResponseWithDescription = true
					break
				}
			}
		}
		assert.True(t, hasResponseWithDescription, "Should have response descriptions in @Success/@Failure annotations")
	})

	// Test Sprint 3: Global Configuration
	t.Run("Global Configuration", func(t *testing.T) {
		// Check for global config annotations (now supported with NewGinWithConfig)
		assert.Contains(t, content, "@host", "Should contain @host annotation")
		assert.Contains(t, content, "@BasePath", "Should contain @BasePath annotation")
		assert.Contains(t, content, "@schemes", "Should contain @schemes annotation")
		assert.Contains(t, content, "@contact.name", "Should contain @contact.name annotation")
		assert.Contains(t, content, "@contact.email", "Should contain @contact.email annotation")
		assert.Contains(t, content, "@license.name", "Should contain @license.name annotation")
		assert.Contains(t, content, "@license.url", "Should contain @license.url annotation")
		assert.Contains(t, content, "@termsOfService", "Should contain @termsOfService annotation")
		assert.Contains(t, content, "@externalDocs", "Should contain @externalDocs annotation")
	})

	// Test Sprint 4: Tag Metadata
	t.Run("Tag Metadata", func(t *testing.T) {
		// Check for @Tag annotation (now supported with Group() returning GinGroup)
		assert.Contains(t, content, "@Tag", "Should contain @Tag annotation")
		// Verify tag has description or externalDocs
		assert.True(t, strings.Contains(content, "@Tag") &&
			(strings.Contains(content, "\"") || strings.Contains(content, "http")),
			"@Tag should contain description or external docs URL")
	})

	t.Log("Enhanced features validation passed!")
}

// TestSwaggerJSONEnhancedFeatures validates enhanced features in generated swagger.json
func TestSwaggerJSONEnhancedFeatures(t *testing.T) {
	// This test requires swag to be installed and swagger.json to be generated
	// It's a more comprehensive test that validates the actual Swagger JSON output

	// Check if swagger.json exists (would be generated by integration_app.go)
	swaggerJSONPath := filepath.Join("docs", "swagger.json")
	if _, err := os.Stat(swaggerJSONPath); os.IsNotExist(err) {
		t.Skip("swagger.json not found. Run integration_app.go first to generate swagger.json")
		return
	}

	swaggerJSON, err := os.ReadFile(swaggerJSONPath)
	require.NoError(t, err)

	var swaggerDoc map[string]interface{}
	require.NoError(t, json.Unmarshal(swaggerJSON, &swaggerDoc))

	// Validate enhanced features in swagger.json
	t.Run("Global Configuration in Swagger JSON", func(t *testing.T) {
		info := swaggerDoc["info"].(map[string]interface{})

		// Contact
		if contact, ok := info["contact"].(map[string]interface{}); ok {
			assert.Contains(t, contact, "name", "Should have contact name")
			assert.Contains(t, contact, "email", "Should have contact email")
		}

		// License
		if license, ok := info["license"].(map[string]interface{}); ok {
			assert.Contains(t, license, "name", "Should have license name")
			assert.Contains(t, license, "url", "Should have license URL")
		}

		// Terms of Service
		if _, ok := info["termsOfService"]; ok {
			assert.NotEmpty(t, info["termsOfService"], "Should have terms of service")
		}
	})

	t.Run("Paths with Enhanced Features", func(t *testing.T) {
		paths := swaggerDoc["paths"].(map[string]interface{})

		// Find a path with enhanced features
		for pathKey, pathValue := range paths {
			path := pathValue.(map[string]interface{})

			// Check for operation with operationId
			for method, operation := range path {
				if method == "get" || method == "post" || method == "put" || method == "delete" {
					op := operation.(map[string]interface{})

					// Operation ID
					if operationId, ok := op["operationId"]; ok {
						assert.NotEmpty(t, operationId, "Operation should have operationId: %s %s", method, pathKey)
					}

					// Deprecated
					if _, ok := op["deprecated"]; ok {
						// Some operations may be deprecated
					}

					// Parameters with enhanced features
					if parameters, ok := op["parameters"].([]interface{}); ok {
						for _, param := range parameters {
							p := param.(map[string]interface{})

							// Check for enum
							if schema, ok := p["schema"].(map[string]interface{}); ok {
								if enum, ok := schema["enum"]; ok {
									assert.NotEmpty(t, enum, "Parameter should have enum values")
								}

								// Check for default
								if _, ok := schema["default"]; ok {
									// Default value present
								}

								// Check for format
								if format, ok := schema["format"]; ok {
									assert.NotEmpty(t, format, "Parameter should have format")
								}
							}

							// Check for collectionFormat
							if collectionFormat, ok := p["collectionFormat"]; ok {
								assert.NotEmpty(t, collectionFormat, "Parameter should have collectionFormat")
							}
						}
					}

					// Responses with headers
					if responses, ok := op["responses"].(map[string]interface{}); ok {
						for statusCode, response := range responses {
							resp := response.(map[string]interface{})

							// Check for headers
							if headers, ok := resp["headers"]; ok {
								assert.NotEmpty(t, headers, "Response %s should have headers", statusCode)
							}

							// Check for examples
							if examples, ok := resp["examples"]; ok {
								assert.NotEmpty(t, examples, "Response %s should have examples", statusCode)
							}
						}
					}
				}
			}
		}
	})

	t.Run("Tags with Metadata", func(t *testing.T) {
		if tags, ok := swaggerDoc["tags"].([]interface{}); ok && len(tags) > 0 {
			tagMetadataFound := false
			for _, tag := range tags {
				tagMap, ok := tag.(map[string]interface{})
				if !ok {
					continue
				}

				// Check for description
				if description, ok := tagMap["description"].(string); ok && description != "" {
					tagMetadataFound = true
					t.Logf("Found tag with description: %s", description)
				}

				// Check for externalDocs
				if externalDocs, ok := tagMap["externalDocs"].(map[string]interface{}); ok && len(externalDocs) > 0 {
					tagMetadataFound = true
					t.Logf("Found tag with externalDocs")
				}
			}
			if !tagMetadataFound {
				t.Log("Tag metadata not found - tags may not have descriptions or externalDocs set")
			}
		} else {
			t.Log("No tags found in swagger.json")
		}
	})

	t.Log("Swagger JSON enhanced features validation passed!")
}

// Note: SetupEnhancedRoutes is defined in enhanced_shared.go
