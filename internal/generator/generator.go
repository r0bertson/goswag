package generator

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/r0bertson/goswag/models"
)

const fileName = "goswag.go"

type Param struct {
	Name        string
	Description string
	ParamType   string
	Required    bool
	// Extended fields for Swagger 2.0 support
	Default         interface{}
	Example         interface{}
	CollectionFormat string
	AllowEmptyValue  bool
	Minimum         *float64
	Maximum         *float64
	MinLength       *int
	MaxLength       *int
	Pattern         string
	Enum            []interface{}
	Format          string
}

type Route struct {
	Path        string
	Method      string
	FuncName    string // it will be used to generate the function on the goswag.go file
	Summary     string
	Description string
	Tags        []string
	Accepts     []string
	Produces    []string
	Reads       interface{}
	// ReadFieldDescriptions is used to add descriptions to struct fields in the request body.
	// The key should be the JSON field name (e.g., "id", "name", "email").
	ReadFieldDescriptions map[string]string
	Returns               []models.ReturnType // example: map[statusCode]responseBody
	QueryParams           []Param
	HeaderParams          []Param
	PathParams            []Param
	FormDataParams        []Param // For multipart/form-data and application/x-www-form-urlencoded
	Security              []string
	// Operation-level enhancements
	OperationID string
	Deprecated  bool
	Schemes     []string
	ExternalDocs *models.ExternalDocs
}

type Group struct {
	GroupName string
	Routes    []Route
	Groups    []Group
}

func GenerateSwagger(routes []Route, groups []Group, defaultResponses []models.ReturnType, config *models.SwaggerConfig) {
	var (
		packagesToImport = make(map[string]bool)
		fullFileContent  = &strings.Builder{}
		wrapperStructs   = &strings.Builder{} // Store wrapper structs with descriptions
	)

	log.Printf("Generating %s file...", fileName)

	routes, groups = addDefaultResponses(routes, groups, defaultResponses)

	// Write global swagger configuration if provided
	if config != nil {
		writeGlobalConfig(fullFileContent, config)
	}

	if routes != nil {
		writeRoutes("", routes, fullFileContent, packagesToImport, wrapperStructs, config)
	}

	if groups != nil {
		writeGroup(groups, fullFileContent, packagesToImport, wrapperStructs, config)
	}

	f, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Write wrapper structs first, then the rest of the content
	writeFileContent(f, wrapperStructs.String()+fullFileContent.String(), packagesToImport)

	log.Printf("%s file generated successfully!", fileName)
}

// addDefaultResponses adds the default responses to the routes and groups if it are not empty
func addDefaultResponses(routes []Route, groups []Group, defaultResponses []models.ReturnType) ([]Route, []Group) {
	if len(defaultResponses) == 0 {
		return routes, groups
	}

	for i := range routes {
		routes[i].Returns = append(routes[i].Returns, defaultResponses...)
	}

	for i := range groups {
		groups[i].Routes, groups[i].Groups = addDefaultResponses(groups[i].Routes, groups[i].Groups, defaultResponses)
	}

	return routes, groups
}

func writeFileContent(file io.Writer, content string, packagesToImport map[string]bool) {
	fmt.Fprintf(file, "package main\n\n")

	if len(packagesToImport) > 0 {
		fmt.Fprintf(file, "import (\n")

		for pkg := range packagesToImport {
			fmt.Fprintf(file, "\t_ \"%s\"\n", pkg)
		}

		fmt.Fprintf(file, ")\n\n")
	}

	fmt.Fprintf(file, "%s", content)
}

