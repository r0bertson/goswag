package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/r0bertson/goswag/models"
)

func TestNewHTTP(t *testing.T) {
	mux := http.NewServeMux()
	swagger := NewHTTP(mux)

	if swagger == nil {
		t.Error("Expected NewHTTP to return a non-nil value")
	}

	if swagger.Mux() != mux {
		t.Error("Expected Mux() to return the same mux instance")
	}
}

func TestHTTP_GET(t *testing.T) {
	mux := http.NewServeMux()
	swagger := NewHTTP(mux)

	var handlerCalled bool
	handler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}

	route := swagger.GET("/test", handler)
	if route == nil {
		t.Error("Expected GET to return a non-nil route")
	}

	// Test the route
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHTTP_POST(t *testing.T) {
	mux := http.NewServeMux()
	swagger := NewHTTP(mux)

	var handlerCalled bool
	handler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusCreated)
	}

	route := swagger.POST("/test", handler)
	if route == nil {
		t.Error("Expected POST to return a non-nil route")
	}

	// Test the route
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestHTTP_Group(t *testing.T) {
	mux := http.NewServeMux()
	swagger := NewHTTP(mux)

	group := swagger.Group("/api")
	if group == nil {
		t.Error("Expected Group to return a non-nil group")
	}

	var handlerCalled bool
	handler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}

	route := group.GET("/users", handler)
	if route == nil {
		t.Error("Expected GET to return a non-nil route")
	}

	// Test the route
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHTTP_MethodNotAllowed(t *testing.T) {
	mux := http.NewServeMux()
	swagger := NewHTTP(mux)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	swagger.GET("/test", handler)

	// Test with wrong method
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHTTPRoute_SwaggerMethods(t *testing.T) {
	mux := http.NewServeMux()
	swagger := NewHTTP(mux)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	type TestResponse struct {
		Message string `json:"message"`
	}

	route := swagger.GET("/test", handler).
		Summary("Test endpoint").
		Description("A test endpoint for swagger generation").
		Tags("test").
		Accepts("application/json").
		Produces("application/json").
		Returns([]models.ReturnType{
			{
				StatusCode: 200,
				Body:       TestResponse{},
			},
		}).
		QueryParam("id", "User ID", "string", true).
		HeaderParam("Authorization", "Bearer token", "string", true).
		PathParam("user_id", "User ID", "string", true)

	if route == nil {
		t.Error("Expected route to be non-nil")
	}
}

func TestHTTPRoute_SamePathDifferentMethod(t *testing.T) {
	mux := http.NewServeMux()
	swagger := NewHTTP(mux)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	type TestResponse struct {
		Message string `json:"message"`
	}

	swagger.GET("/test", handler).
		Summary("Test endpoint").
		Description("A test endpoint for swagger generation").
		Tags("test").
		Accepts("application/json").
		Produces("application/json").
		Returns([]models.ReturnType{
			{
				StatusCode: 200,
				Body:       TestResponse{},
			},
		}).
		QueryParam("id", "User ID", "string", true).
		HeaderParam("Authorization", "Bearer token", "string", true).
		PathParam("user_id", "User ID", "string", true)

	swagger.POST("/test", handler).
		Summary("Test endpoint").
		Description("A test endpoint for swagger generation").
		Tags("test").
		Accepts("application/json").
		Produces("application/json").
		Returns([]models.ReturnType{
			{
				StatusCode: 200,
				Body:       TestResponse{},
			},
		}).
		QueryParam("id", "User ID", "string", true).
		HeaderParam("Authorization", "Bearer token", "string", true).
		PathParam("user_id", "User ID", "string", true)

	swagger.mux.Handle("/test", swagger.mux)
}

func TestHTTPRoute_ReadFieldDescriptions(t *testing.T) {
	mux := http.NewServeMux()
	swagger := NewHTTP(mux)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	type CreateUserRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	route := swagger.POST("/users", handler).
		Read(CreateUserRequest{}).
		ReadFieldDescriptions(map[string]string{
			"name":  "The full name of the user",
			"email": "The user's email address",
		})

	if route == nil {
		t.Error("Expected route to be non-nil")
	}

	// Verify that ReadFieldDescriptions was set
	httpRoute := route.(*httpRoute)
	if httpRoute.Route.ReadFieldDescriptions == nil {
		t.Error("Expected ReadFieldDescriptions to be set")
	}

	if httpRoute.Route.ReadFieldDescriptions["name"] != "The full name of the user" {
		t.Errorf("Expected name description to be 'The full name of the user', got '%s'", httpRoute.Route.ReadFieldDescriptions["name"])
	}

	if httpRoute.Route.ReadFieldDescriptions["email"] != "The user's email address" {
		t.Errorf("Expected email description to be 'The user's email address', got '%s'", httpRoute.Route.ReadFieldDescriptions["email"])
	}
}

func TestHTTPRoute_ReturnsWithFieldDescriptions(t *testing.T) {
	mux := http.NewServeMux()
	swagger := NewHTTP(mux)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	type UserResponse struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	route := swagger.GET("/users/{id}", handler).
		Returns([]models.ReturnType{
			{
				StatusCode: 200,
				Body:       UserResponse{},
				FieldDescriptions: map[string]string{
					"id":    "Unique identifier for the user",
					"name":  "The full name of the user",
					"email": "The user's email address",
				},
			},
		})

	if route == nil {
		t.Error("Expected route to be non-nil")
	}

	// Verify that FieldDescriptions was set
	httpRoute := route.(*httpRoute)
	if len(httpRoute.Route.Returns) == 0 {
		t.Error("Expected Returns to be set")
	}

	returnType := httpRoute.Route.Returns[0]
	if returnType.FieldDescriptions == nil {
		t.Error("Expected FieldDescriptions to be set")
	}

	if returnType.FieldDescriptions["id"] != "Unique identifier for the user" {
		t.Errorf("Expected id description to be 'Unique identifier for the user', got '%s'", returnType.FieldDescriptions["id"])
	}
}
