package http

import (
	"net/http"

	"github.com/r0bertson/goswag/internal/generator"
	"github.com/r0bertson/goswag/models"
)

type httpSwagger struct {
	mux              *http.ServeMux
	groups           []*httpGroup
	routes           []*httpRoute
	defaultResponses []models.ReturnType
}

func NewHTTP(mux *http.ServeMux, defaultResponses ...models.ReturnType) *httpSwagger {
	return &httpSwagger{
		mux:              mux,
		defaultResponses: defaultResponses,
	}
}

func (s *httpSwagger) Mux() *http.ServeMux {
	return s.mux
}

func (s *httpSwagger) GenerateSwagger() {
	generator.GenerateSwagger(toGoSwagRoute(s.routes), toGoSwagGroup(s.groups), s.defaultResponses)
}

func (s *httpSwagger) Group(relativePath string, handlers ...http.HandlerFunc) models.HTTPRouter {
	g := &httpGroup{prefix: relativePath, mux: s.mux, groupName: relativePath}
	s.groups = append(s.groups, g)

	return g
}

func (s *httpSwagger) Handle(httpMethod, relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	// For net/http, we need to create a custom handler that checks the method
	handler := createMethodHandler(httpMethod, handlers...)
	s.mux.Handle(relativePath, handler)

	hr := &httpRoute{
		Route: generator.Route{
			Path:     relativePath,
			Method:   httpMethod,
			FuncName: getFuncName(handlers...),
		},
	}

	s.routes = append(s.routes, hr)

	return hr
}

func (s *httpSwagger) POST(relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	return s.Handle(http.MethodPost, relativePath, handlers...)
}

func (s *httpSwagger) GET(relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	return s.Handle(http.MethodGet, relativePath, handlers...)
}

func (s *httpSwagger) PUT(relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	return s.Handle(http.MethodPut, relativePath, handlers...)
}

func (s *httpSwagger) DELETE(relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	return s.Handle(http.MethodDelete, relativePath, handlers...)
}

func (s *httpSwagger) PATCH(relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	return s.Handle(http.MethodPatch, relativePath, handlers...)
}

func (s *httpSwagger) OPTIONS(relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	return s.Handle(http.MethodOptions, relativePath, handlers...)
}

func (s *httpSwagger) HEAD(relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	return s.Handle(http.MethodHead, relativePath, handlers...)
}

type httpGroup struct {
	prefix    string
	mux       *http.ServeMux
	groupName string
	routes    []*httpRoute
}

func (g *httpGroup) Handle(httpMethod, relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	fullPath := getFullPath(g.prefix, relativePath)
	handler := createMethodHandler(httpMethod, handlers...)
	g.mux.Handle(fullPath, handler)

	hr := &httpRoute{
		Route: generator.Route{
			Path:     fullPath,
			Method:   httpMethod,
			FuncName: getFuncName(handlers...),
		},
	}

	g.routes = append(g.routes, hr)

	return hr
}

func (g *httpGroup) POST(relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	return g.Handle(http.MethodPost, relativePath, handlers...)
}

func (g *httpGroup) GET(relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	return g.Handle(http.MethodGet, relativePath, handlers...)
}

func (g *httpGroup) PUT(relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	return g.Handle(http.MethodPut, relativePath, handlers...)
}

func (g *httpGroup) DELETE(relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	return g.Handle(http.MethodDelete, relativePath, handlers...)
}

func (g *httpGroup) PATCH(relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	return g.Handle(http.MethodPatch, relativePath, handlers...)
}

func (g *httpGroup) OPTIONS(relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	return g.Handle(http.MethodOptions, relativePath, handlers...)
}

func (g *httpGroup) HEAD(relativePath string, handlers ...http.HandlerFunc) models.Swagger {
	return g.Handle(http.MethodHead, relativePath, handlers...)
}

type httpRoute struct {
	Route generator.Route
}

func (r *httpRoute) Summary(summary string) models.Swagger {
	r.Route.Summary = summary
	return r
}

func (r *httpRoute) Description(description string) models.Swagger {
	r.Route.Description = description
	return r
}

func (r *httpRoute) Tags(tags ...string) models.Swagger {
	r.Route.Tags = tags
	return r
}

func (r *httpRoute) Accepts(accepts ...string) models.Swagger {
	r.Route.Accepts = accepts
	return r
}

func (r *httpRoute) Produces(produces ...string) models.Swagger {
	r.Route.Produces = produces
	return r
}

func (r *httpRoute) Read(reads interface{}) models.Swagger {
	r.Route.Reads = reads
	return r
}

func (r *httpRoute) Returns(returns []models.ReturnType) models.Swagger {
	r.Route.Returns = returns
	return r
}

func (r *httpRoute) QueryParam(name, description, paramType string, required bool) models.Swagger {
	r.Route.QueryParams = append(r.Route.QueryParams, generator.Param{
		Name:        name,
		Description: description,
		ParamType:   paramType,
		Required:    required,
	})
	return r
}

func (r *httpRoute) HeaderParam(name, description, paramType string, required bool) models.Swagger {
	r.Route.HeaderParams = append(r.Route.HeaderParams, generator.Param{
		Name:        name,
		Description: description,
		ParamType:   paramType,
		Required:    required,
	})
	return r
}

func (r *httpRoute) PathParam(name, description, paramType string, required bool) models.Swagger {
	r.Route.PathParams = append(r.Route.PathParams, generator.Param{
		Name:        name,
		Description: description,
		ParamType:   paramType,
		Required:    required,
	})
	return r
}

// createMethodHandler creates a handler that checks the HTTP method before executing the handlers
func createMethodHandler(method string, handlers ...http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Execute all handlers in sequence
		for _, handler := range handlers {
			handler(w, r)
		}
	})
}