func writeRoutes(groupName string, routes []Route, s *strings.Builder, packagesToImport map[string]bool, wrapperStructs *strings.Builder, config *models.SwaggerConfig) {
	for _, r := range routes {
		addLineIfNotEmpty(s, r.Summary, "// @Summary %s\n")
		addTextIfNotEmptyOrDefault(s, r.Summary, "// @Description %s\n", r.Description)

		if len(r.Tags) > 0 {
			s.WriteString(fmt.Sprintf("// @Tags %s\n", strings.Join(r.Tags, ",")))
		} else if groupName != "" {
			s.WriteString(fmt.Sprintf("// @Tags %s\n", groupName))
		}

		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			// methods like get or delete do not have a request body
			addTextIfNotEmptyOrDefault(s, "json", "// @Accept %s\n", r.Accepts...)
		}

		if r.Returns != nil {
			// only add the produces if there is a return
			addTextIfNotEmptyOrDefault(s, "json", "// @Produce %s\n", r.Produces...)
		}

		if r.Reads != nil {
			structName := getStructAndPackageName(r.Reads)
			// If field descriptions are provided, generate a wrapper struct
			if len(r.ReadFieldDescriptions) > 0 {
				wrapperName := generateWrapperStruct(r.Reads, r.ReadFieldDescriptions, wrapperStructs, packagesToImport, "Request")
				structName = wrapperName
			}
			s.WriteString(fmt.Sprintf("// @Param request body %s true \"Request\"\n", structName))
		}

		for _, param := range r.PathParams {
			writeParam(s, param, "path")
		}

		for _, param := range r.QueryParams {
			writeParam(s, param, "query")
		}

		for _, param := range r.HeaderParams {
			writeParam(s, param, "header")
		}

		for _, param := range r.FormDataParams {
			writeParam(s, param, "formData")
		}

		// Apply global security if no operation-specific security is defined
		securitySchemes := r.Security
		if len(securitySchemes) == 0 && config != nil && len(config.GlobalSecurity) > 0 {
			securitySchemes = config.GlobalSecurity
		}

		if len(securitySchemes) > 0 {
			for _, scheme := range securitySchemes {
				if strings.TrimSpace(scheme) == "" { // skip empty
					continue
				}
				s.WriteString(fmt.Sprintf("// @Security %s\n", scheme))
			}
		}

		// Operation-level enhancements
		if r.OperationID != "" {
			s.WriteString(fmt.Sprintf("// @OperationID %s\n", r.OperationID))
		}

		if r.Deprecated {
			s.WriteString("// @Deprecated\n")
		}

		if len(r.Schemes) > 0 {
			s.WriteString(fmt.Sprintf("// @Schemes %s\n", strings.Join(r.Schemes, " ")))
		}

		if r.ExternalDocs != nil {
			if r.ExternalDocs.Description != "" {
				s.WriteString(fmt.Sprintf("// @ExternalDocs %s \"%s\"\n", r.ExternalDocs.URL, r.ExternalDocs.Description))
			} else {
				s.WriteString(fmt.Sprintf("// @ExternalDocs %s\n", r.ExternalDocs.URL))
			}
		}

		if r.Returns != nil {
			writeReturns(r.Returns, s, packagesToImport, wrapperStructs)
		}

		if r.Path != "" {
			s.WriteString(fmt.Sprintf("// @Router %s [%s]\n", r.Path, strings.ToLower(r.Method)))
		}

		if r.FuncName != "" {
			s.WriteString(fmt.Sprintf("func %s() {} //nolint:unused \n", r.FuncName))
		}

		s.WriteString("\n")
	}
}

func writeReturns(returns []models.ReturnType, s *strings.Builder, packagesToImport map[string]bool, wrapperStructs *strings.Builder) {
	for _, data := range returns {
		if data.StatusCode == 0 {
			continue
		}

		respType := "@Success"
		firstDigit := data.StatusCode / 100

		if firstDigit != http.StatusOK/100 { // <> 2xx
			respType = "@Failure"
		}

		// Build response description if provided
		responseDesc := ""
		if data.Description != "" {
			responseDesc = " \"" + data.Description + "\""
		}

		if data.Body == nil {
			// Response without body - can still have headers and description
			s.WriteString(fmt.Sprintf("// %s %d%s", respType, data.StatusCode, responseDesc))
			writeResponseHeaders(s, data.Headers)
			writeResponseExamples(s, data.Examples)
			s.WriteString("\n")
			continue
		}

		var isGeneric bool = writeIfIsGenericType(s, data, respType)

		structName := getStructAndPackageName(data.Body)
		// If field descriptions are provided, generate a wrapper struct
		if len(data.FieldDescriptions) > 0 && !isGeneric {
			wrapperName := generateWrapperStruct(data.Body, data.FieldDescriptions, wrapperStructs, packagesToImport, "Response")
			structName = wrapperName
		}

		if !isGeneric {
			// if it is not a generic type, we can write the response normally
			s.WriteString(fmt.Sprintf("// %s %d {object} %s%s", respType, data.StatusCode, structName, responseDesc))
		}

		addPackageToImport(data, packagesToImport)
		handleOverrideStructFields(s, data)

		// Write response headers if any
		writeResponseHeaders(s, data.Headers)

		// Write response examples if any
		writeResponseExamples(s, data.Examples)

		s.WriteString("\n")
	}
}

