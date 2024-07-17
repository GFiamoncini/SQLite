Este projetos consiste na aplicação das seguintes ferramentas:
  # Context
    -> Define o tipo de contexto, que carrega prazos, sinais de cancelamento e outros valores com escopo de solicitação através dos limites da API e entre os processos.
  # Gorm
    -> ORM para a linguagem GO

# Estrutura do projeto
  -> client.go
  -> server.go
# Utilizar GORM
  -> Instalar as bibliotecas necessárias para uso do GORM
    comando - **go mod tidy**
# Utilizar MYSQL
  -> Criar arquivo docker-compose com as informações necessárias para subir o DB. 
# Requisitos:
  ## Criar uma struct para armazenar os dados em Json.
    -> Necessário criar outra struct para referenciar o cabeçalho do Json
  - JSON da API
  {
    "USDBRL": {
        "code": "USD",
        "codein": "BRL",
        "name": "Dólar Americano/Real Brasileiro",
        "high": "5.4654",
        "low": "5.4155",
        "varBid": "-0.0086",
        "pctChange": "-0.16",
        "bid": "5.4291",
        "ask": "5.4298",
        "timestamp": "1720817996",
        "create_date": "2024-07-12 17:59:56"
    }
  }

  ## Structs criadas usando alias.
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
  ## Criar uma struct para gravar os dados no banco
   type Cotacao struct {
	    ID    int       `gorm:"primaryKey"`
	    Valor string    `json:"valor"`
	    Data  time.Time `json:"data"`
   }
  # Client
   -> O client.go deverá realizar uma requisição HTTP no server.go solicitando a cotação do dólar.
   -> O endpoint necessário gerado pelo server.go para este desafio será: /cotacao. 
   -> O client.go terá que salvar a cotação atual em um arquivo "cotacao.txt" no formato: Dólar: {valor}
   -> O client.go precisará receber do server.go apenas o valor atual do câmbio (campo "bid" do JSON).
  # Server  
   -> Porta a ser utilizada pelo servidor HTTP será a 8080. 
   -> O server.go deverá consumir a API contendo o câmbio de Dólar e Real no endereço: 
     - https://economia.awesomeapi.com.br/json/last/USD-BRL deverá retornar no formato JSON o resultado para o cliente.
  ## Usando o package "context", o server.go deverá registrar no banco de dados (MySQL) cada cotação recebida
   -> Timeout máximo para chamar a API de cotação do dólar 200ms.
   -> Timeout máximo para conseguir persistir os dados no banco deverá ser de 10ms.
  ## Utilizar pacote "context".
   -> Os 3 contextos deverão retornar erro nos logs caso o tempo de execução seja insuficiente.
   
   
   
 
