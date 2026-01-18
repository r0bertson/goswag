# LLM Guide: Generating Swagger Documentation with GoSwag

This guide is designed for LLM agents to automatically generate Swagger documentation using GoSwag from unannotated Go HTTP handlers.

## Task Overview

Your goal is to analyze Go HTTP handler code and generate equivalent GoSwag code that will produce complete Swagger 2.0 documentation. The handlers have **no existing annotations** - you must infer all Swagger metadata from the handler code itself.

## Analysis Process

### Step 1: Identify the Handler Structure

Analyze the handler function to extract:
- **HTTP Method**: GET, POST, PUT, DELETE, PATCH, etc.
- **Route Path**: Extract from route registration or handler signature
- **Handler Function**: The actual handler implementation

### Step 2: Extract Request Information

From the handler code, identify:

1. **Path Parameters**: Variables in the URL path (e.g., `/users/{id}`)
   - Type: Usually `string` or `int`
   - Required: Always `true` for path params
   - Format: Check if UUID, integer, etc.

2. **Query Parameters**: From `r.URL.Query()` or framework-specific methods
   - Name: Query key name
   - Type: Infer from usage (`strconv.Atoi` → `int`, `strconv.ParseBool` → `bool`)
   - Required: Check if there's validation/error handling for missing values
   - Default: Check if there's a default value assignment
   - Validation: Look for min/max checks, enum values, regex patterns

3. **Header Parameters**: From `r.Header.Get()` or framework methods
   - Common: Authorization, Content-Type, X-Request-ID, etc.
   - Type: Usually `string`
   - Required: Check validation logic

4. **Request Body**: From JSON unmarshaling or form parsing
   - Type: The struct being unmarshaled
   - Content-Type: Usually `application/json` or `multipart/form-data`
   - Field Descriptions: Infer from field names and validation tags

5. **File Uploads**: Look for `multipart.File` or `multipart.FileHeader`
   - Parameter type: `file`
   - Required: Check validation

### Step 3: Extract Response Information

From the handler code, identify:

1. **Status Codes**: All `http.StatusXXX` or framework status codes used
2. **Response Bodies**: Structs being marshaled to JSON
3. **Response Headers**: Headers being set (e.g., `w.Header().Set("Location", ...)`)
4. **Error Responses**: Common error structures returned

### Step 4: Infer Metadata

From code patterns, infer:

- **Summary**: From function name or first comment
- **Description**: From function comments or logic description
- **Tags**: From route path or handler grouping
- **OperationID**: From function name (camelCase)
- **Deprecated**: Look for deprecation comments or legacy route patterns
- **Security**: Look for authentication middleware or token validation
- **Examples**: Can be inferred from test data or default values

## Code Patterns to Recognize

### Pattern 1: Basic GET Handler

```go
// Handler Code
func GetUser(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    // ... fetch user logic
    json.NewEncoder(w).Encode(user)
}

// Generated GoSwag Code
gh.GET("/users/{id}").
    Summary("Get user").
    Description("Retrieve a user by ID").
    Tags("users").
    OperationID("getUser").
    PathParam("id", "User ID", goswag.StringType, true).
    Returns([]models.ReturnType{
        {StatusCode: http.StatusOK, Body: User{}},
    })
```

### Pattern 2: Query Parameters with Validation

```go
// Handler Code
func ListUsers(w http.ResponseWriter, r *http.Request) {
    pageStr := r.URL.Query().Get("page")
    page := 1
    if pageStr != "" {
        p, _ := strconv.Atoi(pageStr)
        if p > 0 {
            page = p
        }
    }
    
    limitStr := r.URL.Query().Get("limit")
    limit := 10
    if limitStr != "" {
        l, _ := strconv.Atoi(limitStr)
        if l > 0 && l <= 100 {
            limit = l
        }
    }
    // ... fetch logic
}

// Generated GoSwag Code
gh.GET("/users").
    Summary("List users").
    Tags("users").
    QueryParamWithOptions("page", "Page number", goswag.IntType, false,
        models.NewParamOptions().
            WithDefault(1).
            WithMinimum(1),
    ).
    QueryParamWithOptions("limit", "Items per page", goswag.IntType, false,
        models.NewParamOptions().
            WithDefault(10).
            WithMinimum(1).
            WithMaximum(100),
    ).
    Returns([]models.ReturnType{
        {StatusCode: http.StatusOK, Body: UserListResponse{}},
    })
```

### Pattern 3: POST with Request Body

