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
	Security              []string
}

type Group struct {
	GroupName string
	Routes    []Route
	Groups    []Group
}

func GenerateSwagger(routes []Route, groups []Group, defaultResponses []models.ReturnType) {
	var (
		packagesToImport = make(map[string]bool)
		fullFileContent  = &strings.Builder{}
		wrapperStructs   = &strings.Builder{} // Store wrapper structs with descriptions
	)

	log.Printf("Generating %s file...", fileName)

	routes, groups = addDefaultResponses(routes, groups, defaultResponses)

	if routes != nil {
		writeRoutes("", routes, fullFileContent, packagesToImport, wrapperStructs)
	}

	if groups != nil {
		writeGroup(groups, fullFileContent, packagesToImport, wrapperStructs)
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

func writeRoutes(groupName string, routes []Route, s *strings.Builder, packagesToImport map[string]bool, wrapperStructs *strings.Builder) {
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
			s.WriteString(fmt.Sprintf("// @Param %s path %s %t \"%s\"\n",
				param.Name, param.ParamType, param.Required, param.Description),
			)
		}

		for _, param := range r.QueryParams {
			s.WriteString(fmt.Sprintf("// @Param %s query %s %t \"%s\"\n",
				param.Name, param.ParamType, param.Required, param.Description),
			)
		}

		for _, param := range r.HeaderParams {
			s.WriteString(fmt.Sprintf("// @Param %s header %s %t \"%s\"\n",
				param.Name, param.ParamType, param.Required, param.Description),
			)
		}

		if len(r.Security) > 0 {
			for _, scheme := range r.Security {
				if strings.TrimSpace(scheme) == "" { // skip empty
					continue
				}
				s.WriteString(fmt.Sprintf("// @Security %s\n", scheme))
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

		if data.Body == nil {
			s.WriteString(fmt.Sprintf("// %s %d\n", respType, data.StatusCode))
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
			s.WriteString(fmt.Sprintf("// %s %d {object} %s", respType, data.StatusCode, structName))
		}

		addPackageToImport(data, packagesToImport)
		handleOverrideStructFields(s, data)

		s.WriteString("\n")
	}
}

func writeGroup(groups []Group, s *strings.Builder, packagesToImport map[string]bool, wrapperStructs *strings.Builder) {
	for _, g := range groups {
		writeRoutes(g.GroupName, g.Routes, s, packagesToImport, wrapperStructs)

		if g.Groups != nil {
			writeGroup(g.Groups, s, packagesToImport, wrapperStructs)
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
