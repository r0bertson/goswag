package models

// SwaggerConfig provides global configuration for Swagger documentation.
// This configuration is used to generate the top-level Swagger info object.
type SwaggerConfig struct {
	// Host is the host (name or IP) serving the API.
	// This MUST be the host only and does not include the scheme nor sub-paths.
	// It MAY include a port. Example: "api.example.com" or "localhost:8080"
	Host string

	// BasePath is the base path on which the API is served, relative to the host.
	// If it is not included, the API is served directly under the host.
	// The value MUST start with a leading slash (/). Example: "/v1" or "/api/v2"
	BasePath string

	// Schemes specifies the transfer protocols of the API.
	// Valid values: "http", "https", "ws", "wss"
	Schemes []string

	// Contact provides contact information for the API.
	Contact *ContactInfo

	// License provides license information for the API.
	License *LicenseInfo

	// TermsOfService is a URL to the Terms of Service for the API.
	TermsOfService string

	// ExternalDocs provides external documentation links for the API.
	ExternalDocs *ExternalDocs

	// GlobalSecurity specifies default security requirements applied to all operations.
	// Each string should match a security scheme name defined in securityDefinitions.
	// Operations can override this by specifying their own security requirements.
	GlobalSecurity []string
}

// ContactInfo provides contact information for the API.
type ContactInfo struct {
	Name  string // Contact name
	Email string // Contact email
	URL   string // Contact URL
}

// NewContactInfo creates a new ContactInfo instance.
func NewContactInfo(name, email, url string) *ContactInfo {
	return &ContactInfo{
		Name:  name,
		Email: email,
		URL:   url,
	}
}

// LicenseInfo provides license information for the API.
type LicenseInfo struct {
	Name string // License name (e.g., "Apache 2.0", "MIT")
	URL  string // License URL
}

// NewLicenseInfo creates a new LicenseInfo instance.
func NewLicenseInfo(name, url string) *LicenseInfo {
	return &LicenseInfo{
		Name: name,
		URL:  url,
	}
}

// NewSwaggerConfig creates a new SwaggerConfig with default values.
func NewSwaggerConfig() *SwaggerConfig {
	return &SwaggerConfig{}
}

// WithHost sets the host for the API.
func (c *SwaggerConfig) WithHost(host string) *SwaggerConfig {
	c.Host = host
	return c
}

// WithBasePath sets the base path for the API.
func (c *SwaggerConfig) WithBasePath(basePath string) *SwaggerConfig {
	c.BasePath = basePath
	return c
}

// WithSchemes sets the allowed schemes for the API.
func (c *SwaggerConfig) WithSchemes(schemes ...string) *SwaggerConfig {
	c.Schemes = schemes
	return c
}

// WithContact sets the contact information for the API.
func (c *SwaggerConfig) WithContact(contact *ContactInfo) *SwaggerConfig {
	c.Contact = contact
	return c
}

// WithLicense sets the license information for the API.
func (c *SwaggerConfig) WithLicense(license *LicenseInfo) *SwaggerConfig {
	c.License = license
	return c
}

// WithTermsOfService sets the Terms of Service URL for the API.
func (c *SwaggerConfig) WithTermsOfService(url string) *SwaggerConfig {
	c.TermsOfService = url
	return c
}

// WithExternalDocs sets the external documentation for the API.
func (c *SwaggerConfig) WithExternalDocs(externalDocs *ExternalDocs) *SwaggerConfig {
	c.ExternalDocs = externalDocs
	return c
}

// WithGlobalSecurity sets the global security requirements for all operations.
func (c *SwaggerConfig) WithGlobalSecurity(schemes ...string) *SwaggerConfig {
	c.GlobalSecurity = schemes
	return c
}