```go
// Handler Code
func CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Validation
    if req.Name == "" || len(req.Name) < 2 {
        http.Error(w, "Name required", http.StatusBadRequest)
        return
    }
    
    // ... create logic
    w.Header().Set("Location", fmt.Sprintf("/users/%d", user.ID))
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

// Generated GoSwag Code
gh.POST("/users").
    Summary("Create user").
    Description("Create a new user account").
    Tags("users").
    OperationID("createUser").
    Accepts("application/json").
    Read(CreateUserRequest{}).
    ReadFieldDescriptions(map[string]string{
        "name": "User's full name (required, min 2 characters)",
    }).
    Returns([]models.ReturnType{
        {
            StatusCode: http.StatusCreated,
            Body:       User{},
            Headers: map[string]*models.ResponseHeader{
                "Location": models.NewResponseHeader("string", "URL of the created user"),
            },
        },
        {
            StatusCode: http.StatusBadRequest,
            Body:       ErrorResponse{},
            Description: "Invalid request body",
        },
    })
```

### Pattern 4: File Upload

```go
// Handler Code
func UploadAvatar(w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm(10 << 20) // 10MB
    file, handler, err := r.FormFile("avatar")
    if err != nil {
        http.Error(w, "File required", http.StatusBadRequest)
        return
    }
    defer file.Close()
    // ... upload logic
}

// Generated GoSwag Code
gh.POST("/users/{id}/avatar").
    Summary("Upload avatar").
    Description("Upload user avatar image").
    Tags("users").
    Accepts("multipart/form-data").
    PathParam("id", "User ID", goswag.StringType, true).
    FileParam("avatar", "Avatar image file", true).
    Returns([]models.ReturnType{
        {StatusCode: http.StatusOK, Body: UploadResponse{}},
        {StatusCode: http.StatusBadRequest, Body: ErrorResponse{}},
    })
```

### Pattern 5: Authentication/Authorization

```go
// Handler Code
func GetProfile(w http.ResponseWriter, r *http.Request) {
    token := r.Header.Get("Authorization")
    if token == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    // ... validate token
    // ... fetch profile
}

// Generated GoSwag Code
gh.GET("/users/me").
    Summary("Get current user profile").
    Tags("users").
    Security("BearerAuth").
    HeaderParam("Authorization", "Bearer token", goswag.StringType, true).
    Returns([]models.ReturnType{
        {StatusCode: http.StatusOK, Body: User{}},
        {StatusCode: http.StatusUnauthorized, Body: ErrorResponse{}},
    })
```

### Pattern 6: Enum/Validation Patterns

```go
// Handler Code
func ListUsers(w http.ResponseWriter, r *http.Request) {
    status := r.URL.Query().Get("status")
    validStatuses := map[string]bool{"active": true, "inactive": true, "pending": true}
    if status != "" && !validStatuses[status] {
        http.Error(w, "Invalid status", http.StatusBadRequest)
        return
    }
    // ... fetch logic
}

// Generated GoSwag Code
gh.GET("/users").
    QueryParamWithOptions("status", "Filter by status", goswag.StringType, false,
        models.NewParamOptions().
            WithEnum("active", "inactive", "pending"),
    )
```

### Pattern 7: UUID Pattern Recognition

```go
// Handler Code
func GetUser(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    _, err := uuid.Parse(id)
    if err != nil {
        http.Error(w, "Invalid UUID", http.StatusBadRequest)
        return
    }
    // ... fetch logic
}

// Generated GoSwag Code
gh.GET("/users/{id}").
    PathParamWithOptions("id", "User ID", goswag.StringType, true,
        models.NewParamOptions().
            WithFormat("uuid").
            WithPattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"),
    )
```

## Framework-Specific Patterns

### net/http (Standard Library)

```go
// Handler Code
func handler(w http.ResponseWriter, r *http.Request) {
    // Method: r.Method
    // Path: from route registration
    // Query: r.URL.Query().Get()
    // Headers: r.Header.Get()
    // Body: json.NewDecoder(r.Body).Decode()
}

// Generated GoSwag Code
gh := goswag.NewHTTP(mux, defaultResponses...)
gh.GET("/path").Summary("...")
```

### Gin Framework

```go
// Handler Code
func handler(c *gin.Context) {
    // Path param: c.Param("id")
    // Query: c.Query("page")
    // Header: c.GetHeader("Authorization")
    // Body: c.ShouldBindJSON(&req)
    // Status: c.JSON(http.StatusOK, data)
}

// Generated GoSwag Code
gg := goswag.NewGin(ginEngine, defaultResponses...)
gg.GET("/path").Summary("...")
```

### Echo Framework

```go
// Handler Code
func handler(c echo.Context) error {
    // Path param: c.Param("id")
    // Query: c.QueryParam("page")
    // Header: c.Request().Header.Get("Authorization")
    // Body: c.Bind(&req)
    // Status: return c.JSON(http.StatusOK, data)
}

// Generated GoSwag Code
ge := goswag.NewEcho(defaultResponses...)
ge.GET("/path").Summary("...")
```

