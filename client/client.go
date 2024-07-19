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
	/*
	  TestCase-01- Adicionando ao context da API em MiliSecond, retorna o valor esperado.
	  TestCase-02- Adicionando ao context da API em NanoSecond, retorna o valor esperado - "context deadline exceeded"
	*/

	//TestCase01 - 200Ms
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)

	//TestCase02 - 200Ns
	//ctx, cancel := context.WithTimeout(context.Background(), 200*time.Nanosecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatalf("Erro ao criar a requisição: %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Erro ao fazer a requisição: %v", err)
	}
	defer res.Body.Close()
	io.Copy(os.Stdout, res.Body)
}
