package models

// ResponseHeader defines a response header in Swagger 2.0.
type ResponseHeader struct {
	Type        string // Header type: string, integer, number, boolean
	Description string // Header description
	Format      string // Optional format (e.g., "int32", "int64", "date", "date-time")
}

// NewResponseHeader creates a new ResponseHeader instance.
func NewResponseHeader(headerType, description string) *ResponseHeader {
	return &ResponseHeader{
		Type:        headerType,
		Description: description,
	}
}

// WithFormat sets the format for the response header.
func (h *ResponseHeader) WithFormat(format string) *ResponseHeader {
	h.Format = format
	return h
}
