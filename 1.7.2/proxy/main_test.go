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

	// Запускаем мок-сервер reverseProxy
	go func() {
		log.Fatal(http.ListenAndServe(":7070", nil))
	}()

	requestAddressSearch := RequestAddressSearch{Query: "московский проспект 14"}
	jsonData, _ := json.Marshal(requestAddressSearch)
	var dataRequestAddressSearch = strings.NewReader(string(jsonData))

	requestAddressGeocode := RequestAddressGeocode{Lat: "55.776755", Lng: "37.756263"}
	jsonData, _ = json.Marshal(requestAddressGeocode)
	var dataRequestAddressGeocode = strings.NewReader(string(jsonData))

	testCases := []TestData{
		// Проверяем перенаправление на мок-сервер
		{n: "static/1", expected: "404 Not Found"},

		{n: "api", n1: nil, expected: "404 Not Found"},
		{n: "api/address/search", n1: dataRequestAddressSearch, expected: "200 OK"},
		{n: "api/address/geocode", n1: dataRequestAddressGeocode, expected: "200 OK"},

		{n: "api/address/search", n1: dataRequestAddressGeocode, expected: "400 Bad Request"},
		{n: "api/address/geocode", n1: dataRequestAddressSearch, expected: "400 Bad Request"},
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

		res, err := httpClient.Do(req)
		if err != nil {
			t.Errorf("unexpected result. Error: %v", err)
			return
		}

		if res.Status != tc.expected {
			t.Errorf("unexpected result. Input route: %v, Expected: %v, Got: %v", tc.n, tc.expected, res.Status)
		}
	}
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

	requestAddressGeocode := RequestAddressGeocode{Lat: "55.776755", Lng: "37.756263"}
	jsonData, _ = json.Marshal(requestAddressGeocode)
	var dataRequestAddressGeocode = strings.NewReader(string(jsonData))

	testCases := []TestData{
		{n: "api/address/search", n1: dataRequestAddressSearch, expected: "doRequest: Response not OK: HTTP response: 401 "},
		{n: "api/address/geocode", n1: dataRequestAddressGeocode, expected: "{\"addresses\":null}"},
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
