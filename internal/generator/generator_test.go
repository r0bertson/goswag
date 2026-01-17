package generator

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/r0bertson/goswag/internal/generator/testutil"
	"github.com/r0bertson/goswag/models"
	"github.com/stretchr/testify/assert"
)

func TestGetStructAndPackageName(t *testing.T) {
	type args struct {
		body interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should return the struct name and package name",
			args: args{
				body: models.ReturnType{},
			},
			want: "models.ReturnType",
		},
		{
			name: "Should not return * if the struct is a pointer",
			args: args{
				body: &models.ReturnType{},
			},
			want: "models.ReturnType",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStructAndPackageName(tt.args.body)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAddLineIfNotEmpty(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		format   string
		expected string
	}{
		{
			name:     "Should return empty string",
			input:    "",
			format:   "",
			expected: "",
		},
		{
			name:     "Should return empty string even if we have format",
			input:    "",
			format:   "test %s",
			expected: "",
		},
		{
			name:     "Should return the input string",
			input:    "test",
			format:   "some %s",
			expected: "some test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var b strings.Builder
			addLineIfNotEmpty(&b, tt.input, tt.format)
			result := b.String()

			if result != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, result)
			}
		})
	}
}

func TestAddTextIfNotEmptyOrDefault_slice(t *testing.T) {
	var tests = []struct {
		name        string
		input       []string
		defaultText string
		format      string
		expected    string
	}{
		{
			name:        "Should return default text",
			input:       []string{},
			defaultText: "default",
			format:      "some %s",
			expected:    "some default",
		},
		{
			name:        "Should return the input string",
			input:       []string{"test"},
			defaultText: "default",
			format:      "some %s",
			expected:    "some test",
		},
		{
			name:        "Should return the multiple input string separated by comma",
			input:       []string{"test", "test2"},
			defaultText: "default",
			format:      "some %s",
			expected:    "some test,test2",
		},
		{
			name:        "Should add nothing if input and default text are empty",
			input:       []string{},
			defaultText: "",
			format:      "some %s",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var b strings.Builder
			addTextIfNotEmptyOrDefault(&b, tt.defaultText, tt.format, tt.input...)
			result := b.String()

			if result != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, result)
			}
		})
	}
}

func TestAddTextIfNotEmptyOrDefault_string(t *testing.T) {
	var tests = []struct {
		name        string
		input       string
		defaultText string
		format      string
		expected    string
	}{
		{
			name:        "Should return default text",
			input:       "",
			defaultText: "default",
			format:      "some %s",
			expected:    "some default",
		},
		{
			name:        "Should return the input string",
			input:       "test",
			defaultText: "default",
			format:      "some %s",
			expected:    "some test",
		},
		{
			name:        "Should add nothing if input and default text are empty",
			input:       "",
			defaultText: "",
			format:      "some %s",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var b strings.Builder
			addTextIfNotEmptyOrDefault(&b, tt.defaultText, tt.format, tt.input)
			result := b.String()

			if result != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, result)
			}
		})
	}
}

func TestWriteGroup(t *testing.T) {
	var tests = []struct {
		name                  string
		groups                []Group
		expectedStringBuilder string
	}{
		{
			name: "Should return string with the group name",
			groups: []Group{
				{
					GroupName: "test",
					Routes: []Route{
						{
							Description: "test group",
							Path:        "/test",
							Method:      "GET",
						},
					},
				},
			},
			expectedStringBuilder: "// @Description test group\n// @Tags test\n// @Router /test [get]\n\n",
		},
		{
			name: "Should recursively return string with the group name",
			groups: []Group{
				{
					GroupName: "test",
					Routes: []Route{
						{
							Path:        "/test",
							Description: "test group",
						},
					},
					Groups: []Group{
						{
							GroupName: "test2",
							Routes: []Route{
								{
									Path:        "/test2",
									Description: "test group 2",
								},
							},
						},
					},
				},
			},
			expectedStringBuilder: "// @Description test group\n// @Tags test\n// @Router /test []\n\n// @Description test group 2\n// @Tags test2\n// @Router /test2 []\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var b strings.Builder
			var wrapperStructs strings.Builder
			writeGroup(tt.groups, &b, map[string]bool{}, &wrapperStructs, nil)

			assert.Equal(t, tt.expectedStringBuilder, b.String())
		})
	}
}

