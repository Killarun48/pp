package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
)

type TestData struct {
	n         string
	expected  string
}

func TestSerever(t *testing.T) {
	go func() {
		log.Fatal(NewServer(":8080", "localhost", "7070").Start())
	}()

	// Запускаем мок-сервер reverseProxy
	go func() {
		log.Fatal(http.ListenAndServe(":7070", nil))
	}()

	testCases := []TestData{
		{n: "api", expected: "Hello from API"},
		{n: "api/h/j", expected: "404 page not found\n"},
		// Проверяем перенаправление на мок-сервер
		{n: "static/1", expected: "404 page not found\n"},
	}

	for _, tc := range testCases {
		httpClient := &http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/%v", tc.n), nil)
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

		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("unexpected result. Error: %v", err)
			return
		}

		result := string(body)
		if result != tc.expected {
			t.Errorf("unexpected result. Input route: %v, Expected: %v, Got: %v", tc.n, tc.expected, result)
		}
	}
}
