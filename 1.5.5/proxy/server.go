package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

type Server struct {
	addr   string
	router http.Handler
}

func NewServer(addr string, hostProxy string, portProxy string) *Server {
	r := chi.NewRouter()
	//r.Use(middleware.Logger)

	rp := NewReverseProxy(hostProxy, portProxy)
	r.Use(rp.ReverseProxy)

	r.Route("/api", func(r chi.Router) {
		r.Get("/", helloFromAPI)
		r.Post("/", helloFromAPI)
		r.Route("/{}", func(r chi.Router) {
			r.Get("/", helloFromAPI)
			r.Post("/", helloFromAPI)
		})
	})

	return &Server{
		addr:   addr,
		router: r,
	}
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.addr, s.router)
}

func helloFromAPI(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from API")
}
