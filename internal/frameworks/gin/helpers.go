package gin

import (
	"path"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/r0bertson/goswag/internal/generator"
)

// getFuncName retrieves the name of the function associated with the last handler in the given list of gin.HandlerFunc.
// It uses the reflect package to obtain the function name from the pointer value of the last handler.
// The function name is extracted by splitting the full function name string using the dot separator and returning the last element.
// The retrieved function name is then returned as a string.
// If no handlers are provided, it returns "handler" as a default function name.
func getFuncName(handlers ...gin.HandlerFunc) string {
	if len(handlers) == 0 {
		return "handler"
	}

	lastHandler := handlers[len(handlers)-1]
	if lastHandler == nil {
		return "handler"
	}

	fullFuncName := runtime.FuncForPC(reflect.ValueOf(lastHandler).Pointer()).Name()
	if fullFuncName == "" {
		return "handler"
	}

	funcNameSplit := strings.Split(fullFuncName, ".")
	if len(funcNameSplit) == 0 {
		return "handler"
	}

	funcName := funcNameSplit[len(funcNameSplit)-1]
	funcName = strings.TrimSuffix(funcName, "-fm")

	if funcName == "" {
		return "handler"
	}

	return funcName
}

// toGoSwagRoute converts a slice of ginRoute to a slice of generator.Route.
// It iterates over each ginRoute in the input slice and appends its Route field to the output slice.
// Returns the converted slice of generator.Route.
func toGoSwagRoute(from []*ginRoute) []generator.Route {
	var routes []generator.Route
	for _, r := range from {
		routes = append(routes, r.Route)
	}

	return routes
}

// toGoSwagGroup converts a slice of ginGroup objects to a slice of generator.Group.
// It iterates over each ginGroup and creates a generator.Group object with the corresponding properties.
// The converted generator.Group objects are then returned as a slice.
func toGoSwagGroup(from []*ginGroup) []generator.Group {
	var result []generator.Group
	for _, g := range from {
		var routes []generator.Route
		for _, r := range g.routes {
			routes = append(routes, r.Route)
		}

		var groups []generator.Group
		if g.groups != nil {
			groups = toGoSwagGroup(g.groups)
		}

		result = append(result, generator.Group{
			GroupName:      g.groupName,
			Routes:         routes,
			Groups:         groups,
			TagDescription: g.tagDescription,
			TagExternalDocs: g.tagExternalDocs,
		})
	}
	return result
}

func getFullPath(groupName, relativePath string) string {
	if groupName == "" {
		return relativePath
	}

	fullPath := path.Join(groupName, relativePath)

	if strings.HasSuffix(relativePath, "/") {
		fullPath += "/"
	}

	return fullPath
}
