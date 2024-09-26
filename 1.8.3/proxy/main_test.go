package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
)

type TestData struct {
	n        string
	n1       io.Reader
	expected string
}

func TestServer(t *testing.T) {
	s := NewServer(":8080", "localhost", "7070")
	go func() {
		s.Start()
	}()

	go func() {
		log.Fatal(http.ListenAndServe(":7070", nil))
	}()

	// Запускаем мок-сервер reverseProxy
	/* chStop := make(chan struct{})
	go func() {
		go func() {
			http.ListenAndServe(":7070", nil)
		}()
		<-chStop
	}() */

	requestAddressSearch := RequestAddressSearch{Query: "московский проспект 14"}
	jsonData, _ := json.Marshal(requestAddressSearch)
	var dataRequestAddressSearch = strings.NewReader(string(jsonData))

	requestAddressGeocode := RequestAddressGeocode{Lat: "55.776755", Lng: "37.756263"}
	jsonData, _ = json.Marshal(requestAddressGeocode)
	var dataRequestAddressGeocode = strings.NewReader(string(jsonData))

	requestRegisterLogin := RequestRegisterLogin{Login: "flip", Password: "flop"}
	jsonData, _ = json.Marshal(requestRegisterLogin)
	var dataRequestRegisterLogin = strings.NewReader(string(jsonData))

	var dataRequestRegisterLogin2 = strings.NewReader(string(jsonData))

	requestRegisterLogin_BadLogin := RequestRegisterLogin{Login: "f", Password: "flop"}
	jsonData, _ = json.Marshal(requestRegisterLogin_BadLogin)
	var dataRequestRegisterLogin_BadLogin = strings.NewReader(string(jsonData))

	requestRegisterLogin_BadLogin2 := RequestRegisterLogin{Login: "flip", Password: "f"}
	jsonData, _ = json.Marshal(requestRegisterLogin_BadLogin2)
	var dataRequestRegisterLogin_BadLogin2 = strings.NewReader(string(jsonData))

	testCases := []TestData{
		// Проверяем перенаправление на мок-сервер
		{n: "static/1", expected: "404 Not Found"},

		{n: "api", n1: nil, expected: "404 Not Found"},
		{n: "api/address/search", n1: dataRequestAddressSearch, expected: "200 OK"},
		{n: "api/address/geocode", n1: dataRequestAddressGeocode, expected: "200 OK"},

		{n: "api/address/search", n1: dataRequestAddressGeocode, expected: "400 Bad Request"},
		{n: "api/address/geocode", n1: dataRequestAddressSearch, expected: "400 Bad Request"},

		{n: "api/register", n1: dataRequestRegisterLogin, expected: "200 OK"},
		{n: "api/login", n1: dataRequestRegisterLogin2, expected: "200 OK"},

		{n: "api/login", n1: dataRequestRegisterLogin_BadLogin, expected: "200 OK"},
		{n: "api/login", n1: dataRequestRegisterLogin_BadLogin2, expected: "200 OK"},

		{n: "api/register", n1: dataRequestRegisterLogin, expected: "400 Bad Request"},
		{n: "api/login", n1: dataRequestRegisterLogin_BadLogin2, expected: "400 Bad Request"},
	}

	for _, tc := range testCases {
		httpClient := &http.Client{}
		req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:8080/%v", tc.n), tc.n1)
		if err != nil {
			t.Errorf("unexpected result. Error: %v", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6ImZsaXAifQ.Y9-jQnXwst0qxMVv6gi662IVq_c0F5T_MZOen7SjWG4")

		res, err := httpClient.Do(req)
		if err != nil {
			t.Errorf("unexpected result. Error: %v", err)
			return
		}

		if res.Status != tc.expected {
			t.Errorf("unexpected result. Input route: %v, Expected: %v, Got: %v", tc.n, tc.expected, res.Status)
		}
	}
	//chStop <- struct{}{}
	s.Stop()
}

func TestSereverWithBadGeoService(t *testing.T) {
	// Ломаем geoService
	os.Setenv("API_KEY_GEO_SERVICE", "")
	os.Setenv("SECRET_KEY_GEO_SERVICE", "")

	go func() {
		NewServer(":8080", "localhost", "8080").Start()
	}()

	requestAddressSearch := RequestAddressSearch{Query: "московский проспект 14"}
	jsonData, _ := json.Marshal(requestAddressSearch)
	var dataRequestAddressSearch = strings.NewReader(string(jsonData))

	testCases := []TestData{
		{n: "api/address/search", n1: dataRequestAddressSearch, expected: `{"message":"doRequest: Response not OK: HTTP response: 401 "}`},
		{n: "", n1: nil, expected: "404 page not found\n"},
	}

	for _, tc := range testCases {
		httpClient := &http.Client{}
		req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:8080/%v", tc.n), tc.n1)
		if err != nil {
			t.Errorf("unexpected result. Error: %v", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6ImZsaXAifQ.Y9-jQnXwst0qxMVv6gi662IVq_c0F5T_MZOen7SjWG4")

		res, err := httpClient.Do(req)
		if err != nil {
			t.Errorf("unexpected result. Error: %v", err)
			continue
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("unexpected result. Error: %v", err)
		}

		result := string(body)
		if result != tc.expected {
			t.Errorf("unexpected result. Input route: %v, Expected: %v, Got: %v", tc.n, tc.expected, result)
		}
	}
}