func writeGroup(groups []Group, s *strings.Builder, packagesToImport map[string]bool, wrapperStructs *strings.Builder, config *models.SwaggerConfig) {
	for _, g := range groups {
		writeRoutes(g.GroupName, g.Routes, s, packagesToImport, wrapperStructs, config)

		if g.Groups != nil {
			writeGroup(g.Groups, s, packagesToImport, wrapperStructs, config)
		}
	}
}

// addPackageToImport adds the package to import.
func addPackageToImport(data models.ReturnType, packagesToImport map[string]bool) {
	if data.Body == nil {
		return
	}
	t := reflect.TypeOf(data.Body)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.PkgPath() != "" {
		packagesToImport[t.PkgPath()] = true
	}
}

// writeIfIsGenericType writes the correctly response type if it is a generic type
// and returns the packages to import that need to be added to the goswag.go file to make it work
func writeIfIsGenericType(s *strings.Builder, data models.ReturnType, respType string) (isGeneric bool) {
	bodyName := getStructAndPackageName(data.Body)

	// generic last character here will be ']'
	// testutil.StructGeneric[testutil.TestGeneric]
	isGeneric = bodyName[len(bodyName)-1:] == "]"
	if !isGeneric {
		return
	}

	isArray := strings.Contains(bodyName, "[[]")
	hasSlash := strings.Contains(bodyName, "/")

	if isArray && hasSlash {
		// example: testutil.StructGeneric[[]github.com/r0bertson/goswag/internal/generator/testutil.TestGeneric]

		bodyRemovedLastChar := bodyName[:len(bodyName)-1] // testutil.StructGeneric[[]github.com/r0bertson/goswag/internal/generator/testutil.TestGeneric

		// get the last text after '/'
		str := strings.Split(bodyRemovedLastChar, "/")
		insideGenericsFullName := str[len(str)-1] // testutil.TestGeneric

		insidePkg := strings.Split(bodyRemovedLastChar, "[[]")[1]                 // github.com/r0bertson/goswag/internal/generator/testutil.TestGeneric
		removedType := strings.Replace(insidePkg, insideGenericsFullName, "", -1) // github.com/r0bertson/goswag/internal/generator/

		correctlyResponseType := strings.Replace(bodyName, removedType, "", -1) // remove full package from the struct name

		s.WriteString(fmt.Sprintf("// %s %d {object} %s", respType, data.StatusCode, correctlyResponseType))

		return isGeneric
	}

	if hasSlash {
		// example: testutil.StructGeneric[github.com/r0bertson/goswag/internal/generator/testutil.TestGeneric]

		bodyRemovedLastChar := bodyName[:len(bodyName)-1] // testutil.StructGeneric[github.com/r0bertson/goswag/internal/generator/testutil.TestGeneric

		// get the last text after '/'
		str := strings.Split(bodyRemovedLastChar, "/")
		insideGenericsFullName := str[len(str)-1] // testutil.TestGeneric

		insidePkg := strings.Split(bodyRemovedLastChar, "[")[1]                   // github.com/r0bertson/goswag/internal/generator/testutil.TestGeneric
		removedType := strings.Replace(insidePkg, insideGenericsFullName, "", -1) // github.com/r0bertson/goswag/internal/generator/

		correctlyResponseType := strings.Replace(bodyName, removedType, "", -1) // remove full package from the struct name

		s.WriteString(fmt.Sprintf("// %s %d {object} %s", respType, data.StatusCode, correctlyResponseType))

		return isGeneric
	}

	// example: genericStruct[int] or genericStruct[string] or genericStruct[bool]
	// primitive types do not need to import packages

	s.WriteString(fmt.Sprintf("// %s %d {object} %s", respType, data.StatusCode, bodyName))

	return isGeneric
}

func handleOverrideStructFields(s *strings.Builder, data models.ReturnType) {
	if data.OverrideStructFields != nil {
		i := 0
		for key, object := range data.OverrideStructFields {
			if i == 0 {
				s.WriteString("{")
			}

			s.WriteString(fmt.Sprintf("%s=%s", key, getStructAndPackageName(object)))
			if i == len(data.OverrideStructFields)-1 {
				s.WriteString("}")
			} else {
				s.WriteString(",")
			}
			i++
		}
	}
}

