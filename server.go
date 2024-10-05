package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExchangeRate struct {
	Bid string `json:"bid"`
}

func main() {
	db, err := sql.Open("sqlite3", "./exchange_rates.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS exchange_rate (id INTEGER PRIMARY KEY, bid TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancel()

		rate, err := fetchExchangeRate(ctx)
		if err != nil {
			log.Println("Error fetching exchange rate:", err)
			http.Error(w, "Error fetching exchange rate", http.StatusInternalServerError)
			return
		}

		ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancelDB()

		if err := saveExchangeRate(ctxDB, db, rate.Bid); err != nil {
			log.Println("Error saving exchange rate:", err)
		}

		json.NewEncoder(w).Encode(rate)
	})

	log.Println("Server started at http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}

func fetchExchangeRate(ctx context.Context) (*ExchangeRate, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]ExchangeRate
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	rate := result["USDBRL"]
	return &rate, nil
}

func saveExchangeRate(ctx context.Context, db *sql.DB, bid string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := db.ExecContext(ctx, "INSERT INTO exchange_rate (bid) VALUES (?)", bid)
		return err
	}
}
