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
