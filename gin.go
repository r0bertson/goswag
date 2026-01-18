package goswag

import (
	"github.com/gin-gonic/gin"
	ginWrapper "github.com/r0bertson/goswag/internal/frameworks/gin"
	"github.com/r0bertson/goswag/models"
)

type Gin interface {
	models.GinRouter
	models.GinGroup
	GenerateSwagger()
	Gin() *gin.Engine
}

// NewGin returns the interface that wraps the basic Gin methods and add the swagger methods
// defaultResponses is an optional parameter that can be used to set the default responses for all routes
func NewGin(g *gin.Engine, defaultResponses ...models.ReturnType) Gin {
	return ginWrapper.NewGin(g, defaultResponses...)
}

// NewGinWithConfig creates a new Gin swagger instance with global configuration.
// Use this when you want to set host, basePath, schemes, contact, license, etc. globally.
// Example:
//   config := models.NewSwaggerConfig().
//       WithHost("api.example.com").
//       WithBasePath("/v1").
//       WithSchemes("https").
//       WithContact(models.NewContactInfo("API Support", "support@example.com", "")).
//       WithLicense(models.NewLicenseInfo("MIT", "https://opensource.org/licenses/MIT"))
//   gh := goswag.NewGinWithConfig(g, config, defaultResponses...)
func NewGinWithConfig(g *gin.Engine, config *models.SwaggerConfig, defaultResponses ...models.ReturnType) Gin {
	return ginWrapper.NewGinWithConfig(g, config, defaultResponses...)
}