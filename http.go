package goswag

import (
	"net/http"

	httpWrapper "github.com/r0bertson/goswag/internal/frameworks/http"
	"github.com/r0bertson/goswag/models"
)

type HTTP interface {
	models.HTTPRouter
	models.HTTPGroup
	GenerateSwagger()
	Mux() *http.ServeMux
}

// NewHTTP returns the interface that wraps the basic HTTP methods and add the swagger methods
// defaultResponses is an optional parameter that can be used to set the default responses for all routes
func NewHTTP(mux *http.ServeMux, defaultResponses ...models.ReturnType) HTTP {
	return httpWrapper.NewHTTP(mux, defaultResponses...)
}

// NewHTTPWithConfig creates a new HTTP swagger instance with global configuration.
// Use this when you want to set host, basePath, schemes, contact, license, etc. globally.
// Example:
//   config := models.NewSwaggerConfig().
//       WithHost("api.example.com").
//       WithBasePath("/v1").
//       WithSchemes("https").
//       WithContact(models.NewContactInfo("API Support", "support@example.com", "")).
//       WithLicense(models.NewLicenseInfo("MIT", "https://opensource.org/licenses/MIT"))
//   gh := goswag.NewHTTPWithConfig(mux, config, defaultResponses...)
func NewHTTPWithConfig(mux *http.ServeMux, config *models.SwaggerConfig, defaultResponses ...models.ReturnType) HTTP {
	return httpWrapper.NewHTTPWithConfig(mux, config, defaultResponses...)
}