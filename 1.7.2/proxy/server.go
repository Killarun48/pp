package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	//_ "test/1.7.2/proxy/docs"
	_ "test/docs"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func init() {
	godotenv.Load()
}

type Server struct {
	addr   string
	router http.Handler
	quitch chan struct{}
}

type GeoServiceProxy struct {
	geoService *GeoService
}

func NewGeoServiceProxy() *GeoServiceProxy {
	apiKey := os.Getenv("API_KEY_GEO_SERVICE")
	secretKey := os.Getenv("SECRET_KEY_GEO_SERVICE")

	gs := NewGeoService(apiKey, secretKey)

	return &GeoServiceProxy{
		geoService: gs,
	}
}

func (gp *GeoServiceProxy) AddressSearch(input string) ([]*Address, error) {
	return gp.geoService.AddressSearch(input)
}

func (gp *GeoServiceProxy) GeoCode(lat, lng string) ([]*Address, error) {
	return gp.geoService.GeoCode(lat, lng)
}

type RequestAddressSearch struct {
	Query string `json:"query" example:"Московский проспект 14"`
}

type ResponseAddress struct {
	Addresses []*Address `json:"addresses"`
}

type RequestAddressGeocode struct {
	Lat string `json:"lat" example:"59.923013"`
	Lng string `json:"lng" example:"30.318105"`
}

func NewServer(addr string, hostProxy string, portProxy string) *Server {
	// Инициализируем маршруты
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	rp := NewReverseProxy(hostProxy, portProxy)
	r.Use(rp.ReverseProxy)

	r.Route("/api/address", func(r chi.Router) {
		r.Post("/search", search)
		r.Post("/geocode", geocode)
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		//httpSwagger.URL("http://localhost:8000/swagger/doc.json"), //The url pointing to API definition
		httpSwagger.URL(fmt.Sprintf("%v/swagger/doc.json", addr)),
	))

	return &Server{
		addr:   addr,
		router: r,
		quitch: make(chan struct{}),
	}
}

func (s *Server) Start() error {
	e := make(chan error)
	go func() {
		e <- http.ListenAndServe(s.addr, s.router)
	}()

	select {
	case err := <-e:
		return err
	case <-time.After(100 * time.Millisecond):
	}

	<-s.quitch

	return nil
}

func (s *Server) Stop() {
	s.quitch <- struct{}{}
}

// @Summary Геокодирование (координаты по адресу)
// @Tags api
// @Accept json
// @Produce json
// @Param query body RequestAddressSearch true "Адрес"
// @Success 200 {object} ResponseAddress
// @Failure 400 {string} string
// @Router /api/address/search [post]
func search(w http.ResponseWriter, r *http.Request) {
	var body RequestAddressSearch

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	//query = body.Query
	gs := NewGeoServiceProxy()
	resp, err := gs.AddressSearch(body.Query)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var respAdresses ResponseAddress
	respAdresses.Addresses = resp

	respJSON, _ := json.Marshal(respAdresses)
	/* if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} */

	responseString := string(respJSON)
	fmt.Fprint(w, responseString)
}

// @Summary Обратное геокодирование (адрес по координатам)
// @Tags api
// @Accept json
// @Produce json
// @Param lat,lng body RequestAddressGeocode true "Координаты. lat - Географическая широта. lng - Географическая долгота."
// @Success 200 {object} ResponseAddress
// @Failure 400 {string} string
// @Router /api/address/geocode [post]
func geocode(w http.ResponseWriter, r *http.Request) {
	var body RequestAddressGeocode

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	gs := NewGeoServiceProxy()
	resp, err := gs.GeoCode(body.Lat, body.Lng)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var respAdresses ResponseAddress
	respAdresses.Addresses = resp

	respJSON, _ := json.Marshal(respAdresses)
	/* if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} */

	responseString := string(respJSON)
	fmt.Fprint(w, responseString)
}