// generateWrapperStruct generates a wrapper struct with field descriptions as comments.
// It returns the name of the generated wrapper struct.
func generateWrapperStruct(originalStruct interface{}, fieldDescriptions map[string]string, wrapperStructs *strings.Builder, packagesToImport map[string]bool, suffix string) string {
	t := reflect.TypeOf(originalStruct)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Generate a unique wrapper struct name
	originalName := t.Name()
	if t.PkgPath() != "" {
		// Extract package name from full path
		parts := strings.Split(t.PkgPath(), "/")
		pkgName := parts[len(parts)-1]
		originalName = pkgName + "." + originalName
	}

	// Create a unique wrapper name
	wrapperName := fmt.Sprintf("Wrapper%s%s", sanitizeStructName(originalName), suffix)

	// Check if we've already generated this wrapper (avoid duplicates)
	// For now, we'll generate it each time - could optimize later with a map

	// Write the wrapper struct definition
	wrapperStructs.WriteString(fmt.Sprintf("// %s is a wrapper struct with field descriptions\n", wrapperName))
	wrapperStructs.WriteString(fmt.Sprintf("type %s struct {\n", wrapperName))

	// Iterate through struct fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		jsonName := field.Name
		isPointer := field.Type.Kind() == reflect.Ptr

		// Extract JSON field name from tag
		if jsonTag != "" && jsonTag != "-" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				jsonName = parts[0]
			}
		}

		// Add description comment if available
		if desc, ok := fieldDescriptions[jsonName]; ok {
			wrapperStructs.WriteString(fmt.Sprintf("\t// %s\n", desc))
		}

		// Handle pointer fields - ensure they're marked as optional/nullable
		fieldType := field.Type.String()
		updatedTag := field.Tag

		if isPointer {
			// For pointer fields, ensure omitempty is in JSON tag and add binding tag for swag
			// swaggo/swag should automatically detect pointer fields as optional/nullable
			// The omitempty tag and pointer type together should be sufficient
			updatedTag = ensurePointerTags(field.Tag)
		}

		// Write field definition
		wrapperStructs.WriteString(fmt.Sprintf("\t%s %s `%s`\n", field.Name, fieldType, updatedTag))
	}

	wrapperStructs.WriteString("}\n\n")

	// Add package to imports
	if t.PkgPath() != "" {
		packagesToImport[t.PkgPath()] = true
	}

	return wrapperName
}

// ensurePointerTags ensures pointer fields have proper tags for Swagger to recognize them as optional/nullable
func ensurePointerTags(tag reflect.StructTag) reflect.StructTag {
	jsonTag := tag.Get("json")
	bindingTag := tag.Get("binding")

	// Parse JSON tag
	jsonParts := strings.Split(jsonTag, ",")
	hasOmitempty := false
	for _, part := range jsonParts {
		if strings.TrimSpace(part) == "omitempty" {
			hasOmitempty = true
			break
		}
	}

	// Add omitempty to JSON tag if not present and tag is not empty
	if !hasOmitempty && jsonTag != "" && jsonTag != "-" {
		if len(jsonParts) == 1 && jsonParts[0] == jsonTag {
			jsonTag = jsonTag + ",omitempty"
		} else {
			jsonTag = strings.Join(append(jsonParts, "omitempty"), ",")
		}
	}

	// Ensure binding tag has omitempty for swag
	if bindingTag == "" {
		bindingTag = "omitempty"
	} else if !strings.Contains(bindingTag, "omitempty") {
		bindingTag = bindingTag + ",omitempty"
	}

	// Reconstruct the tag string
	tagStr := string(tag)

	// Replace json tag if we modified it
	if jsonTag != tag.Get("json") {
		// Extract the tag name (e.g., `json:"..."`)
		if strings.Contains(tagStr, `json:"`) {
			// Replace the json tag value
			start := strings.Index(tagStr, `json:"`)
			end := strings.Index(tagStr[start+6:], `"`)
			if end != -1 {
				end = start + 6 + end + 1
				oldJsonTag := tagStr[start:end]
				newJsonTag := fmt.Sprintf(`json:"%s"`, jsonTag)
				tagStr = strings.Replace(tagStr, oldJsonTag, newJsonTag, 1)
			}
		}
	}

	// Add or update binding tag
	if strings.Contains(tagStr, `binding:"`) {
		// Replace existing binding tag
		start := strings.Index(tagStr, `binding:"`)
		end := strings.Index(tagStr[start+9:], `"`)
		if end != -1 {
			end = start + 9 + end + 1
			oldBindingTag := tagStr[start:end]
			newBindingTag := fmt.Sprintf(`binding:"%s"`, bindingTag)
			tagStr = strings.Replace(tagStr, oldBindingTag, newBindingTag, 1)
		}
	} else {
		// Add binding tag
		if tagStr != "" && !strings.HasSuffix(tagStr, "`") {
			tagStr = tagStr + " "
		}
		tagStr = tagStr + fmt.Sprintf(`binding:"%s"`, bindingTag)
	}

	return reflect.StructTag(tagStr)
}

