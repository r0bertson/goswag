package goswag

import (
	"github.com/gin-gonic/gin"
	ginWrapper "github.com/r0bertson/goswag/internal/frameworks/gin"
	"github.com/r0bertson/goswag/models"
)

type Gin interface {
	models.GinRouter
	models.GinGroup
	GenerateSwagger()
	Gin() *gin.Engine
}

// NewGin returns the interface that wraps the basic Gin methods and add the swagger methods
// defaultResponses is an optional parameter that can be used to set the default responses for all routes
func NewGin(g *gin.Engine, defaultResponses ...models.ReturnType) Gin {
	return ginWrapper.NewGin(g, defaultResponses...)
}
