package models

// ExternalDocs provides external documentation links for operations, tags, or the API itself.
type ExternalDocs struct {
	URL         string
	Description string
}

// NewExternalDocs creates a new ExternalDocs instance.
func NewExternalDocs(url, description string) *ExternalDocs {
	return &ExternalDocs{
		URL:         url,
		Description: description,
	}
}