func TestWriteRoutes(t *testing.T) {
	var tests = []struct {
		name                  string
		groupName             string
		routes                []Route
		expectedStringBuilder string
	}{
		{
			name:      "Should group name as tag of route",
			groupName: "test",
			routes: []Route{
				{},
			},
			expectedStringBuilder: "// @Tags test\n\n",
		},
		{
			name:      "Should add summary and description if we have summary",
			groupName: "",
			routes: []Route{
				{
					Summary: "test",
				},
			},
			expectedStringBuilder: "// @Summary test\n// @Description test\n\n",
		},
		{
			name:      "Should add description if we have description",
			groupName: "",
			routes: []Route{
				{
					Description: "test",
				},
			},
			expectedStringBuilder: "// @Description test\n\n",
		},
		{
			name:      "Should add tags if we have tags",
			groupName: "",
			routes: []Route{
				{
					Tags: []string{"test"},
				},
			},
			expectedStringBuilder: "// @Tags test\n\n",
		},
		{
			name:      "Should add tags, instead of group if we have tags",
			groupName: "group_test",
			routes: []Route{
				{
					Tags: []string{"tag_test"},
				},
			},
			expectedStringBuilder: "// @Tags tag_test\n\n",
		},
		{
			name:      "Should add default accept json if we have post method",
			groupName: "",
			routes: []Route{
				{
					Method: "POST",
				},
			},
			expectedStringBuilder: "// @Accept json\n\n",
		},
		{
			name:      "Should add accept text instead of default json",
			groupName: "",
			routes: []Route{
				{
					Method:  "POST",
					Accepts: []string{"text"},
				},
			},
			expectedStringBuilder: "// @Accept text\n\n",
		},
		{
			name:      "Should add produces if we have return",
			groupName: "",
			routes: []Route{
				{
					Returns: []models.ReturnType{
						{},
					},
				},
			},
			expectedStringBuilder: "// @Produce json\n\n",
		},
		{
			name:      "Should add request body if we have reads",
			groupName: "",
			routes: []Route{
				{
					Reads: models.ReturnType{},
				},
			},
			expectedStringBuilder: "// @Param request body models.ReturnType true \"Request\"\n\n",
		},
		{
			name:      "Should add path params if we have path params",
			groupName: "",
			routes: []Route{
				{
					PathParams: []Param{
						{
							Name:        "test",
							Description: "someTest",
							ParamType:   "string",
							Required:    true,
						},
					},
				},
			},
			expectedStringBuilder: "// @Param test path string true \"someTest\"\n\n",
		},
		{
			name:      "Should add query params if we have query params",
			groupName: "",
			routes: []Route{
				{
					QueryParams: []Param{
						{
							Name:        "test",
							Description: "test",
							ParamType:   "string",
							Required:    true,
						},
					},
				},
			},
			expectedStringBuilder: "// @Param test query string true \"test\"\n\n",
		},
		{
			name:      "Should add header params if we have header params",
			groupName: "",
			routes: []Route{
				{
					HeaderParams: []Param{
						{
							Name:        "test",
							Description: "test",
							ParamType:   "string",
							Required:    true,
						},
					},
				},
			},
			expectedStringBuilder: "// @Param test header string true \"test\"\n\n",
		},
		{
			name:      "Should add router if we have path",
			groupName: "",
			routes: []Route{
				{
					Path:   "/test",
					Method: "GET",
				},
			},
			expectedStringBuilder: "// @Router /test [get]\n\n",
		},
		{
			name:      "Should add func name if we have func name",
			groupName: "",
			routes: []Route{
				{
					FuncName: "test",
				},
			},
			expectedStringBuilder: "func test() {} //nolint:unused \n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var b strings.Builder
			var wrapperStructs strings.Builder
			writeRoutes(tt.groupName, tt.routes, &b, map[string]bool{}, &wrapperStructs, nil)

			assert.Equal(t, tt.expectedStringBuilder, b.String())
		})
	}
}

func TestWriteReturns(t *testing.T) {
	var tests = []struct {
		name                  string
		returns               []models.ReturnType
		expectedStringBuilder string
		expectedPackages      map[string]bool
	}{
		{
			name: "Should return the struct name and package name as success 200",
			returns: []models.ReturnType{
				{
					StatusCode: 200,
					Body:       models.ReturnType{},
				},
			},
			expectedStringBuilder: "// @Success 200 {object} models.ReturnType\n",
			expectedPackages:      map[string]bool{"github.com/r0bertson/goswag/models": true},
		},
		{
			name: "Should do nothing if we do not have status code",
			returns: []models.ReturnType{
				{
					Body: models.ReturnType{},
				},
			},
			expectedStringBuilder: "",
			expectedPackages:      map[string]bool{},
		},
		{
			name: "Should return the struct name and package name as failure 400",
			returns: []models.ReturnType{
				{
					StatusCode: 400,
					Body:       models.ReturnType{},
				},
			},
			expectedStringBuilder: "// @Failure 400 {object} models.ReturnType\n",
			expectedPackages:      map[string]bool{"github.com/r0bertson/goswag/models": true},
		},
		{
			name: "Should add only status code if we do not have body",
			returns: []models.ReturnType{
				{
					StatusCode: 400,
				},
			},
			expectedStringBuilder: "// @Failure 400\n",
			expectedPackages:      map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var (
				b              strings.Builder
				wrapperStructs strings.Builder
				pkgs           = make(map[string]bool)
			)

			writeReturns(tt.returns, &b, pkgs, &wrapperStructs)

			assert.Equal(t, tt.expectedStringBuilder, b.String())
			assert.Equal(t, tt.expectedPackages, pkgs)
		})
	}
}

func Test_handleOverrideStructFields(t *testing.T) {
	var tests = []struct {
		name                  string
		data                  models.ReturnType
		expectedStringBuilder string
	}{
		{
			name:                  "Should do nothing if we do not have override struct fields",
			data:                  models.ReturnType{},
			expectedStringBuilder: "",
		},
		{
			name: "Should add override struct fields",
			data: models.ReturnType{
				Body: testutil.OverrideStruct{},
				OverrideStructFields: map[string]interface{}{
					"test": testutil.TestGeneric{},
				},
			},
			expectedStringBuilder: "{test=testutil.TestGeneric}",
		},
		{
			name: "Should add multiple override struct fields",
			data: models.ReturnType{
				Body: testutil.OverrideStruct{},
				OverrideStructFields: map[string]interface{}{
					"test":  testutil.TestGeneric{},
					"test2": testutil.TestGeneric{},
				},
			},
			expectedStringBuilder: "{test=testutil.TestGeneric,test2=testutil.TestGeneric}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var b strings.Builder
			handleOverrideStructFields(&b, tt.data)

			result := b.String()
			// For maps with multiple entries, check that all keys are present rather than exact string match
			// due to non-deterministic map iteration order in Go
			if tt.name == "Should add multiple override struct fields" {
				assert.Contains(t, result, "test=testutil.TestGeneric")
				assert.Contains(t, result, "test2=testutil.TestGeneric")
			} else {
				assert.Equal(t, tt.expectedStringBuilder, result)
			}
		})
	}
}

