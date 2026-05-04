package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	viaCEPBaseURL     = "https://viacep.com.br/ws"
	weatherAPIBaseURL = "http://api.weatherapi.com/v1"
)

type WeatherResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type weatherAPIResp struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

func main() {
	port := getenv("PORT", "8080")
	http.HandleFunc("/", handleCEP)
	log.Printf("Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleCEP(w http.ResponseWriter, r *http.Request) {
	cep := strings.TrimPrefix(r.URL.Path, "/")

	if len(cep) != 8 || !isAllDigits(cep) {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	city, err := fetchCity(cep)
	if err != nil {
		if err.Error() == "not found" {
			http.Error(w, "can not find zipcode", http.StatusNotFound)
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	tempC, err := fetchTemperature(city)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := WeatherResponse{
		TempC: tempC,
		TempF: celsiusToFahrenheit(tempC),
		TempK: celsiusToKelvin(tempC),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func fetchCity(cep string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s/json/", viaCEPBaseURL, cep))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return "", fmt.Errorf("not found")
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if _, hasErr := result["erro"]; hasErr {
		return "", fmt.Errorf("not found")
	}

	localidade, ok := result["localidade"].(string)
	if !ok || localidade == "" {
		return "", fmt.Errorf("not found")
	}

	return localidade, nil
}

func fetchTemperature(city string) (float64, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	apiURL := fmt.Sprintf("%s/current.json?key=%s&q=%s&aqi=no",
		weatherAPIBaseURL, apiKey, url.QueryEscape(city))

	resp, err := http.Get(apiURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("weather API returned %d", resp.StatusCode)
	}

	var wr weatherAPIResp
	if err := json.NewDecoder(resp.Body).Decode(&wr); err != nil {
		return 0, err
	}

	return wr.Current.TempC, nil
}

func celsiusToFahrenheit(c float64) float64 {
	return c*1.8 + 32
}

func celsiusToKelvin(c float64) float64 {
	return c + 273
}

func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
