package middleware

import (
	"net/http"

	"karma8/internal/lib/context"

	"github.com/google/uuid"
)

const (
	HeaderRequestID = "request-id"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(HeaderRequestID)
		if rid == "" {
			newRequestID, err := uuid.NewUUID()
			if err == nil {
				rid = newRequestID.String()
				r.Header.Set(HeaderRequestID, rid)
			}
		}

		if rid != "" {
			ctx := context.WithRequestID(r.Context(), rid)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(rw, r)
	})
}