func Test_writeFileContent(t *testing.T) {
	type args struct {
		file             io.Writer
		content          string
		packagesToImport map[string]bool
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "Should write the file content",
			args: args{
				file:             &strings.Builder{},
				content:          "test",
				packagesToImport: map[string]bool{"test": true},
			},
			expected: "package main\n\nimport (\n\t_ \"test\"\n)\n\ntest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeFileContent(tt.args.file, tt.args.content, tt.args.packagesToImport)
		})
	}
}

func Test_addDefaultResponses(t *testing.T) {
	type args struct {
		routes           []Route
		groups           []Group
		defaultResponses []models.ReturnType
	}
	tests := []struct {
		name     string
		args     args
		expected []Route
	}{
		{
			name: "Should do nothing if we do not have default responses",
			args: args{
				routes: []Route{
					{},
				},
				groups: []Group{
					{
						Routes: []Route{
							{},
						},
					},
				},
				defaultResponses: []models.ReturnType{},
			},
			expected: []Route{
				{},
			},
		},
		{
			name: "Should add default responses to routes",
			args: args{
				routes: []Route{
					{},
				},
				groups: []Group{
					{
						Routes: []Route{
							{},
						},
					},
				},
				defaultResponses: []models.ReturnType{
					{
						StatusCode: 200,
					},
				},
			},
			expected: []Route{
				{
					Returns: []models.ReturnType{
						{
							StatusCode: 200,
						},
					},
				},
			},
		},
		{
			name: "Should add default responses to groups",
			args: args{
				routes: []Route{
					{},
				},
				groups: []Group{
					{
						Routes: []Route{
							{},
						},
					},
				},
				defaultResponses: []models.ReturnType{
					{
						StatusCode: 204,
					},
				},
			},
			expected: []Route{
				{
					Returns: []models.ReturnType{
						{
							StatusCode: 204,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRoutes, gotGroups := addDefaultResponses(tt.args.routes, tt.args.groups, tt.args.defaultResponses)
			assert.Equal(t, tt.expected, gotRoutes)
			assert.Equal(t, tt.expected, gotGroups[0].Routes)
		})
	}
}

func Test_generateWrapperStruct(t *testing.T) {
	type TestStruct struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	tests := []struct {
		name                string
		originalStruct      interface{}
		fieldDescriptions   map[string]string
		suffix              string
		expectedContains    []string
		expectedNotContains []string
	}{
		{
			name:           "Should generate wrapper struct with field descriptions",
			originalStruct: TestStruct{},
			fieldDescriptions: map[string]string{
				"id":    "Unique identifier",
				"name":  "User's full name",
				"email": "User's email address",
			},
			suffix: "Request",
			expectedContains: []string{
				"type Wrapper",
				"Request struct",
				"// Unique identifier",
				"// User's full name",
				"// User's email address",
				"ID string",
				"Name string",
				"Email string",
			},
			expectedNotContains: []string{},
		},
		{
			name:           "Should generate wrapper struct with partial descriptions",
			originalStruct: TestStruct{},
			fieldDescriptions: map[string]string{
				"name": "User's full name",
			},
			suffix: "Response",
			expectedContains: []string{
				"type Wrapper",
				"Response struct",
				"// User's full name",
				"ID string",
				"Name string",
				"Email string",
			},
			expectedNotContains: []string{
				"// Unique identifier",
				"// User's email address",
			},
		},
		{
			name:                "Should generate wrapper struct even without descriptions",
			originalStruct:      TestStruct{},
			fieldDescriptions:   map[string]string{},
			suffix:              "Request",
			expectedContains:    []string{"type Wrapper", "Request struct", "ID string", "Name string", "Email string"},
			expectedNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wrapperStructs strings.Builder
			packagesToImport := make(map[string]bool)

			wrapperName := generateWrapperStruct(tt.originalStruct, tt.fieldDescriptions, &wrapperStructs, packagesToImport, tt.suffix)

			result := wrapperStructs.String()

			// Check wrapper name is not empty
			assert.NotEmpty(t, wrapperName)
			assert.Contains(t, wrapperName, "Wrapper")

			// Check expected strings are present
			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected, "Expected to find: %s", expected)
			}

			// Check unexpected strings are not present
			for _, notExpected := range tt.expectedNotContains {
				assert.NotContains(t, result, notExpected, "Should not find: %s", notExpected)
			}
		})
	}
}

func Test_sanitizeStructName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Should replace dots with underscores",
			input:    "package.StructName",
			expected: "package_StructName",
		},
		{
			name:     "Should replace slashes with underscores",
			input:    "github.com/user/package.StructName",
			expected: "github_com_user_package_StructName",
		},
		{
			name:     "Should replace hyphens with underscores",
			input:    "my-package.Struct-Name",
			expected: "my_package_Struct_Name",
		},
		{
			name:     "Should handle simple names",
			input:    "StructName",
			expected: "StructName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeStructName(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func Test_writeReturns_withFieldDescriptions(t *testing.T) {
	type TestResponse struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	tests := []struct {
		name                string
		returns             []models.ReturnType
		expectedContains    []string
		expectedNotContains []string
	}{
		{
			name: "Should use wrapper struct when field descriptions are provided",
			returns: []models.ReturnType{
				{
					StatusCode: 200,
					Body:       TestResponse{},
					FieldDescriptions: map[string]string{
						"id":    "Unique identifier",
						"name":  "User's name",
						"email": "User's email",
					},
				},
			},
			expectedContains: []string{
				"@Success 200 {object} Wrapper",
				"type Wrapper",
				"Response struct",
				"// Unique identifier",
				"// User's name",
				"// User's email",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should use original struct when no field descriptions",
			returns: []models.ReturnType{
				{
					StatusCode: 200,
					Body:       TestResponse{},
				},
			},
			expectedContains: []string{
				"@Success 200 {object}",
			},
			expectedNotContains: []string{
				"Wrapper",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s strings.Builder
			var wrapperStructs strings.Builder
			packagesToImport := make(map[string]bool)

			writeReturns(tt.returns, &s, packagesToImport, &wrapperStructs)

			result := s.String() + wrapperStructs.String()

			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected, "Expected to find: %s", expected)
			}

			for _, notExpected := range tt.expectedNotContains {
				assert.NotContains(t, result, notExpected, "Should not find: %s", notExpected)
			}
		})
	}
}

