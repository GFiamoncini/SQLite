package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatalf("Erro ao Realizar Request: %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Erro ao Prcessar Request %v", err)
	}
	defer res.Body.Close()
	io.Copy(os.Stdout, res.Body)
}
