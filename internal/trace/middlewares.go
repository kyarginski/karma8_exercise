package trace

import (
	"net/http"
)

func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// трассировка
		if UseTracing {
			context, span := CreateMasterSpan(r)
			if span != nil {
				defer span.End()
			}
			req := r.WithContext(context)
			next.ServeHTTP(w, req)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

/*
Using:

	router.HandleFunc("/api/file", tracing.TracingMiddleware2(handler.PutFileItem(srv))).Methods("PUT")
*/
func TracingMiddleware2(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// трассировка
		if UseTracing {
			context, span := CreateMasterSpan(r)
			if span != nil {
				defer span.End()
			}
			req := r.WithContext(context)
			next.ServeHTTP(w, req)
		} else {
			next.ServeHTTP(w, r)
		}
	}
}