## Decision Tree for Parameter Types

```
Is it in the URL path?
├─ Yes → PathParam (always required)
│   └─ Check format: UUID? → WithFormat("uuid")
│                   Integer? → goswag.IntType
│                   String? → goswag.StringType
│
└─ No → Is it in query string?
    ├─ Yes → QueryParam
    │   └─ Check type: Atoi/ParseInt → IntType
    │                  ParseBool → BoolType
    │                  ParseFloat → NumberType
    │                  Default → StringType
    │
    └─ No → Is it in headers?
        ├─ Yes → HeaderParam
        │   └─ Usually StringType
        │
        └─ No → Is it in request body?
            ├─ Yes → Read() with struct
            │
            └─ No → Is it a file upload?
                └─ Yes → FileParam
```

## Common Response Patterns

### Success Responses
- `200 OK`: Standard success
- `201 Created`: Resource created (usually includes Location header)
- `204 No Content`: Success with no body

### Error Responses
- `400 Bad Request`: Invalid input, validation errors
- `401 Unauthorized`: Missing/invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict (e.g., duplicate)
- `500 Internal Server Error`: Server errors

## Best Practices for LLM Generation

### 1. Always Include Error Responses

```go
// Good: Includes error responses
Returns([]models.ReturnType{
    {StatusCode: http.StatusOK, Body: User{}},
    {StatusCode: http.StatusNotFound, Body: ErrorResponse{}},
    {StatusCode: http.StatusBadRequest, Body: ErrorResponse{}},
})

// Bad: Only success response
Returns([]models.ReturnType{
    {StatusCode: http.StatusOK, Body: User{}},
})
```

### 2. Infer Required Parameters

```go
// If handler checks for missing value and returns error → required
if param == "" {
    http.Error(w, "Param required", http.StatusBadRequest)
    return
}
// → Required: true

// If handler uses default value → optional
param := r.URL.Query().Get("page")
if param == "" {
    param = "1" // default
}
// → Required: false, WithDefault(1)
```

### 3. Extract Validation Rules

```go
// Pattern: Min/Max checks
if age < 18 || age > 120 {
    // → WithMinimum(18).WithMaximum(120)
}

// Pattern: Length checks
if len(name) < 2 || len(name) > 100 {
    // → WithMinLength(2).WithMaxLength(100)
}

// Pattern: Enum validation
validValues := []string{"active", "inactive"}
if !contains(validValues, status) {
    // → WithEnum("active", "inactive")
}
```

### 4. Group Related Routes

```go
// If routes share a prefix, use groups
gh.Group("/users").GET("/").Summary("List users")
gh.Group("/users").GET("/{id}").Summary("Get user")
// Better:
userGroup := gh.Group("/users")
userGroup.GET("/").Summary("List users")
userGroup.GET("/{id}").Summary("Get user")
```

### 5. Use Meaningful OperationIDs

```go
// Function: GetUserByID → OperationID("getUserByID")
// Function: CreateUser → OperationID("createUser")
// Function: UpdateUser → OperationID("updateUser")
// Function: DeleteUser → OperationID("deleteUser")
```

## Complete Example: Handler → GoSwag

### Input: Handler Code

```go
package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    
    "github.com/go-chi/chi"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

type ErrorResponse struct {
    Error string `json:"error"`
}

// GetUser retrieves a user by ID
func GetUser(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }
    
    // ... fetch user logic (assume user found)
    user := User{ID: id, Name: "John", Email: "john@example.com"}
    json.NewEncoder(w).Encode(user)
}

// ListUsers retrieves a paginated list of users
func ListUsers(w http.ResponseWriter, r *http.Request) {
    pageStr := r.URL.Query().Get("page")
    page := 1
    if pageStr != "" {
        p, _ := strconv.Atoi(pageStr)
        if p > 0 {
            page = p
        }
    }
    
    limitStr := r.URL.Query().Get("limit")
    limit := 10
    if limitStr != "" {
        l, _ := strconv.Atoi(limitStr)
        if l > 0 && l <= 100 {
            limit = l
        }
    }
    
    status := r.URL.Query().Get("status")
    validStatuses := map[string]bool{"active": true, "inactive": true}
    if status != "" && !validStatuses[status] {
        http.Error(w, "Invalid status", http.StatusBadRequest)
        return
    }
    
    // ... fetch users logic
    users := []User{{ID: 1, Name: "John"}}
    json.NewEncoder(w).Encode(map[string]interface{}{
        "users": users,
        "page":  page,
        "limit": limit,
    })
}

// CreateUser creates a new user
func CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    if req.Name == "" || len(req.Name) < 2 {
        http.Error(w, "Name must be at least 2 characters", http.StatusBadRequest)
        return
    }
    
    if req.Age < 18 || req.Age > 120 {
        http.Error(w, "Age must be between 18 and 120", http.StatusBadRequest)
        return
    }
    
    // ... create user logic
    user := User{ID: 1, Name: req.Name, Email: req.Email}
    w.Header().Set("Location", "/users/1")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
```

