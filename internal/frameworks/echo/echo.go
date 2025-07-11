package echo

import (
	"github.com/labstack/echo/v4"
	"github.com/r0bertson/goswag/internal/generator"
	"github.com/r0bertson/goswag/models"
)

type echoSwagger struct {
	e                *echo.Echo
	groups           []*echoGroup
	routes           []*echoRoute
	defaultResponses []models.ReturnType
}

func NewEcho(defaultResponses ...models.ReturnType) *echoSwagger {
	return &echoSwagger{
		e:                echo.New(),
		defaultResponses: defaultResponses,
	}
}

func (s *echoSwagger) Echo() *echo.Echo {
	return s.e
}

func (s *echoSwagger) GenerateSwagger() {
	generator.GenerateSwagger(toGoSwagRoute(s.routes), toGoSwagGroup(s.groups), s.defaultResponses)
}

func (s *echoSwagger) Group(prefix string, m ...echo.MiddlewareFunc) models.EchoGroup {
	g := &echoGroup{g: s.e.Group(prefix, m...), groupName: prefix}
	s.groups = append(s.groups, g)

	return g
}

func (s *echoSwagger) POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) models.Swagger {
	r := s.e.POST(path, h, m...)

	er := &echoRoute{
		Route: generator.Route{
			Path:     r.Path,
			Method:   r.Method,
			FuncName: getFuncName(r.Name),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoSwagger) GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) models.Swagger {
	r := s.e.GET(path, h, m...)

	er := &echoRoute{
		Route: generator.Route{
			Path:     r.Path,
			Method:   r.Method,
			FuncName: getFuncName(r.Name),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoSwagger) PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) models.Swagger {
	r := s.e.PUT(path, h, m...)

	er := &echoRoute{
		Route: generator.Route{
			Path:     r.Path,
			Method:   r.Method,
			FuncName: getFuncName(r.Name),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoSwagger) DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) models.Swagger {
	r := s.e.DELETE(path, h, m...)

	er := &echoRoute{
		Route: generator.Route{
			Path:     r.Path,
			Method:   r.Method,
			FuncName: getFuncName(r.Name),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoSwagger) PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) models.Swagger {
	r := s.e.PATCH(path, h, m...)

	er := &echoRoute{
		Route: generator.Route{
			Path:     r.Path,
			Method:   r.Method,
			FuncName: getFuncName(r.Name),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoSwagger) OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) models.Swagger {
	r := s.e.OPTIONS(path, h, m...)

	er := &echoRoute{
		Route: generator.Route{
			Path:     r.Path,
			Method:   r.Method,
			FuncName: getFuncName(r.Name),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoSwagger) HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) models.Swagger {
	r := s.e.HEAD(path, h, m...)

	er := &echoRoute{
		Route: generator.Route{
			Path:     r.Path,
			Method:   r.Method,
			FuncName: getFuncName(r.Name),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

type echoGroup struct {
	g         *echo.Group
	groupName string
	groups    []*echoGroup
	routes    []*echoRoute
}

// Group creates a new sub-group with prefix and optional sub-group-level middleware.
func (s *echoGroup) Group(prefix string, m ...echo.MiddlewareFunc) models.EchoGroup {
	g := &echoGroup{g: s.g.Group(prefix, m...), groupName: prefix}
	s.groups = append(s.groups, g)

	return g
}

func (s *echoGroup) POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) models.Swagger {
	r := s.g.POST(path, h, m...)

	er := &echoRoute{
		Route: generator.Route{
			Path:     r.Path,
			Method:   r.Method,
			FuncName: getFuncName(r.Name),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoGroup) GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) models.Swagger {
	r := s.g.GET(path, h, m...)

	er := &echoRoute{
		Route: generator.Route{
			Path:     r.Path,
			Method:   r.Method,
			FuncName: getFuncName(r.Name),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoGroup) PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) models.Swagger {
	r := s.g.PUT(path, h, m...)

	er := &echoRoute{
		Route: generator.Route{
			Path:     r.Path,
			Method:   r.Method,
			FuncName: getFuncName(r.Name),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoGroup) DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) models.Swagger {
	r := s.g.DELETE(path, h, m...)

	er := &echoRoute{
		Route: generator.Route{
			Path:     r.Path,
			Method:   r.Method,
			FuncName: getFuncName(r.Name),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoGroup) PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) models.Swagger {
	r := s.g.PATCH(path, h, m...)

	er := &echoRoute{
		Route: generator.Route{
			Path:     r.Path,
			Method:   r.Method,
			FuncName: getFuncName(r.Name),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoGroup) OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) models.Swagger {
	r := s.g.OPTIONS(path, h, m...)

	er := &echoRoute{
		Route: generator.Route{
			Path:     r.Path,
			Method:   r.Method,
			FuncName: getFuncName(r.Name),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

func (s *echoGroup) HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) models.Swagger {
	r := s.g.HEAD(path, h, m...)

	er := &echoRoute{
		Route: generator.Route{
			Path:     r.Path,
			Method:   r.Method,
			FuncName: getFuncName(r.Name),
		},
	}

	s.routes = append(s.routes, er)

	return er
}

type echoRoute struct {
	generator.Route
}

func (r *echoRoute) Summary(value string) models.Swagger {
	r.Route.Summary = value
	return r
}

func (r *echoRoute) Description(value string) models.Swagger {
	r.Route.Description = value
	return r
}

func (r *echoRoute) Tags(value ...string) models.Swagger {
	r.Route.Tags = value
	return r
}

func (r *echoRoute) Accepts(value ...string) models.Swagger {
	r.Route.Accepts = value
	return r
}

func (r *echoRoute) Produces(value ...string) models.Swagger {
	r.Route.Produces = value
	return r
}

func (r *echoRoute) Read(value interface{}) models.Swagger {
	r.Route.Reads = value
	return r
}

func (r *echoRoute) Returns(returns []models.ReturnType) models.Swagger {
	r.Route.Returns = returns
	return r
}

func (r *echoRoute) QueryParam(name, description, paramType string, required bool) models.Swagger {
	r.Route.QueryParams = append(r.Route.QueryParams, generator.Param{
		Name:        name,
		Description: description,
		ParamType:   paramType,
		Required:    required,
	})

	return r
}

func (r *echoRoute) HeaderParam(name, description, paramType string, required bool) models.Swagger {
	r.Route.HeaderParams = append(r.Route.HeaderParams, generator.Param{
		Name:        name,
		Description: description,
		ParamType:   paramType,
		Required:    required,
	})

	return r
}

func (r *echoRoute) PathParam(name, description, paramType string, required bool) models.Swagger {
	r.Route.PathParams = append(r.Route.PathParams, generator.Param{
		Name:        name,
		Description: description,
		ParamType:   paramType,
		Required:    required,
	})

	return r
}
