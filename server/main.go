package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Cotacao representa a estrutura dos dados de cotação recebidos do servidor
type Cotacao struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/cotacao", BuscaCotacaoHandler)
	http.ListenAndServe(":8080", nil)
}

func BuscaCotacaoHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
	defer cancel()

	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Chamando a API para buscar cotação
	cotacao, err := BuscaCotacao(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Salva a cotação no banco de dados SQLite
	ctxDB, cancel := context.WithTimeout(r.Context(), time.Millisecond*10)
	defer cancel()
	err = saveCotacaoToDB(ctxDB, *cotacao)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao salvar cotação no banco de dados: %v", err), http.StatusInternalServerError)
		return
	}

	// Retorna a cotação para o cliente
	response := map[string]string{"bid": cotacao.USDBRL.Bid}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao criar resposta JSON: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func BuscaCotacao(ctx context.Context) (*Cotacao, error) {
	resp, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cotacao Cotacao
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		return nil, err
	}
	return &cotacao, nil
}

func saveCotacaoToDB(ctx context.Context, cotacao Cotacao) error {
	// Abre uma conexão com o banco de dados SQLite (certifique-se de ter o driver mattn/go-sqlite3 instalado)
	db, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		return err
	}
	defer db.Close()

	// Cria a tabela se não existir
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS cotacoes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			valor REAL,
			data_criacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Insere a cotação no banco de dados
	_, err = db.ExecContext(ctx, "INSERT INTO cotacoes (valor) VALUES (?)", cotacao.USDBRL.Bid)
	if err != nil {
		return err
	}

	return nil
}