// sanitizeStructName removes special characters to create a valid Go identifier
func sanitizeStructName(name string) string {
	// Replace dots and other invalid characters with underscores
	result := strings.ReplaceAll(name, ".", "_")
	result = strings.ReplaceAll(result, "-", "_")
	result = strings.ReplaceAll(result, "/", "_")
	return result
}

func getStructAndPackageName(body any) string {
	isPointer := reflect.TypeOf(body).Kind() == reflect.Ptr
	if isPointer {
		body = reflect.ValueOf(body).Elem().Interface()
	}

	return reflect.TypeOf(body).String()
}

func addTextIfNotEmptyOrDefault(s *strings.Builder, defaultText, format string, text ...string) {
	if text != nil {
		if len(text) >= 1 && strings.TrimSpace(text[0]) != "" {
			s.WriteString(fmt.Sprintf(format, strings.Join(text, ",")))
			return
		}
	}

	if defaultText != "" {
		s.WriteString(fmt.Sprintf(format, defaultText))
	}
}

func addLineIfNotEmpty(s *strings.Builder, data, format string) {
	if data != "" {
		s.WriteString(fmt.Sprintf(format, data))
	}
}

// writeResponseHeaders writes response header annotations.
// Swaggo/swag format: @Header name {type} "description"
func writeResponseHeaders(s *strings.Builder, headers map[string]*models.ResponseHeader) {
	if headers == nil || len(headers) == 0 {
		return
	}

	for name, header := range headers {
		if header == nil {
			continue
		}
		headerType := header.Type
		if headerType == "" {
			headerType = "string" // default type
		}
		desc := header.Description
		if header.Format != "" {
			desc = fmt.Sprintf("%s (format: %s)", desc, header.Format)
		}
		s.WriteString(fmt.Sprintf("// @Header %s {%s} %q\n", name, headerType, desc))
	}
}

// writeResponseExamples writes response example annotations.
// Swaggo/swag format: @Example {content-type} {example}
func writeResponseExamples(s *strings.Builder, examples map[string]interface{}) {
	if examples == nil || len(examples) == 0 {
		return
	}

	for contentType, example := range examples {
		if example == nil {
			continue
		}
		// Convert example to string representation
		exampleStr := fmt.Sprintf("%v", example)
		// Escape quotes if needed
		exampleStr = strings.ReplaceAll(exampleStr, "\"", "\\\"")
		s.WriteString(fmt.Sprintf("// @Example %s %q\n", contentType, exampleStr))
	}
}

