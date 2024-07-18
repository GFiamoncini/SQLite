package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	//_ "github.com/go-sql-driver/mysql"
	//"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Necessário criar Cabeçalho
type ResponserDolar struct {
	USDBRL Dolar `json:"USDBRL"`
}

// Struct conforme JSON Dolar
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

// Tabela cotacao
type Cotacao struct {
	ID    int       `gorm:"primaryKey"`
	Valor string    `json:"valor"`
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

	http.HandleFunc("/", handler)
	log.Println("Servidor iniciado na porta:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Falha ao iniciar o servidor: %v", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Println("Request Iniciada")

	done := make(chan struct{})
	errC := make(chan error, 1)

	go func() {
		defer close(done)
		defer close(errC)
		dolar, err := GetCotacao() // Chamada de função
		if err != nil {
			log.Printf("Falha ao obter dados: %v\n", err)
			http.Error(w, "Falha ao obter dados", http.StatusInternalServerError)
			errC <- err
			return
		}

		if dolar == nil {
			log.Println("Dados do dólar são nulos")
			http.Error(w, "Dados do dólar são nulos", http.StatusInternalServerError)
			errC <- fmt.Errorf("Dados Inválidos")
			return
		}

		log.Printf("Cotação do dólar recebida: %v\n", dolar.Bid)
		cotacao := Cotacao{
			Valor: dolar.Bid,
			Data:  time.Now(),
		}
		//Ctx para o banco de dados.
		dbctx, dbcancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer dbcancel()
		if err := dbstresstest(dbctx, cotacao); err != nil {
			log.Printf("Erro ao salvar informações no DB: %v\n", err)
			http.Error(w, "Erro ao salvar informações no DB", http.StatusInternalServerError)
			errC <- err
			return
		}
		w.Write([]byte("Request processada com sucesso.\nCotação do Dolar atual: $" + dolar.Bid))
	}()
	//defer log.Println("Request Finalizada")
	select {
	case <-ctx.Done():
		log.Println("Request Cancelada pelo Usuário!")
		http.Error(w, "Request Cancelada pelo Usuário!", http.StatusRequestTimeout)
	case err := <-errC:
		if err != nil {
			http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		}
	case <-done:
		log.Println("Request processada com sucesso!")
	}
	// select {
	// case <-time.After(5 * time.Second):

	// case <-ctx.Done():
	// 	log.Println("Request Cancelada pelo Usuário !")
	// }
}

func GetCotacao() (*Dolar, error) {
	resp, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if err != nil {
		log.Printf("Erro ao fazer request para a API: %v\n", err)
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Erro ao fechar o corpo da resposta: %v\n", cerr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro de leitura do body da resposta: %v\n", err)
		return nil, err
	}
	log.Printf("Retorno API: %s\n", string(body))

	var d ResponserDolar
	if err := json.Unmarshal(body, &d); err != nil {
		log.Printf("Erro ao deserializar JSON: %v\n", err)
		return nil, err
	}
	if d.USDBRL.Bid == "" {
		log.Println("Dados da cotação do dólar são Incorretos")
		return nil, fmt.Errorf("Valor do Dolar vazio.")
	}
	return &d.USDBRL, nil
}

func dbstresstest(ctx context.Context, cotacao Cotacao) error {
	done := make(chan error, 1)
	go func() {
		done <- db.WithContext(ctx).Create(&cotacao).Error
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
