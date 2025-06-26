package goswag

import (
	"net/http"

	httpWrapper "github.com/r0bertson/goswag/internal/frameworks/http"
	"github.com/r0bertson/goswag/models"
)

type HTTP interface {
	models.HTTPRouter
	models.HTTPGroup
	GenerateSwagger()
	Mux() *http.ServeMux
}

// NewHTTP returns the interface that wraps the basic HTTP methods and add the swagger methods
// defaultResponses is an optional parameter that can be used to set the default responses for all routes
func NewHTTP(mux *http.ServeMux, defaultResponses ...models.ReturnType) HTTP {
	return httpWrapper.NewHTTP(mux, defaultResponses...)
}
