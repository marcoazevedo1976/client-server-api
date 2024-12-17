package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	url := "http://localhost:8080/cotacao"

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	cotacaoJson, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var cotacao Cotacao
	err = json.Unmarshal(cotacaoJson, &cotacao)
	if err != nil {
		log.Fatalln(err)
	}

	conteudoArquivo := fmt.Sprintf("Dólar:{%s}", cotacao.Bid)
	err = os.WriteFile("cotacao.txt", []byte(conteudoArquivo), 0644)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Cotação Gerada com Sucesso.")
}
