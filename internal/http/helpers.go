package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/AYaSmyslov/faqapi/internal/service"
)

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})

}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func statusFromError(err error) int {
	switch err {
	case service.ErrBadRequest:
		return http.StatusBadRequest
	case service.ErrNotFound, service.ErrNoSuchQuestion:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("%s %s done in %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func pathInt64(r *http.Request, key string) (int64, error) {
	v := r.PathValue(key)
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid " + key)
	}
	return id, nil
}
