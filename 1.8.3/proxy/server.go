package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	//_ "test/1.7.2/proxy/docs"
	_ "test/docs"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/joho/godotenv"
	"github.com/lestrrat-go/jwx/jwt"
	"golang.org/x/crypto/bcrypt"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func init() {
	godotenv.Load()
}

type Server struct {
	addr   string
	router http.Handler
	users  map[string]string
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

type RequestRegisterLogin struct {
	Login    string `json:"login" example:"flip"`
	Password string `json:"password" example:"flop"`
}

type ResponseRegister struct {
	ID string `json:"id" example:"999"`
}

type ResponseLogin struct {
	Token string `json:"token" example:"qpJhbGciOiJIUzI1NiIsInR5cCI6IlkXVCJ9.kaJsb2dpbiI6ImZsaXAifQ.N2Ycrfyww7I46L51y0MlofV2ef2iBVfsZaQ6J8EgOfk"`
}

type errorResponse struct {
	Message string `json:"message" example:"no token found"`
}

func NewErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	errResponse := errorResponse{
		Message: message,
	}
	jsonResponse, _ := json.Marshal(errResponse)
	w.Write(jsonResponse)
}

func NewServer(addr string, hostProxy string, portProxy string) *Server {
	server := &Server{
		addr:   addr,
		quitch: make(chan struct{}),
		users:  make(map[string]string),
	}

	// Инициализируем маршруты
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	rp := NewReverseProxy(hostProxy, portProxy)
	r.Use(rp.ReverseProxy)

	r.Post("/api/register", server.register)
	r.Post("/api/login", server.login)

	tokenAuth := jwtauth.New("HS256", []byte("gunmode"), nil)
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(customAuthenticator)

		r.Route("/api/address", func(r chi.Router) {
			r.Post("/search", search)
			r.Post("/geocode", geocode)
		})
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		//httpSwagger.URL("http://localhost:8000/swagger/doc.json"), //The url pointing to API definition
		//httpSwagger.URL(fmt.Sprintf("%v/swagger/doc.json", addr)),
		httpSwagger.URL("doc.json"),
	))

	server.router = r
	return server
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

func customAuthenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())

		if err != nil {
			NewErrorResponse(w, http.StatusForbidden, err.Error())
			//http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		if token == nil || jwt.Validate(token) != nil {
			NewErrorResponse(w, http.StatusForbidden, http.StatusText(http.StatusForbidden))
			//http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		// Token is authenticated, pass it through
		next.ServeHTTP(w, r)
	})
}

// @Summary Геокодирование (координаты по адресу)
// @Security ApiKeyAuth
// @Tags api
// @Accept json
// @Produce json
// @Param query body RequestAddressSearch true "Адрес"
// @Success 200 {object} ResponseAddress
// @Failure 400,403 {object} errorResponse
// @Router /api/address/search [post]
func search(w http.ResponseWriter, r *http.Request) {
	var body RequestAddressSearch

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		/* w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error())) */
		return
	}

	//query = body.Query
	gs := NewGeoServiceProxy()
	resp, err := gs.AddressSearch(body.Query)
	if err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
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
// @Security ApiKeyAuth
// @Tags api
// @Accept json
// @Produce json
// @Param lat,lng body RequestAddressGeocode true "Координаты. lat - Географическая широта. lng - Географическая долгота."
// @Success 200 {object} ResponseAddress
// @Failure 400,403 {object} errorResponse
// @Router /api/address/geocode [post]
func geocode(w http.ResponseWriter, r *http.Request) {
	var body RequestAddressGeocode

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	gs := NewGeoServiceProxy()
	resp, err := gs.GeoCode(body.Lat, body.Lng)
	if err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var respAdresses ResponseAddress
	respAdresses.Addresses = resp

	respJSON, _ := json.Marshal(respAdresses)

	responseString := string(respJSON)
	fmt.Fprint(w, responseString)
}

// @Summary Регистрация пользователя
// @Tags api
// @Accept json
// @Produce json
// @Param login,password body RequestRegisterLogin true "Учетные данные"
// @Success 200 {object} ResponseRegister
// @Failure 400,403 {object} errorResponse
// @Router /api/register [post]
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var body RequestRegisterLogin

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 14)
	if err != nil {
		NewErrorResponse(w, http.StatusBadRequest, err.Error())
		/* errResponse := errorResponse{
			Message: err.Error(),
		}
		jsonResponse, _ := json.Marshal(errResponse)
		w.Write(jsonResponse) */
		return
	}

	s.users[body.Login] = string(hash)

	rand.Seed(time.Now().UnixNano())
	id := rand.Intn(100)

	responseRegister := ResponseRegister{ID: fmt.Sprint(id)}
	jsonResp, _ := json.Marshal(responseRegister)

	responseString := string(jsonResp)
	fmt.Fprint(w, responseString)
}

// @Summary Авторизация пользователя
// @Tags api
// @Accept json
// @Produce json
// @Param login,password body RequestRegisterLogin true "Учетные данные"
// @Success 200 {object} ResponseLogin
// @Failure 400,403 {object} errorResponse
// @Router /api/login [post]
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var body RequestRegisterLogin

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		//NewErrorResponse(w, http.StatusBadRequest, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		errResponse := errorResponse{
			Message: err.Error(),
		}
		jsonResponse, _ := json.Marshal(errResponse)
		w.Write(jsonResponse)
		return
	}

	if v, ok := s.users[body.Login]; ok {
		err := bcrypt.CompareHashAndPassword([]byte(v), []byte(body.Password))
		if err != nil {
			//NewErrorResponse(w, http.StatusOK, "неверный пароль")
			w.WriteHeader(http.StatusOK)
			errResponse := errorResponse{
				Message: "неверный пароль",
			}
			jsonResponse, _ := json.Marshal(errResponse)
			w.Write(jsonResponse)
			return
		}
	} else {
		//NewErrorResponse(w, http.StatusOK, "пользователь не найден")
		w.WriteHeader(http.StatusOK)
		errResponse := errorResponse{
			Message: "пользователь не найден",
		}
		jsonResponse, _ := json.Marshal(errResponse)
		w.Write(jsonResponse)
		return
	}

	tokenAuth := jwtauth.New("HS256", []byte("gunmode"), nil)
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"login": body.Login})

	responseLogin := ResponseLogin{Token: tokenString}
	jsonResp, _ := json.Marshal(responseLogin)

	responseString := string(jsonResp)
	fmt.Fprint(w, responseString)
}