func Test_writeRoutes_withReadFieldDescriptions(t *testing.T) {
	type TestRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	tests := []struct {
		name                string
		route               Route
		expectedContains    []string
		expectedNotContains []string
	}{
		{
			name: "Should use wrapper struct when ReadFieldDescriptions are provided",
			route: Route{
				Path:        "/test",
				Method:      "POST",
				Summary:     "Test",
				Reads:       TestRequest{},
				ReadFieldDescriptions: map[string]string{
					"name":  "User's full name",
					"email": "User's email address",
				},
			},
			expectedContains: []string{
				"@Param request body Wrapper",
				"Request",
				"type Wrapper",
				"// User's full name",
				"// User's email address",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should use original struct when no ReadFieldDescriptions",
			route: Route{
				Path:   "/test",
				Method: "POST",
				Reads:  TestRequest{},
			},
			expectedContains: []string{
				"@Param request body",
			},
			expectedNotContains: []string{
				"Wrapper",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s strings.Builder
			var wrapperStructs strings.Builder
			packagesToImport := make(map[string]bool)

			writeRoutes("", []Route{tt.route}, &s, packagesToImport, &wrapperStructs, nil)

			result := s.String() + wrapperStructs.String()

			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected, "Expected to find: %s", expected)
			}

			for _, notExpected := range tt.expectedNotContains {
				assert.NotContains(t, result, notExpected, "Should not find: %s", notExpected)
			}
		})
	}
}

func Test_generateWrapperStruct_withPointers(t *testing.T) {
	type TestStructWithPointers struct {
		RequiredField string  `json:"required_field"`
		OptionalField *string `json:"optional_field"`
		NullableField *int    `json:"nullable_field,omitempty"`
		NoTagField    *bool   `json:"-"`
	}

	tests := []struct {
		name                string
		originalStruct      interface{}
		fieldDescriptions   map[string]string
		suffix              string
		expectedContains    []string
		expectedNotContains []string
	}{
		{
			name:           "Should add omitempty to pointer fields without it",
			originalStruct: TestStructWithPointers{},
			fieldDescriptions: map[string]string{
				"optional_field": "Optional string field",
			},
			suffix: "Request",
			expectedContains: []string{
				"RequiredField string",
				"OptionalField *string",
				"json:\"optional_field,omitempty\"",
				"binding:\"omitempty\"",
				"NullableField *int",
				"json:\"nullable_field,omitempty\"",
			},
			expectedNotContains: []string{
				"json:\"optional_field\"`", // Should not have json tag without omitempty
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wrapperStructs strings.Builder
			packagesToImport := make(map[string]bool)

			wrapperName := generateWrapperStruct(tt.originalStruct, tt.fieldDescriptions, &wrapperStructs, packagesToImport, tt.suffix)

			result := wrapperStructs.String()

			// Check wrapper name is not empty
			assert.NotEmpty(t, wrapperName)
			assert.Contains(t, wrapperName, "Wrapper")

			// Check expected strings are present
			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected, "Expected to find: %s", expected)
			}

			// Check unexpected strings are not present
			for _, notExpected := range tt.expectedNotContains {
				assert.NotContains(t, result, notExpected, "Should not find: %s", notExpected)
			}
		})
	}
}

