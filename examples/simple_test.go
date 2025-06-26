package main

import (
	"net/http"

	"github.com/r0bertson/goswag"
	"github.com/r0bertson/goswag/models"
)

type TestResponse struct {
	Message string `json:"message"`
}

func testSwaggerGeneration() {
	mux := http.NewServeMux()
	swagger := goswag.NewHTTP(mux)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	swagger.GET("/test", handler).
		Summary("Test endpoint").
		Description("A test endpoint").
		Tags("test").
		Returns([]models.ReturnType{
			{
				StatusCode: 200,
				Body:       TestResponse{},
			},
		})

	// Generate swagger documentation
	swagger.GenerateSwagger()

	println("Swagger documentation generated successfully!")
}
