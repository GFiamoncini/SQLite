package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ResponserDolar struct {
	USDBRL Dolar `json:"USDBRL"`
}

type Dolar struct {
	Code       string `json:"code"`
	CodeIn     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Time       string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type Cotacao struct {
	ID    int       `gorm:"primaryKey"`
	Valor string    `gorm:"index"`
	Data  time.Time `json:"data"`
}

var db *gorm.DB

func main() {
	var err error

	db, err = gorm.Open(sqlite.Open("dolar.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Falha na conexão com o banco de dados: %v", err)
	}
	log.Println("Conectado ao banco de dados com sucesso")

	err = db.AutoMigrate(&Cotacao{})
	if err != nil {
		log.Fatalf("Erro na migração da tabela %v", err)
	}
	log.Println("Migração com sucesso da tabela")

	http.HandleFunc("/cotacao", handler)
	log.Println("Servidor iniciado na porta:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Falha ao iniciar o servidor: %v", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Println("Request Iniciada")

	start := time.Now()

	/*
		TestCase-01- Adicionando ao context da API em MiliSecond, retorna o valor esperado.
		TestCase-02- Adicionando ao context da API em NanoSecond, retorna o valor esperado - "context deadline exceeded"
	*/

	//TestCase01 - "200Ms"
	apiCtx, apiCancel := context.WithTimeout(ctx, 200*time.Millisecond)

	//TestCase02 - "200Ns"
	//apiCtx, apiCancel := context.WithTimeout(ctx, 200*time.Nanosecond)
	defer apiCancel()

	dolar, err := GetCotacao(apiCtx)
	if err != nil {
		log.Printf("Falha ao obter dados: %v\n", err)
		http.Error(w, "Erro ao obter a cotação do dólar. Timeout", http.StatusInternalServerError)
		return
	}

	if dolar == nil {
		log.Println("Dados do dólar são nulos")
		http.Error(w, "Dados do dólar são nulos", http.StatusInternalServerError)
		return
	}

	log.Printf("Cotação do dólar recebida: %v\n", dolar.Bid)
	cotacao := Cotacao{
		Valor: dolar.Bid,
		Data:  time.Now(),
	}

	/*
		       TestCase-01- Adicionando ao context do DB em MiliSecond, retorna o valor esperado.
			   TestCase-02- Adicionando ao context do DB em NanoSecond, retorna o valor esperado - "context deadline exceeded"
	*/

	//TesteCase01 - 10Ms - DB
	dbCtx, dbCancel := context.WithTimeout(ctx, 10*time.Millisecond)

	//TesteCase02 - 10Ns - DB
	//dbCtx, dbCancel := context.WithTimeout(ctx, 10*time.Nanosecond)

	defer dbCancel()
	if err := dbstresstest(dbCtx, cotacao); err != nil {
		log.Printf("Erro ao salvar informações no DB: %v\n", err)
		http.Error(w, "Erro ao salvar informações no DB. Tente novamente(Timeout).", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Request processada com sucesso.Cotação do Dolar atual: $" + dolar.Bid + "\n"))
	log.Printf("Tempo de Operacao: %v\n", time.Since(start))
}

func GetCotacao(ctx context.Context) (*Dolar, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Erro ao fechar o corpo da resposta: %v\n", cerr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("Retorno API: %s\n", string(body))

	var d ResponserDolar
	if err := json.Unmarshal(body, &d); err != nil {
		return nil, err
	}
	if d.USDBRL.Bid == "" {
		return nil, fmt.Errorf("Valor do Dolar vazio.")
	}
	return &d.USDBRL, nil
}

func dbstresstest(ctx context.Context, cotacao Cotacao) error {
	return db.WithContext(ctx).Create(&cotacao).Error
}
