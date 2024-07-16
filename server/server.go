package main

import (
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
	log.Println("Servidor iniciado na porta 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Falha ao iniciar o servidor: %v", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Println("Request Iniciada")

	defer log.Println("Request Finalizada")
	select {
	case <-time.After(5 * time.Second):

		dolar, err := GetCotacao() // Chamada de função
		if err != nil {
			log.Printf("Falha ao obter cotação: %v\n", err)
			http.Error(w, "Falha ao obter dados", http.StatusInternalServerError)
			return
		}

		if dolar == nil {
			log.Println("Dados do dólar são nulos")
			http.Error(w, "Dados do dólar são nulos", http.StatusInternalServerError)
			return
		}

		log.Printf("Cotação do dólar recebida: %+v\n", dolar)

		cotacao := Cotacao{
			Valor: dolar.Bid,
			Data:  time.Now(),
		}

		if err := db.Create(&cotacao).Error; err != nil {
			log.Printf("Erro ao salvar informações no DB: %v\n", err)
			http.Error(w, "Erro ao salvar informações no DB", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Request processada com sucesso. Valor do Dolar " + dolar.Bid))

	case <-ctx.Done():
		log.Println("Request cancelada pelo usuário")
	}
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
		log.Printf("Erro ao ler o corpo da resposta: %v\n", err)
		return nil, err
	}
	log.Printf("Resposta da API: %s\n", string(body))

	var d ResponserDolar
	if err := json.Unmarshal(body, &d); err != nil {
		log.Printf("Erro ao deserializar JSON: %v\n", err)
		return nil, err
	}
	if d.USDBRL.Bid == "" {
		log.Println("Dados da cotação do dólar são inválidos")
		return nil, fmt.Errorf("Valor do Dolar esta Nulo")
	}
	return &d.USDBRL, nil
}
