package example

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func getRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", getWelcome)
	return r
}

func getWelcome(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	w.Header().Add("Content-Type", "application/json")
	if id != "" {
		w.WriteHeader(400)
		w.Write([]byte(`{"message": "not found"}`))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(`{"message": "welcome"}`))
}

type OkResponse struct {
	Message string `json:"message"`
}
