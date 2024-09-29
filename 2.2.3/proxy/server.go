package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"test/controller"
	_ "test/docs"
	"test/geoservice"
	"test/responder"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/ptflp/godecoder"
	"go.uber.org/zap"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Server struct {
	srv     *http.Server
	users   map[string]string
	sigChan chan os.Signal
}

func NewServer(addr string, hostProxy string, portProxy string) *Server {
	server := &Server{
		sigChan: make(chan os.Signal, 1),
		users:   make(map[string]string),
	}

	signal.Notify(server.sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// Инициализируем маршруты
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	rp := NewReverseProxy(hostProxy, portProxy)
	r.Use(rp.ReverseProxy)

	// Инициализируем контроллер
	logger, _ := zap.NewDevelopment()
	responder := responder.NewResponder(godecoder.NewDecoder(), logger)
	geoService := geoservice.NewGeoServiceProxy()
	
	c := controller.NewController(
		controller.WithResponder(responder),
		controller.WithGeoService(geoService),
	)

	r.Group(func(r chi.Router) {

		r.Route("/api/address", func(r chi.Router) {
			r.Post("/search", c.Search)
			r.Post("/geocode", c.Geocode)
		})
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
	))

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	server.srv = srv
	return server
}

func (s *Server) Serve() {
	go func() {
		log.Println("Starting server...")
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-s.sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	log.Println("Server stopped gracefully")
}

func (s *Server) Stop() {
	s.sigChan <- syscall.Signal(1)
}