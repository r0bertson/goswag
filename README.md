# goswag

<p align="center">
 <b>To help you auto generate swagger for your golang APIs</b><br>
 <img src='./assets/gopher-swagger.png' width='350'> <br>
    <a href="https://github.com/r0bertson/goswag/tags" alt="GitHub tag">
     <img src="https://img.shields.io/github/tag/r0bertson/goswag.svg" />
    </a>
    <a href='https://coveralls.io/github/r0bertson/goswag?branch=main'>
     <img src='https://coveralls.io/repos/github/r0bertson/goswag/badge.svg?branch=main' alt='Coverage Status' />
    </a>
    <a href="https://github.com/r0bertson/goswag/actions">
     <img src="https://github.com/r0bertson/goswag/actions/workflows/ci.yaml/badge.svg" alt="build status">
    </a>
    <a href="https://github.com/r0bertson/goswag/contributors" alt="Contributors">
     <img src="https://img.shields.io/github/contributors/r0bertson/goswag" />
    </a>
    <a href="https://opensource.org/licenses/MIT">
     <img src="https://img.shields.io/badge/License-MIT-yellow.svg" />
    </a>
    <a href='https://goreportcard.com/badge/github.com/r0bertson/goswag'>
     <img src='https://goreportcard.com/badge/github.com/r0bertson/goswag' alt='Go Report'/>
    </a>
</p>

## Introduction
It will extend your Go framework by providing methods to generate a separate file containing all the necessary comments. These comments enable the Swag library to effortlessly generate Swagger files.

