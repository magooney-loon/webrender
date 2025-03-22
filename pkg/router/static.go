package router

import (
	"net/http"
	"path/filepath"
)

// RegisterStaticHandler sets up a static file server for the specified directory
// This is compatible with Gorilla Mux, unlike the version in the template package
func (r *Router) RegisterStaticHandler(rootDir string, urlPrefix string) {
	// Ensure directory path is properly formatted
	rootDir = filepath.Clean(rootDir)

	// Create a file server handler
	fileServer := http.FileServer(http.Dir(rootDir))

	// Register the handler with the router
	// Using PathPrefix allows handling of all files under the static directory
	r.PathPrefix(urlPrefix + "/").Handler(http.StripPrefix(urlPrefix, fileServer))
}
