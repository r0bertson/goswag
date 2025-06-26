package http

import (
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strings"

	"github.com/r0bertson/goswag/internal/generator"
)

// getFuncName retrieves the name of the function associated with the last handler in the given list of http.HandlerFunc.
// It uses the reflect package to obtain the function name from the pointer value of the last handler.
// The function name is extracted by splitting the full function name string using the dot separator and returning the last element.
// The retrieved function name is then returned as a string.
func getFuncName(handlers ...http.HandlerFunc) string {
	lastHandler := handlers[len(handlers)-1]

	fullFuncName := runtime.FuncForPC(reflect.ValueOf(lastHandler).Pointer()).Name()
	funcNameSplit := strings.Split(fullFuncName, ".")
	funcName := funcNameSplit[len(funcNameSplit)-1]
	funcName = strings.TrimSuffix(funcName, "-fm")

	return funcName
}

// toGoSwagRoute converts a slice of httpRoute to a slice of generator.Route.
// It iterates over each httpRoute in the input slice and appends its Route field to the output slice.
// Returns the converted slice of generator.Route.
func toGoSwagRoute(from []*httpRoute) []generator.Route {
	var routes []generator.Route
	for _, r := range from {
		routes = append(routes, r.Route)
	}

	return routes
}

// toGoSwagGroup converts a slice of httpGroup objects to a slice of generator.Group.
// It iterates over each httpGroup and creates a generator.Group object with the corresponding properties.
// The converted generator.Group objects are then returned as a slice.
func toGoSwagGroup(from []*httpGroup) []generator.Group {
	var groups []generator.Group
	for _, g := range from {
		groups = append(groups, generator.Group{
			GroupName: g.groupName,
			Routes:    toGoSwagRoute(g.routes),
		})
	}

	return groups
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
