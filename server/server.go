package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

type Cotacao struct {
	Dolar USDBRL `json:"USDBRL"`
}

type USDBRL struct {
	Bid string `json:"bid"`
}

func main() {
	db, err := iniciaBD()
	if err != nil {
		log.Fatal(err)
	}

	mux := CriaRotas(db)

	fmt.Println("Ouvindo na porta 8080...")
	log.Fatalln(http.ListenAndServe(":8080", mux))
}

func chamaAPICotacaoDolar() (*Cotacao, error) {
	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cotacaoJson, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cotacao Cotacao
	err = json.Unmarshal(cotacaoJson, &cotacao)
	if err != nil {
		return nil, err
	}

	return &cotacao, nil
}

func persisteDadosBD(db *sql.DB, dolar USDBRL) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := db.ExecContext(ctx, "INSERT INTO cotacao (data_hora, valor) VALUES (?, ?)", time.Now(), dolar.Bid)
	if err != nil {
		return fmt.Errorf("cotação não inserida: %w", err)
	}

	return nil
}

func iniciaBD() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./cotacao.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS cotacao (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		data_hora DATETIME,
		valor text
	)
	`)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func CriaRotas(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		cotacao, err := chamaAPICotacaoDolar()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = persisteDadosBD(db, cotacao.Dolar)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		dolarJson, err := json.Marshal(cotacao.Dolar)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(dolarJson)
	})

	return mux
}
