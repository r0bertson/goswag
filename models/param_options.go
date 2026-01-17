package models

// ParamOptions provides extended options for parameter definitions in Swagger 2.0.
// All fields are optional and can be used to enhance parameter documentation and validation.
type ParamOptions struct {
	// Default is the default value for the parameter.
	// This will be used when the parameter is not provided.
	Default interface{}

	// Example is an example value for the parameter.
	// This helps API consumers understand what values are expected.
	Example interface{}

	// CollectionFormat specifies how array parameters are serialized.
	// Valid values: "csv", "ssv", "tsv", "pipes", "multi"
	// - csv: comma-separated values (default for query/formData)
	// - ssv: space-separated values
	// - tsv: tab-separated values
	// - pipes: pipe-separated values
	// - multi: multiple parameter instances (e.g., ?id=1&id=2)
	CollectionFormat string

	// AllowEmptyValue indicates whether the parameter allows an empty value.
	// Only applicable for query and formData parameters.
	AllowEmptyValue bool

	// Minimum is the minimum value for numeric parameters.
	Minimum *float64

	// Maximum is the maximum value for numeric parameters.
	Maximum *float64

	// MinLength is the minimum length for string parameters.
	MinLength *int

	// MaxLength is the maximum length for string parameters.
	MaxLength *int

	// Pattern is a regex pattern that the parameter value must match.
	// Only applicable for string parameters.
	Pattern string

	// Enum is a list of allowed values for the parameter.
	// The parameter value must be one of these values.
	Enum []interface{}

	// Format specifies an extended format for the parameter type.
	// Common formats: "int32", "int64", "float", "double", "date", "date-time", "email", "uuid", "byte", "binary"
	Format string
}

// NewParamOptions creates a new ParamOptions with the provided options.
// This is a convenience function for creating parameter options.
func NewParamOptions() *ParamOptions {
	return &ParamOptions{}
}

// WithDefault sets the default value for the parameter.
func (p *ParamOptions) WithDefault(value interface{}) *ParamOptions {
	p.Default = value
	return p
}

// WithExample sets an example value for the parameter.
func (p *ParamOptions) WithExample(value interface{}) *ParamOptions {
	p.Example = value
	return p
}

// WithCollectionFormat sets the collection format for array parameters.
func (p *ParamOptions) WithCollectionFormat(format string) *ParamOptions {
	p.CollectionFormat = format
	return p
}

// WithAllowEmptyValue sets whether the parameter allows empty values.
func (p *ParamOptions) WithAllowEmptyValue(allow bool) *ParamOptions {
	p.AllowEmptyValue = allow
	return p
}

// WithMinimum sets the minimum value for numeric parameters.
func (p *ParamOptions) WithMinimum(min float64) *ParamOptions {
	p.Minimum = &min
	return p
}

// WithMaximum sets the maximum value for numeric parameters.
func (p *ParamOptions) WithMaximum(max float64) *ParamOptions {
	p.Maximum = &max
	return p
}

// WithMinLength sets the minimum length for string parameters.
func (p *ParamOptions) WithMinLength(min int) *ParamOptions {
	p.MinLength = &min
	return p
}

// WithMaxLength sets the maximum length for string parameters.
func (p *ParamOptions) WithMaxLength(max int) *ParamOptions {
	p.MaxLength = &max
	return p
}

// WithPattern sets the regex pattern for string parameters.
func (p *ParamOptions) WithPattern(pattern string) *ParamOptions {
	p.Pattern = pattern
	return p
}

// WithEnum sets the allowed enum values for the parameter.
func (p *ParamOptions) WithEnum(values ...interface{}) *ParamOptions {
	p.Enum = values
	return p
}

// WithFormat sets the extended format for the parameter type.
func (p *ParamOptions) WithFormat(format string) *ParamOptions {
	p.Format = format
	return p
}
