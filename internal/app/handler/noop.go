package handler

import "net/http"

// noopMiddleware is to be returned on error during middleware init.
var noopMiddleware = func(h http.Handler) http.Handler {
	return h
}
