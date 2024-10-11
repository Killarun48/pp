package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "app/docs"
	"app/internal/infrastructure/db"
	"app/internal/infrastructure/responder"
	"app/internal/modules"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Server struct {
	srv     *http.Server
	users   map[string]string
	sigChan chan os.Signal
}

func NewServer(addr string) *Server {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
		return nil
	}

	server := &Server{
		sigChan: make(chan os.Signal, 1),
		users:   make(map[string]string),
	}

	signal.Notify(server.sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// Инициализируем маршруты
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	//bd, err := db.NewDataBaseSqlite()
	bd, err := db.NewDataBasePostgres()
	if err != nil {
		log.Fatal(err)
		return nil
	}

	err = bd.Migrate()
	if err != nil {
		log.Fatal(err)
		return nil
	}

	repositories := modules.NewRepository(bd.DB)

	services := modules.NewService(repositories)
	respond := responder.NewResponder()

	c := modules.NewController(services, respond)

	r.Route("/api", func(r chi.Router) {
		r.Mount("/", c.InitRoutesUser())
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
	))

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	server.srv = srv
	
	time.Sleep(1 * time.Second)
	
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