// writeParam writes a parameter annotation with all extended Swagger 2.0 features.
// Swaggo/swag supports extended attributes through the description field and struct tags.
// We format the description to include additional metadata that swag can parse.
func writeParam(s *strings.Builder, param Param, location string) {
	// Base @Param annotation format: @Param name location type required "description"
	// Swaggo/swag will parse this and generate the swagger.json with proper structure
	
	// Build description with extended attributes
	desc := param.Description
	
	// Add format if specified
	if param.Format != "" {
		desc = fmt.Sprintf("%s (format: %s)", desc, param.Format)
	}
	
	// Add enum values if specified
	if len(param.Enum) > 0 {
		enumStrs := make([]string, 0, len(param.Enum))
		for _, e := range param.Enum {
			enumStrs = append(enumStrs, fmt.Sprintf("%v", e))
		}
		desc = fmt.Sprintf("%s (enum: %s)", desc, strings.Join(enumStrs, ","))
	}
	
	// Add default value if specified
	if param.Default != nil {
		desc = fmt.Sprintf("%s (default: %v)", desc, param.Default)
	}
	
	// Add example if specified
	if param.Example != nil {
		desc = fmt.Sprintf("%s (example: %v)", desc, param.Example)
	}
	
	// Add validation constraints
	if param.Minimum != nil {
		desc = fmt.Sprintf("%s (min: %v)", desc, *param.Minimum)
	}
	if param.Maximum != nil {
		desc = fmt.Sprintf("%s (max: %v)", desc, *param.Maximum)
	}
	if param.MinLength != nil {
		desc = fmt.Sprintf("%s (minLength: %d)", desc, *param.MinLength)
	}
	if param.MaxLength != nil {
		desc = fmt.Sprintf("%s (maxLength: %d)", desc, *param.MaxLength)
	}
	if param.Pattern != "" {
		desc = fmt.Sprintf("%s (pattern: %s)", desc, param.Pattern)
	}
	
	// Add collectionFormat for arrays (important for query/formData parameters)
	paramType := param.ParamType
	if param.CollectionFormat != "" && strings.HasPrefix(paramType, "array") {
		// Note: swaggo/swag may need the collectionFormat in a specific way
		// For now, we'll include it in the description
		desc = fmt.Sprintf("%s (collectionFormat: %s)", desc, param.CollectionFormat)
	}
	
	// Add allowEmptyValue for query/formData
	if param.AllowEmptyValue && (location == "query" || location == "formData") {
		desc = fmt.Sprintf("%s (allowEmptyValue: true)", desc)
	}
	
	// Write the @Param annotation
	// Format: @Param name location type required "description"
	s.WriteString(fmt.Sprintf("// @Param %s %s %s %t %q\n",
		param.Name, location, paramType, param.Required, desc))
}

// writeGlobalConfig writes global Swagger configuration annotations.
// These annotations are typically placed at the top of the file or in main.go.
// Swaggo/swag reads these from the main package, so we'll generate them as comments
// that can be copied to main.go or included in the generated file.
func writeGlobalConfig(s *strings.Builder, config *models.SwaggerConfig) {
	if config == nil {
		return
	}

	// Note: @title, @version are typically set in main.go, but we can document them here
	// @host, @BasePath, @schemes are global settings
	if config.Host != "" {
		s.WriteString(fmt.Sprintf("// @host %s\n", config.Host))
	}

	if config.BasePath != "" {
		s.WriteString(fmt.Sprintf("// @BasePath %s\n", config.BasePath))
	}

	if len(config.Schemes) > 0 {
		s.WriteString(fmt.Sprintf("// @schemes %s\n", strings.Join(config.Schemes, " ")))
	}

	// Contact information
	if config.Contact != nil {
		if config.Contact.Name != "" {
			s.WriteString(fmt.Sprintf("// @contact.name %s\n", config.Contact.Name))
		}
		if config.Contact.Email != "" {
			s.WriteString(fmt.Sprintf("// @contact.email %s\n", config.Contact.Email))
		}
		if config.Contact.URL != "" {
			s.WriteString(fmt.Sprintf("// @contact.url %s\n", config.Contact.URL))
		}
	}

	// License information
	if config.License != nil {
		if config.License.Name != "" {
			s.WriteString(fmt.Sprintf("// @license.name %s\n", config.License.Name))
		}
		if config.License.URL != "" {
			s.WriteString(fmt.Sprintf("// @license.url %s\n", config.License.URL))
		}
	}

	// Terms of Service
	if config.TermsOfService != "" {
		s.WriteString(fmt.Sprintf("// @termsOfService %s\n", config.TermsOfService))
	}

	// External Documentation
	if config.ExternalDocs != nil {
		if config.ExternalDocs.Description != "" {
			s.WriteString(fmt.Sprintf("// @externalDocs.description %s\n", config.ExternalDocs.Description))
		}
		if config.ExternalDocs.URL != "" {
			s.WriteString(fmt.Sprintf("// @externalDocs.url %s\n", config.ExternalDocs.URL))
		}
	}

	// Global Security (applied to all operations unless overridden)
	// Note: This is written as a comment for documentation, actual application happens in writeRoutes
	if len(config.GlobalSecurity) > 0 {
		s.WriteString(fmt.Sprintf("// @security %s\n", strings.Join(config.GlobalSecurity, " ")))
	}

	s.WriteString("\n")
}
