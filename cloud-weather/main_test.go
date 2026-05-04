package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCelsiusToFahrenheit(t *testing.T) {
	cases := []struct{ c, want float64 }{
		{0, 32},
		{100, 212},
		{-40, -40},
	}
	for _, tc := range cases {
		got := celsiusToFahrenheit(tc.c)
		if got != tc.want {
			t.Errorf("celsiusToFahrenheit(%v) = %v, want %v", tc.c, got, tc.want)
		}
	}
}

func TestCelsiusToKelvin(t *testing.T) {
	cases := []struct{ c, want float64 }{
		{0, 273},
		{100, 373},
		{27, 300},
	}
	for _, tc := range cases {
		got := celsiusToKelvin(tc.c)
		if got != tc.want {
			t.Errorf("celsiusToKelvin(%v) = %v, want %v", tc.c, got, tc.want)
		}
	}
}

func TestHandleCEP_InvalidFormat(t *testing.T) {
	cases := []string{"123", "1234567a", "123456789", ""}
	for _, cep := range cases {
		req := httptest.NewRequest(http.MethodGet, "/"+cep, nil)
		w := httptest.NewRecorder()
		handleCEP(w, req)
		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("cep %q: expected 422, got %d", cep, w.Code)
		}
	}
}

func TestHandleCEP_NotFound(t *testing.T) {
	mockViaCEP := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"erro": true}`))
	}))
	defer mockViaCEP.Close()

	orig := viaCEPBaseURL
	viaCEPBaseURL = mockViaCEP.URL
	defer func() { viaCEPBaseURL = orig }()

	req := httptest.NewRequest(http.MethodGet, "/99999999", nil)
	w := httptest.NewRecorder()
	handleCEP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandleCEP_Success(t *testing.T) {
	mockViaCEP := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"localidade":"São Paulo"}`))
	}))
	defer mockViaCEP.Close()

	mockWeather := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"current":{"temp_c":25.0}}`))
	}))
	defer mockWeather.Close()

	origViaCEP, origWeather := viaCEPBaseURL, weatherAPIBaseURL
	viaCEPBaseURL = mockViaCEP.URL
	weatherAPIBaseURL = mockWeather.URL
	defer func() {
		viaCEPBaseURL = origViaCEP
		weatherAPIBaseURL = origWeather
	}()

	req := httptest.NewRequest(http.MethodGet, "/01310100", nil)
	w := httptest.NewRecorder()
	handleCEP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp WeatherResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if resp.TempC != 25.0 {
		t.Errorf("TempC: got %v, want 25.0", resp.TempC)
	}
	if resp.TempF != 77.0 {
		t.Errorf("TempF: got %v, want 77.0", resp.TempF)
	}
	if resp.TempK != 298.0 {
		t.Errorf("TempK: got %v, want 298.0", resp.TempK)
	}
}
