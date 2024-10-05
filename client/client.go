package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type ExchangeRate struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	rate, err := fetchExchangeRate(ctx)
	if err != nil {
		log.Println("Error fetching exchange rate:", err)
		return
	}

	if err := saveExchangeRate(rate.Bid); err != nil {
		log.Println("Error saving exchange rate:", err)
	}
}

func fetchExchangeRate(ctx context.Context) (*ExchangeRate, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:8080/cotacao", nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var rate ExchangeRate
	if err := json.NewDecoder(res.Body).Decode(&rate); err != nil {
		return nil, err
	}

	return &rate, nil
}

func saveExchangeRate(value string) error {
	content := fmt.Sprintf("Dollar: %s", value)
	return os.WriteFile("rate.txt", []byte(content), 0o644)
}