### - Why: 
I was searching for an automated method to generate Swagger documentation for Golang APIs. I came across the [`swaggo/swag`](https://github.com/swaggo/swag) lib, which seems to be the most popular choice or, at the very least, has numerous articles and tutorials promoting its use. However, I was dissatisfied with the process of adding extensive comments throughout the main file and other files containing handler functions. As a result, I decided to develop the goswag library.

### - How:
Goswag simplifies the process of integrating Swagger documentation generation into your Go projects. By seamlessly aligning with your chosen framework's usage patterns, Goswag ensures a smooth integration experience.

For instance, if your framework uses a POST method with specific parameters, Goswag mirrors these method and parameters, streamlining the integration process. This principle applies across supported frameworks, ensuring consistency and ease of use.

**Write this:**  
<img src="./assets/goswag-way.png" alt="goswag way" width="500"/>  

**Instead of this:**

<img src="./assets/swag-way.png" alt="swag way" width="500"/>  

**And will have the same result:**

<img src="./assets/swagger-result.png" alt="swagger result" width="500"/>

### - Supported Libraries
- [echo](https://github.com/labstack/echo) 
- [gin](https://github.com/gin-gonic/gin)

## Getting started

### 1 - Modifying your current project
When initializing your current framework, such as `e := echo.New()`, begin by replacing it with `ge := goswag.NewEcho()` or `gg := goswag.NewGin(gin)` by passing the Gin instance as a parameter for the Gin framework. 

### 2 - Using original framework configuration
If you intend to utilize the framework with alternative configurations, for instance: `e.Debug = true`, you can access `e` as follows: `ge.Echo().Debug = true` achieving identical results.

### 3 - Add annotations to your routes:
After completing the initial setup, your routes are established without errors and require no further changes. However, your routes will now possess additional methods:
- `Summary`: Provides a brief overview of your route.
- `Description`: Offers a detailed description of your route (If not set, it defaults to the summary).
- `Accepts`: The default value is *json*. f you wish to incorporate different values, please refer to the list of possible values [here](https://github.com/swaggo/swag#mime-types). 
- `Produces`: The default value is *json*. To include different values, consult the list of possible options [here](https://github.com/swaggo/swag#mime-types).
- `Read`: ISpecifies the request body received by your routes.
- `Returns`: Is an array of ReturnType{}. Your route can have multiples returns (e.g., success, errors e etc). Refer to the [interface reference](https://github.com/r0bertson/goswag/blob/main/models/models.go#L64) for detailed usage information.
```go
type ReturnType struct {
	StatusCode int
	Body       any
	// example: map[jsonFieldName]fieldType{}
	OverrideStructFields map[string]any
}
```
- `QueryParam`: Defines the query parameters of the route and specifies if they are required.
- `HeaderParam`: Defines the header parameters of the route and specifies if they are required.
- `PathParam`: Defines the path parameters of the route and specifies if they are required.

### 4 - Generating your Swagger Documentation
The method used to instantiate your router, either `NewEcho()` or `NewGin()` includes a function called `GenerateSwagger()`.  
After setting up all your routes (including annotations), you can invoke `GenerateSwagger()` to generate your swagger documentation. However, this implies that if your route setup relies on services like a running database or RabbitMQ, you can only generate your Swagger documentation when your entire infrastructure is operational, which is not ideal.

#### Recommended Approach:
The recommended approach is to have a separate main file where you do not need to provide real connections for your route setup.

Consider a file named `server.go` with the following setup:
```go
func SetupRoutes(db *sql.DB) *echo.Echo {
    e := echo.New()
    handlers := handlers.New(db)
    e.GET("/hello", handlers.HandleHello)
     
    return e
}
```
In order to start your router, you currently need to provide a database connection. However, this requirement is unnecessary for generating Swagger documentation.

#### Recommended Approach Setup
- Install [`swag`](https://github.com/swaggo/swag/blob/master/README.md#getting-started). It is necessary for generate swagger documents.
- Create a folder in your root project folder named `goswag`.
- Inside the `goswag` folder, create a file called `main.go` with package main.
- Inside of this file, create your main function that will invoke your routerSetup.
    - You can add comments to the main.go file for your Swagger, similar to [this example](https://github.com/swaggo/swag/blob/master/README.md#how-to-use-it-with-gin) in item 2.

```go
// @title           GoSwag example API
// @version         1.0
func main() {
    // Here you have already used goswag for your route setup, added annotations and change it return to goswag.Echo or Gin interfaces
    ge := server.SetupRoutes(nil)
    ge.GenerateSwagger() //will generate your swagger
}
```
- Create a Makefile and add the following command:
```Make
.PHONY: docs
docs:
    @go install github.com/swaggo/swag/cmd/swag@latest
	@cd goswag && \
	go run main.go && \
	cd .. && \
	swag init --pdl=2 --parseInternal -g ./goswag/main.go -o ./docs && \
	swag fmt -d ./goswag/
```

You can now execute the `make docs` command.  
It will generate a new `goswag.go` file inside of your `goswag` directory. This file includes all necessary handlers and comments for the Swag library to generate the Swagger files inside the `docs` directory.

**NOTE**: after the first generation, the `doc.go` file in the `docs` folder will import Swag library. If you haven't used Swag in your project before, you'll need to run `go mod tidy` to ensure the swag package is included in your `go.mod` file. 

## Default Response for all routes
You can add a default responses to all routes when you instantiate the swagger.  
To add default responses, you need to define your list of default returns and add it to instance, ex:
```go
defaultResponses := []models.ReturnType{
    {
        StatusCode: http.StatusBadRequest,
        Body: YourStructOfError,
    },
    {
        StatusCode: http.StatusUnauthorized,
        Body: YourStructOfError,
    },
}

// pass the default responses to instance of your chosen framework
e := NewEcho(defaultResponses)
```
Then it will add default responses for all routes, like the example below:
```go
//	@Summary		Logout
//	@Description	Logout the user
//	@Tags			auth
//	@Param			user-token	header	string	true	"User access token"
//	@Success		200
//	@Failure		400	{object}	YourStructOfError
//	@Failure		401	{object}	YourStructOfError
//	@Router			/auth/logout [post]
func handleLogout() {} //nolint:unused 

//	@Summary		Login
//	@Description	Login the user
//	@Tags			auth
//	@Param			user-token	header	string	true	"User access token"
//	@Success		200
//	@Failure		400	{object}	YourStructOfError
//	@Failure		401	{object}	YourStructOfError
//	@Router			/auth/logout [post]
func handleLogin() {} //nolint:unused 
```
`NewEcho()` and `NewGin()` includes de defaultResponses parameter as optional, then you can pass your default responses only if you want =].
## Example of Usage
To see an example of usage, you can check this [repository](https://github.com/r0bertson/go_boilerplate).
The necessary modifications are located in `transport/rest/server.go` and the `router.go` file inside of each route directory in `transport/rest/routes/`.

## Enhanced Features (New in v2.0)

### Enhanced Parameters

GoSwag now supports comprehensive parameter options including validation, defaults, examples, and more:

```go
// Query parameter with validation and default value
gh.GET("/users").
    QueryParamWithOptions("page", "Page number", goswag.IntType, false,
        models.NewParamOptions().
            WithDefault(1).
            WithMinimum(1).
            WithMaximum(100).
            WithExample(1),
    ).
    // Enum values
    QueryParamWithOptions("status", "Filter by status", goswag.StringType, false,
        models.NewParamOptions().
            WithEnum("active", "inactive", "pending").
            WithDefault("active"),
    ).
    // Collection format (for arrays)
    QueryParamWithOptions("tags", "Filter by tags", goswag.StringType, false,
        models.NewParamOptions().
            WithCollectionFormat("csv").
            WithExample("tag1,tag2,tag3"),
    )

// Path parameter with format and pattern validation
gh.GET("/users/{id}").
    PathParamWithOptions("id", "User ID", goswag.StringType, true,
        models.NewParamOptions().
            WithFormat("uuid").
            WithPattern("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"),
    )
```

**Available parameter options:**
- `WithDefault(value)` - Set default value
- `WithExample(value)` - Set example value
- `WithMinimum(min)` - Set minimum value (for numbers)
- `WithMaximum(max)` - Set maximum value (for numbers)
- `WithMinLength(min)` - Set minimum length (for strings)
- `WithMaxLength(max)` - Set maximum length (for strings)
- `WithPattern(pattern)` - Set regex pattern (for strings)
- `WithEnum(values...)` - Set allowed enum values
- `WithFormat(format)` - Set format (e.g., "uuid", "email", "date-time")
- `WithCollectionFormat(format)` - Set collection format ("csv", "ssv", "tsv", "pipes", "multi")
- `WithAllowEmptyValue()` - Allow empty values

### Form Data and File Uploads

```go
// Form data parameter
gh.POST("/users/{id}/avatar").
    FormDataParam("description", "Image description", goswag.StringType, false).
    FileParam("avatar", "Avatar image file", true)
```

### Operation-Level Enhancements

```go
gh.GET("/users/{id}").
    Summary("Get user by ID").
    OperationID("getUserById").           // Unique operation identifier
    Deprecated().                         // Mark as deprecated
    Schemes("https", "wss").              // Override global schemes
    ExternalDocs("https://docs.example.com/users", "User API documentation")
```

### Response Enhancements

```go
gh.GET("/users").
    Returns([]models.ReturnType{
        {
            StatusCode: http.StatusOK,
            Body:       UserListResponse{},
            Description: "Successfully retrieved users",
            // Response headers
            Headers: map[string]*models.ResponseHeader{
                "X-Total-Count": models.NewResponseHeader("integer", "Total number of users"),
                "X-Request-ID":  models.NewResponseHeader("string", "Request ID").WithFormat("uuid"),
            },
            // Response examples
            Examples: map[string]interface{}{
                "application/json": map[string]interface{}{
                    "users": []User{{ID: 1, Name: "John Doe"}},
                    "total": 1,
                },
            },
        },
    })
```

### Global Configuration

Configure global Swagger settings:

```go
config := models.NewSwaggerConfig().
    WithHost("api.example.com").
    WithBasePath("/v1").
    WithSchemes("https").
    WithContact(models.NewContactInfo("API Support", "support@example.com", "")).
    WithLicense(models.NewLicenseInfo("MIT", "https://opensource.org/licenses/MIT")).
    WithTermsOfService("https://example.com/terms").
    WithExternalDocs(models.NewExternalDocs("https://docs.example.com", "API Documentation")).
    WithGlobalSecurity("BearerAuth")

gh := goswag.NewHTTPWithConfig(mux, config, defaultResponses...)
```

### Tag Metadata

Add descriptions and external documentation to route groups (tags):

```go
userGroup := gh.Group("/users").
    TagDescription("User management operations").
    TagExternalDocs("https://docs.example.com/users", "User API documentation")

userGroup.GET("/").
    Summary("List users").
    Returns([]models.ReturnType{...})
```

## More features
You can add description for fields, add if they are required or not.  
For this struct fields features and more, you can follow the [swag documentation](https://github.com/swaggo/swag) to understand how to add it.

## Migration Guide

### Upgrading from v1.x to v2.0

All existing code continues to work without changes. New features are opt-in:

**Before (v1.x):**
```go
gh := goswag.NewHTTP(mux, defaultResponses...)
gh.GET("/users").
    QueryParam("page", "Page number", goswag.IntType, false)
```

**After (v2.0) - Same code still works:**
```go
gh := goswag.NewHTTP(mux, defaultResponses...)
gh.GET("/users").
    QueryParam("page", "Page number", goswag.IntType, false)
```

**After (v2.0) - Using new features:**
```go
// Option 1: Use enhanced parameters
gh.GET("/users").
    QueryParamWithOptions("page", "Page number", goswag.IntType, false,
        models.NewParamOptions().WithDefault(1).WithMinimum(1),
    )

// Option 2: Add global configuration
config := models.NewSwaggerConfig().
    WithHost("api.example.com").
    WithBasePath("/v1")
gh := goswag.NewHTTPWithConfig(mux, config, defaultResponses...)

// Option 3: Add operation metadata
gh.GET("/users/{id}").
    OperationID("getUserById").
    Deprecated()
```

### Key Changes

1. **New Methods**: All new features use new method names (e.g., `QueryParamWithOptions` instead of modifying `QueryParam`)
2. **Backward Compatible**: All existing methods work exactly as before
3. **Opt-in**: New features are only used when explicitly called
4. **No Breaking Changes**: No changes to existing method signatures

## Complete Example

See [examples/comprehensive_example.go](./examples/comprehensive_example.go) for a complete example demonstrating all features.

## LLM Integration

If you're using an LLM to automatically generate Swagger documentation from unannotated handlers, see [LLM_GUIDE.md](./LLM_GUIDE.md) for detailed instructions on:
- Analyzing handler code patterns
- Inferring Swagger metadata from code
- Transforming handlers to GoSwag code
- Common patterns and best practices

## Contributing

**Contributions are welcomed. :)**  
You can contribute not only with code but also by enhancing these `README.md` docs, writing articles, using Goswag and providing feedback.  
  
To contribute, please follow these steps:

1. Fork the repository
2. Create a new feature branch (`git checkout -b feature/<FEATURE NAME>`)
3. Make the necessary changes
4. Commit your changes (`git commit -m "Add some feature"`)
5. Push your changes to your forked repository (`git push origin feature/<FEATURE NAME>`)
6. Create a pull request to the main branch of the repository

## License

Goswag is [MIT licensed](./LICENSE).

### TODO:
- NewGin() does not implement the method Match or Any from gin, because there are no way (yet) to define the same (summary,responses,bodies) for all methods
