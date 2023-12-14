package health

import "net/http"

func LivenessHandler(checker LivenessChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if checker.LivenessCheck() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func ReadinessHandler(checker ReadinessChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if checker.ReadinessCheck() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
