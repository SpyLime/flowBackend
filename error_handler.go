package main

import (
	"net/http"

	openapi "github.com/SpyLime/flowBackend/go"
	"github.com/go-pkgz/lgr"
)

// ErrorHandler handles errors from the OpenAPI implementation
func ErrorHandler(w http.ResponseWriter, r *http.Request, err error, result *openapi.ImplResponse) {
	lgr.Printf("ERROR %s", err)
}