### Output: Generated GoSwag Code

```go
package main

import (
    "net/http"
    
    "github.com/r0bertson/goswag"
    "github.com/r0bertson/goswag/models"
)

func setupSwagger(mux *http.ServeMux) {
    defaultResponses := []models.ReturnType{
        {StatusCode: http.StatusBadRequest, Body: ErrorResponse{}},
        {StatusCode: http.StatusInternalServerError, Body: ErrorResponse{}},
    }
    
    gh := goswag.NewHTTP(mux, defaultResponses...)
    
    // GetUser handler
    gh.GET("/users/{id}").
        Summary("Get user").
        Description("Retrieve a user by ID").
        Tags("users").
        OperationID("getUser").
        PathParamWithOptions("id", "User ID", goswag.IntType, true,
            models.NewParamOptions().
                WithExample(1),
        ).
        Returns([]models.ReturnType{
            {StatusCode: http.StatusOK, Body: User{}},
            {StatusCode: http.StatusBadRequest, Body: ErrorResponse{}},
        })
    
    // ListUsers handler
    gh.GET("/users").
        Summary("List users").
        Description("Retrieve a paginated list of users").
        Tags("users").
        OperationID("listUsers").
        QueryParamWithOptions("page", "Page number", goswag.IntType, false,
            models.NewParamOptions().
                WithDefault(1).
                WithMinimum(1).
                WithExample(1),
        ).
        QueryParamWithOptions("limit", "Items per page", goswag.IntType, false,
            models.NewParamOptions().
                WithDefault(10).
                WithMinimum(1).
                WithMaximum(100).
                WithExample(10),
        ).
        QueryParamWithOptions("status", "Filter by status", goswag.StringType, false,
            models.NewParamOptions().
                WithEnum("active", "inactive"),
        ).
        Returns([]models.ReturnType{
            {StatusCode: http.StatusOK, Body: UserListResponse{}},
            {StatusCode: http.StatusBadRequest, Body: ErrorResponse{}},
        })
    
    // CreateUser handler
    gh.POST("/users").
        Summary("Create user").
        Description("Create a new user account").
        Tags("users").
        OperationID("createUser").
        Accepts("application/json").
        Read(CreateUserRequest{}).
        ReadFieldDescriptions(map[string]string{
            "name":  "User's full name (required, min 2 characters)",
            "email": "User's email address (required)",
            "age":   "User's age (required, must be between 18 and 120)",
        }).
        Returns([]models.ReturnType{
            {
                StatusCode: http.StatusCreated,
                Body:       User{},
                Headers: map[string]*models.ResponseHeader{
                    "Location": models.NewResponseHeader("string", "URL of the created user"),
                },
            },
            {StatusCode: http.StatusBadRequest, Body: ErrorResponse{}},
        })
    
    gh.GenerateSwagger()
}
```

## Common Pitfalls to Avoid

1. **Don't assume all query params are strings** - Check parsing logic
2. **Don't forget error responses** - Always include 400, 500, etc.
3. **Don't ignore validation logic** - Extract min/max/enum from code
4. **Don't miss response headers** - Check `w.Header().Set()` calls
5. **Don't forget file uploads** - Look for `multipart` usage
6. **Don't ignore authentication** - Check for token validation
7. **Don't skip default values** - Look for default assignments

## Output Format

Generate clean, idiomatic Go code that:
- Uses proper GoSwag method chaining
- Includes all inferred metadata
- Follows the patterns shown in examples
- Is ready to compile and generate Swagger docs

## Reference Files

- `MIGRATION_GUIDE.md` - For understanding feature differences
- `examples/comprehensive_example.go` - For complete feature examples
- `README.md` - For API reference and usage patterns

## Task Checklist

When generating GoSwag code, ensure:

- [ ] All routes are converted
- [ ] Path parameters are identified and typed correctly
- [ ] Query parameters include validation when present in code
- [ ] Request bodies are identified and typed
- [ ] Response status codes match handler logic
- [ ] Response bodies match return types
- [ ] Error responses are included
- [ ] Response headers are captured
- [ ] Authentication/security is identified
- [ ] OperationIDs are meaningful and unique
- [ ] Tags are logical and consistent
- [ ] Groups are used for route organization
- [ ] Default responses are configured appropriately
