package main

import (
	"log"
	"net/http"
)

type Server struct {
	db *DB
}

// recoverMiddleware recovers panics and returns 500 so the client always gets a response.
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic serving %s %s: %v", r.Method, r.URL.Path, rec)
				w.Header().Set("Content-Type", "application/vnd.api+json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"errors":[{"status":"500","title":"Internal Server Error","detail":"An unexpected error occurred"}]}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	db, err := NewDB()
	if err != nil {
		log.Fatal(err)
	}
	srv := &Server{db: db}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /stores", srv.listStores)
	mux.HandleFunc("POST /stores", srv.createStore)
	mux.HandleFunc("GET /stores/{id}", func(w http.ResponseWriter, r *http.Request) {
		srv.getStore(w, r, r.PathValue("id"))
	})
	mux.HandleFunc("PATCH /stores/{id}", func(w http.ResponseWriter, r *http.Request) {
		srv.updateStore(w, r, r.PathValue("id"))
	})
	mux.HandleFunc("DELETE /stores/{id}", func(w http.ResponseWriter, r *http.Request) {
		srv.deleteStore(w, r, r.PathValue("id"))
	})

	mux.HandleFunc("GET /pets", srv.listPets)
	mux.HandleFunc("POST /pets", srv.createPet)
	mux.HandleFunc("GET /pets/{id}", func(w http.ResponseWriter, r *http.Request) {
		srv.getPet(w, r, r.PathValue("id"))
	})
	mux.HandleFunc("PATCH /pets/{id}", func(w http.ResponseWriter, r *http.Request) {
		srv.updatePet(w, r, r.PathValue("id"))
	})
	mux.HandleFunc("DELETE /pets/{id}", func(w http.ResponseWriter, r *http.Request) {
		srv.deletePet(w, r, r.PathValue("id"))
	})

	// Panic recovery then CORS for client example
	handler := cors(recoverMiddleware(mux))
	log.Println("JSON:API server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "http://localhost:8081"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