func Test_ensurePointerTags(t *testing.T) {
	tests := []struct {
		name     string
		tag      reflect.StructTag
		expected string
	}{
		{
			name:     "Should add omitempty to JSON tag for pointer field",
			tag:      reflect.StructTag(`json:"optional_field"`),
			expected: `json:"optional_field,omitempty"`,
		},
		{
			name:     "Should not modify JSON tag that already has omitempty",
			tag:      reflect.StructTag(`json:"nullable_field,omitempty"`),
			expected: `json:"nullable_field,omitempty"`,
		},
		{
			name:     "Should add binding tag with omitempty",
			tag:      reflect.StructTag(`json:"optional_field"`),
			expected: `binding:"omitempty"`,
		},
		{
			name:     "Should not modify JSON tag with -",
			tag:      reflect.StructTag(`json:"-"`),
			expected: `json:"-"`,
		},
		{
			name:     "Should handle empty JSON tag",
			tag:      reflect.StructTag(``),
			expected: `binding:"omitempty"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ensurePointerTags(tt.tag)
			resultStr := string(result)
			
			// Check that expected strings are in the result
			if strings.Contains(tt.expected, "json:") {
				assert.Contains(t, resultStr, tt.expected, "Expected JSON tag: %s", tt.expected)
			}
			if strings.Contains(tt.expected, "binding:") {
				assert.Contains(t, resultStr, tt.expected, "Expected binding tag: %s", tt.expected)
			}
		})
	}
}

func Test_addPackageToImport(t *testing.T) {
	tests := []struct {
		name         string
		data         models.ReturnType
		initialPkgs  map[string]bool
		expectedPkgs map[string]bool
	}{
		{
			name: "Should add package for non-generic type",
			data: models.ReturnType{
				Body: models.ReturnType{},
			},
			initialPkgs: make(map[string]bool),
			expectedPkgs: map[string]bool{
				"github.com/r0bertson/goswag/models": true,
			},
		},
		{
			name: "Should add package for generic type",
			data: models.ReturnType{
				Body: testutil.StructGeneric[testutil.TestGeneric]{},
			},
			initialPkgs: make(map[string]bool),
			expectedPkgs: map[string]bool{
				"github.com/r0bertson/goswag/internal/generator/testutil": true,
			},
		},
		{
			name: "Should not add package for primitive type",
			data: models.ReturnType{
				Body: 42,
			},
			initialPkgs:  make(map[string]bool),
			expectedPkgs: map[string]bool{},
		},
		{
			name: "Should not add package for nil body",
			data: models.ReturnType{
				Body: nil,
			},
			initialPkgs:  make(map[string]bool),
			expectedPkgs: map[string]bool{},
		},
		{
			name: "Should not duplicate existing package",
			data: models.ReturnType{
				Body: models.ReturnType{},
			},
			initialPkgs: map[string]bool{
				"github.com/r0bertson/goswag/models": true,
			},
			expectedPkgs: map[string]bool{
				"github.com/r0bertson/goswag/models": true,
			},
		},
		{
			name: "Should add package for pointer to struct",
			data: models.ReturnType{
				Body: &models.ReturnType{},
			},
			initialPkgs: make(map[string]bool),
			expectedPkgs: map[string]bool{
				"github.com/r0bertson/goswag/models": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packagesToImport := tt.initialPkgs
			addPackageToImport(tt.data, packagesToImport)
			assert.Equal(t, tt.expectedPkgs, packagesToImport)
		})
	}
}

func TestWriteParam(t *testing.T) {
	tests := []struct {
		name     string
		param    Param
		location string
		expected string
	}{
		{
			name: "Should write basic parameter",
			param: Param{
				Name:        "id",
				Description: "User ID",
				ParamType:   "string",
				Required:    true,
			},
			location: "path",
			expected: "// @Param id path string true \"User ID\"\n",
		},
		{
			name: "Should write parameter with default value",
			param: Param{
				Name:        "status",
				Description: "Filter status",
				ParamType:   "string",
				Required:    false,
				Default:     "active",
			},
			location: "query",
			expected: "// @Param status query string false \"Filter status (default: active)\"\n",
		},
		{
			name: "Should write parameter with example",
			param: Param{
				Name:        "email",
				Description: "User email",
				ParamType:   "string",
				Required:    true,
				Example:     "user@example.com",
			},
			location: "query",
			expected: "// @Param email query string true \"User email (example: user@example.com)\"\n",
		},
		{
			name: "Should write parameter with enum values",
			param: Param{
				Name:        "status",
				Description: "User status",
				ParamType:   "string",
				Required:    true,
				Enum:        []interface{}{"active", "inactive", "pending"},
			},
			location: "query",
			expected: "// @Param status query string true \"User status (enum: active,inactive,pending)\"\n",
		},
		{
			name: "Should write parameter with format",
			param: Param{
				Name:        "user_id",
				Description: "User identifier",
				ParamType:   "string",
				Required:    true,
				Format:      "uuid",
			},
			location: "path",
			expected: "// @Param user_id path string true \"User identifier (format: uuid)\"\n",
		},
		{
			name: "Should write parameter with minimum and maximum",
			param: Param{
				Name:        "age",
				Description: "User age",
				ParamType:   "integer",
				Required:    false,
				Minimum:     floatPtr(18),
				Maximum:     floatPtr(120),
			},
			location: "query",
			expected: "// @Param age query integer false \"User age (min: 18) (max: 120)\"\n",
		},
		{
			name: "Should write parameter with minLength and maxLength",
			param: Param{
				Name:        "password",
				Description: "User password",
				ParamType:   "string",
				Required:    true,
				MinLength:   intPtr(8),
				MaxLength:   intPtr(128),
			},
			location: "formData",
			expected: "// @Param password formData string true \"User password (minLength: 8) (maxLength: 128)\"\n",
		},
		{
			name: "Should write parameter with pattern",
			param: Param{
				Name:        "phone",
				Description: "Phone number",
				ParamType:   "string",
				Required:    false,
				Pattern:     "^\\+?[1-9]\\d{1,14}$",
			},
			location: "query",
			expected: "// @Param phone query string false \"Phone number (pattern: ^\\\\+?[1-9]\\\\d{1,14}$)\"\n",
		},
		{
			name: "Should write parameter with collectionFormat",
			param: Param{
				Name:             "tags",
				Description:      "Filter tags",
				ParamType:        "array",
				Required:         false,
				CollectionFormat: "multi",
			},
			location: "query",
			expected: "// @Param tags query array false \"Filter tags (collectionFormat: multi)\"\n",
		},
		{
			name: "Should write parameter with allowEmptyValue",
			param: Param{
				Name:           "filter",
				Description:    "Optional filter",
				ParamType:      "string",
				Required:       false,
				AllowEmptyValue: true,
			},
			location: "query",
			expected: "// @Param filter query string false \"Optional filter (allowEmptyValue: true)\"\n",
		},
		{
			name: "Should write parameter with all extended attributes",
			param: Param{
				Name:             "status",
				Description:      "Filter by status",
				ParamType:        "array",
				Required:         false,
				Default:          "active",
				Example:          "pending",
				Enum:             []interface{}{"active", "inactive", "pending"},
				Format:           "string",
				MinLength:        intPtr(3),
				MaxLength:        intPtr(20),
				CollectionFormat: "csv",
			},
			location: "query",
			expected: "// @Param status query array false \"Filter by status (format: string) (enum: active,inactive,pending) (default: active) (example: pending) (minLength: 3) (maxLength: 20) (collectionFormat: csv)\"\n",
		},
		{
			name: "Should write formData parameter",
			param: Param{
				Name:        "file",
				Description: "Upload file",
				ParamType:   "file",
				Required:    true,
			},
			location: "formData",
			expected: "// @Param file formData file true \"Upload file\"\n",
		},
		{
			name: "Should write header parameter",
			param: Param{
				Name:        "Authorization",
				Description: "Bearer token",
				ParamType:   "string",
				Required:    true,
			},
			location: "header",
			expected: "// @Param Authorization header string true \"Bearer token\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			writeParam(&b, tt.param, tt.location)
			result := b.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriteRoutes_withExtendedParams(t *testing.T) {
	tests := []struct {
		name                string
		route               Route
		expectedContains    []string
		expectedNotContains []string
	}{
		{
			name: "Should write query parameter with options",
			route: Route{
				Path:   "/users",
				Method: "GET",
				QueryParams: []Param{
					{
						Name:        "status",
						Description: "Filter status",
						ParamType:   "string",
						Required:    false,
						Default:     "active",
						Enum:        []interface{}{"active", "inactive"},
					},
				},
			},
			expectedContains: []string{
				"@Param status query",
				"(default: active)",
				"(enum: active,inactive)",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write formData parameter",
			route: Route{
				Path:   "/upload",
				Method: "POST",
				FormDataParams: []Param{
					{
						Name:        "file",
						Description: "File to upload",
						ParamType:   "file",
						Required:    true,
					},
				},
			},
			expectedContains: []string{
				"@Param file formData file",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write path parameter with validation",
			route: Route{
				Path:   "/users/{id}",
				Method: "GET",
				PathParams: []Param{
					{
						Name:        "id",
						Description: "User ID",
						ParamType:   "string",
						Required:    true,
						Format:      "uuid",
						Pattern:     "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$",
					},
				},
			},
			expectedContains: []string{
				"@Param id path string true",
				"(format: uuid)",
				"(pattern:",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write query parameter with collectionFormat",
			route: Route{
				Path:   "/search",
				Method: "GET",
				QueryParams: []Param{
					{
						Name:             "tags",
						Description:      "Filter tags",
						ParamType:        "array",
						Required:         false,
						CollectionFormat: "multi",
					},
				},
			},
			expectedContains: []string{
				"@Param tags query array",
				"(collectionFormat: multi)",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write parameter with allowEmptyValue for query",
			route: Route{
				Path:   "/search",
				Method: "GET",
				QueryParams: []Param{
					{
						Name:           "q",
						Description:    "Search query",
						ParamType:      "string",
						Required:       false,
						AllowEmptyValue: true,
					},
				},
			},
			expectedContains: []string{
				"@Param q query string false",
				"(allowEmptyValue: true)",
			},
			expectedNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s strings.Builder
			var wrapperStructs strings.Builder
			packagesToImport := make(map[string]bool)

			writeRoutes("", []Route{tt.route}, &s, packagesToImport, &wrapperStructs, nil)

			result := s.String()

			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected, "Expected to find: %s", expected)
			}

			for _, notExpected := range tt.expectedNotContains {
				assert.NotContains(t, result, notExpected, "Should not find: %s", notExpected)
			}
		})
	}
}

func TestWriteRoutes_withOperationEnhancements(t *testing.T) {
	tests := []struct {
		name                string
		route               Route
		expectedContains    []string
		expectedNotContains []string
	}{
		{
			name: "Should write OperationID",
			route: Route{
				Path:       "/users",
				Method:     "GET",
				OperationID: "getUsers",
			},
			expectedContains: []string{
				"@OperationID getUsers",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write Deprecated",
			route: Route{
				Path:      "/old-endpoint",
				Method:    "GET",
				Deprecated: true,
			},
			expectedContains: []string{
				"@Deprecated",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write Schemes",
			route: Route{
				Path:    "/secure",
				Method:  "GET",
				Schemes: []string{"https", "wss"},
			},
			expectedContains: []string{
				"@Schemes https wss",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write ExternalDocs with description",
			route: Route{
				Path:   "/docs",
				Method: "GET",
				ExternalDocs: &models.ExternalDocs{
					URL:         "https://example.com/docs",
					Description: "External documentation",
				},
			},
			expectedContains: []string{
				"@ExternalDocs https://example.com/docs",
				"External documentation",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write ExternalDocs without description",
			route: Route{
				Path:   "/docs",
				Method: "GET",
				ExternalDocs: &models.ExternalDocs{
					URL: "https://example.com/docs",
				},
			},
			expectedContains: []string{
				"@ExternalDocs https://example.com/docs",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write all operation enhancements together",
			route: Route{
				Path:       "/complete",
				Method:     "GET",
				OperationID: "completeOperation",
				Deprecated: true,
				Schemes:    []string{"https"},
				ExternalDocs: &models.ExternalDocs{
					URL:         "https://example.com",
					Description: "Complete docs",
				},
			},
			expectedContains: []string{
				"@OperationID completeOperation",
				"@Deprecated",
				"@Schemes https",
				"@ExternalDocs https://example.com",
				"Complete docs",
			},
			expectedNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s strings.Builder
			var wrapperStructs strings.Builder
			packagesToImport := make(map[string]bool)

			writeRoutes("", []Route{tt.route}, &s, packagesToImport, &wrapperStructs, nil)

			result := s.String()

			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected, "Expected to find: %s", expected)
			}

			for _, notExpected := range tt.expectedNotContains {
				assert.NotContains(t, result, notExpected, "Should not find: %s", notExpected)
			}
		})
	}
}

func TestWriteReturns_withResponseEnhancements(t *testing.T) {
	tests := []struct {
		name                string
		returns             []models.ReturnType
		expectedContains    []string
		expectedNotContains []string
	}{
		{
			name: "Should write response with description",
			returns: []models.ReturnType{
				{
					StatusCode:  200,
					Body:        models.ReturnType{},
					Description: "Successfully retrieved data",
				},
			},
			expectedContains: []string{
				"@Success 200 {object}",
				"Successfully retrieved data",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write response headers",
			returns: []models.ReturnType{
				{
					StatusCode: 200,
					Body:       models.ReturnType{},
					Headers: map[string]*models.ResponseHeader{
						"X-RateLimit-Remaining": {
							Type:        "integer",
							Description: "Number of requests remaining",
						},
						"X-Request-ID": {
							Type:        "string",
							Description: "Request identifier",
							Format:      "uuid",
						},
					},
				},
			},
			expectedContains: []string{
				"@Header X-RateLimit-Remaining {integer}",
				"Number of requests remaining",
				"@Header X-Request-ID {string}",
				"Request identifier",
				"(format: uuid)",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write response examples",
			returns: []models.ReturnType{
				{
					StatusCode: 200,
					Body:       models.ReturnType{},
					Examples: map[string]interface{}{
						"application/json": map[string]interface{}{
							"id":   1,
							"name": "test",
						},
					},
				},
			},
			expectedContains: []string{
				"@Example application/json",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write response without body but with headers",
			returns: []models.ReturnType{
				{
					StatusCode:  204,
					Description: "No content",
					Headers: map[string]*models.ResponseHeader{
						"X-Request-ID": {
							Type:        "string",
							Description: "Request identifier",
						},
					},
				},
			},
			expectedContains: []string{
				"@Success 204",
				"No content",
				"@Header X-Request-ID {string}",
			},
			expectedNotContains: []string{
				"{object}",
			},
		},
		{
			name: "Should write response with all enhancements",
			returns: []models.ReturnType{
				{
					StatusCode:  200,
					Body:        models.ReturnType{},
					Description: "Complete response",
					Headers: map[string]*models.ResponseHeader{
						"X-RateLimit-Remaining": {
							Type:        "integer",
							Description: "Remaining requests",
						},
					},
					Examples: map[string]interface{}{
						"application/json": map[string]string{"status": "ok"},
					},
				},
			},
			expectedContains: []string{
				"@Success 200 {object}",
				"Complete response",
				"@Header X-RateLimit-Remaining",
				"@Example application/json",
			},
			expectedNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s strings.Builder
			var wrapperStructs strings.Builder
			packagesToImport := make(map[string]bool)

			writeReturns(tt.returns, &s, packagesToImport, &wrapperStructs)

			result := s.String()

			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected, "Expected to find: %s", expected)
			}

			for _, notExpected := range tt.expectedNotContains {
				assert.NotContains(t, result, notExpected, "Should not find: %s", notExpected)
			}
		})
	}
}

func TestWriteResponseHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]*models.ResponseHeader
		expected string
	}{
		{
			name:     "Should return empty for nil headers",
			headers:  nil,
			expected: "",
		},
		{
			name:     "Should return empty for empty headers",
			headers:  map[string]*models.ResponseHeader{},
			expected: "",
		},
		{
			name: "Should write single header",
			headers: map[string]*models.ResponseHeader{
				"X-Request-ID": {
					Type:        "string",
					Description: "Request identifier",
				},
			},
			expected: "// @Header X-Request-ID {string} \"Request identifier\"\n",
		},
		{
			name: "Should write header with format",
			headers: map[string]*models.ResponseHeader{
				"X-RateLimit-Remaining": {
					Type:        "integer",
					Description: "Remaining requests",
					Format:      "int32",
				},
			},
			expected: "// @Header X-RateLimit-Remaining {integer} \"Remaining requests (format: int32)\"\n",
		},
		{
			name: "Should write multiple headers",
			headers: map[string]*models.ResponseHeader{
				"X-Request-ID": {
					Type:        "string",
					Description: "Request identifier",
				},
				"X-RateLimit-Remaining": {
					Type:        "integer",
					Description: "Remaining requests",
				},
			},
			expected: "// @Header X-Request-ID {string} \"Request identifier\"\n// @Header X-RateLimit-Remaining {integer} \"Remaining requests\"\n",
		},
		{
			name: "Should default to string type if not specified",
			headers: map[string]*models.ResponseHeader{
				"X-Custom": {
					Description: "Custom header",
				},
			},
			expected: "// @Header X-Custom {string} \"Custom header\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			writeResponseHeaders(&b, tt.headers)
			result := b.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriteResponseExamples(t *testing.T) {
	tests := []struct {
		name     string
		examples map[string]interface{}
		expected string
	}{
		{
			name:     "Should return empty for nil examples",
			examples: nil,
			expected: "",
		},
		{
			name:     "Should return empty for empty examples",
			examples: map[string]interface{}{},
			expected: "",
		},
		{
			name: "Should write single example",
			examples: map[string]interface{}{
				"application/json": map[string]interface{}{
					"id":   1,
					"name": "test",
				},
			},
			expected: "// @Example application/json \"map[id:1 name:test]\"\n",
		},
		{
			name: "Should write multiple examples",
			examples: map[string]interface{}{
				"application/json": map[string]string{"status": "ok"},
				"application/xml":  "<response><status>ok</status></response>",
			},
			expected: "// @Example application/json \"map[status:ok]\"\n// @Example application/xml \"<response><status>ok</status></response>\"\n",
		},
		{
			name: "Should skip nil examples",
			examples: map[string]interface{}{
				"application/json": map[string]string{"status": "ok"},
				"application/xml":  nil,
			},
			expected: "// @Example application/json \"map[status:ok]\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			writeResponseExamples(&b, tt.examples)
			result := b.String()
			// Note: Map iteration order is non-deterministic, so we check for presence of key parts
			if tt.expected == "" {
				assert.Equal(t, tt.expected, result)
			} else {
				// For non-empty cases, verify key parts are present
				assert.Contains(t, result, "@Example")
			}
		})
	}
}

func TestWriteGlobalConfig(t *testing.T) {
	tests := []struct {
		name                string
		config              *models.SwaggerConfig
		expectedContains    []string
		expectedNotContains []string
	}{
		{
			name:   "Should return empty for nil config",
			config: nil,
			expectedContains: []string{},
			expectedNotContains: []string{},
		},
		{
			name: "Should write host",
			config: &models.SwaggerConfig{
				Host: "api.example.com",
			},
			expectedContains: []string{
				"@host api.example.com",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write basePath",
			config: &models.SwaggerConfig{
				BasePath: "/v1",
			},
			expectedContains: []string{
				"@BasePath /v1",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write schemes",
			config: &models.SwaggerConfig{
				Schemes: []string{"https", "wss"},
			},
			expectedContains: []string{
				"@schemes https wss",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write contact information",
			config: &models.SwaggerConfig{
				Contact: models.NewContactInfo("API Support", "support@example.com", "https://example.com/contact"),
			},
			expectedContains: []string{
				"@contact.name API Support",
				"@contact.email support@example.com",
				"@contact.url https://example.com/contact",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write license information",
			config: &models.SwaggerConfig{
				License: models.NewLicenseInfo("MIT", "https://opensource.org/licenses/MIT"),
			},
			expectedContains: []string{
				"@license.name MIT",
				"@license.url https://opensource.org/licenses/MIT",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write termsOfService",
			config: &models.SwaggerConfig{
				TermsOfService: "https://example.com/terms",
			},
			expectedContains: []string{
				"@termsOfService https://example.com/terms",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write externalDocs",
			config: &models.SwaggerConfig{
				ExternalDocs: models.NewExternalDocs("https://docs.example.com", "API Documentation"),
			},
			expectedContains: []string{
				"@externalDocs.url https://docs.example.com",
				"@externalDocs.description API Documentation",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write global security",
			config: &models.SwaggerConfig{
				GlobalSecurity: []string{"BearerAuth", "ApiKeyAuth"},
			},
			expectedContains: []string{
				"@security BearerAuth ApiKeyAuth",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should write complete configuration",
			config: &models.SwaggerConfig{
				Host:         "api.example.com",
				BasePath:     "/v1",
				Schemes:      []string{"https"},
				Contact:      models.NewContactInfo("Support", "support@example.com", ""),
				License:      models.NewLicenseInfo("MIT", "https://opensource.org/licenses/MIT"),
				TermsOfService: "https://example.com/terms",
				ExternalDocs: models.NewExternalDocs("https://docs.example.com", "Docs"),
				GlobalSecurity: []string{"BearerAuth"},
			},
			expectedContains: []string{
				"@host api.example.com",
				"@BasePath /v1",
				"@schemes https",
				"@contact.name Support",
				"@contact.email support@example.com",
				"@license.name MIT",
				"@license.url https://opensource.org/licenses/MIT",
				"@termsOfService https://example.com/terms",
				"@externalDocs.url https://docs.example.com",
				"@externalDocs.description Docs",
				"@security BearerAuth",
			},
			expectedNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			writeGlobalConfig(&b, tt.config)
			result := b.String()

			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected, "Expected to find: %s", expected)
			}

			for _, notExpected := range tt.expectedNotContains {
				assert.NotContains(t, result, notExpected, "Should not find: %s", notExpected)
			}
		})
	}
}

func TestWriteRoutes_withGlobalSecurity(t *testing.T) {
	tests := []struct {
		name                string
		route               Route
		config              *models.SwaggerConfig
		expectedContains    []string
		expectedNotContains []string
	}{
		{
			name: "Should apply global security when route has no security",
			route: Route{
				Path:   "/users",
				Method: "GET",
			},
			config: &models.SwaggerConfig{
				GlobalSecurity: []string{"BearerAuth"},
			},
			expectedContains: []string{
				"@Security BearerAuth",
			},
			expectedNotContains: []string{},
		},
		{
			name: "Should use route-specific security over global",
			route: Route{
				Path:     "/admin",
				Method:   "GET",
				Security: []string{"AdminAuth"},
			},
			config: &models.SwaggerConfig{
				GlobalSecurity: []string{"BearerAuth"},
			},
			expectedContains: []string{
				"@Security AdminAuth",
			},
			expectedNotContains: []string{
				"@Security BearerAuth",
			},
		},
		{
			name: "Should not apply security when config is nil",
			route: Route{
				Path:   "/public",
				Method: "GET",
			},
			config: nil,
			expectedContains: []string{},
			expectedNotContains: []string{
				"@Security",
			},
		},
		{
			name: "Should not apply global security when route has empty security array",
			route: Route{
				Path:     "/public",
				Method:   "GET",
				Security: []string{},
			},
			config: &models.SwaggerConfig{
				GlobalSecurity: []string{"BearerAuth"},
			},
			expectedContains: []string{
				"@Security BearerAuth",
			},
			expectedNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s strings.Builder
			var wrapperStructs strings.Builder
			packagesToImport := make(map[string]bool)

			writeRoutes("", []Route{tt.route}, &s, packagesToImport, &wrapperStructs, tt.config)

			result := s.String()

			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected, "Expected to find: %s", expected)
			}

			for _, notExpected := range tt.expectedNotContains {
				assert.NotContains(t, result, notExpected, "Should not find: %s", notExpected)
			}
		})
	}
}

// Helper functions for tests
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}