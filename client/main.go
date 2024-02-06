package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Cotacao representa a estrutura dos dados de cotação recebidos do servidor
type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		fmt.Println("Erro ao criar requisição:", err)
		return
	}

	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Erro ao fazer requisição:", err)
		return
	}
	defer resp.Body.Close()

	// Verifica se o status da resposta é 200 (OK)
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Código de status inesperado:", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Erro ao ler o corpo da resposta:", err)
		return
	}

	// Converte o JSON para a estrutura Cotacao
	var result Cotacao
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}

	// Converte a cotação para float64
	value, err := strconv.ParseFloat(result.Bid, 32)
	if err != nil {
		fmt.Println("Erro ao extrair o valor de lance do JSON")
		return
	}

	// Cria o arquivo cotacao.txt
	err = os.WriteFile("cotacao.txt", []byte(fmt.Sprintf("Dólar: %.2f\n", value)), 0444)
	if err != nil {
		fmt.Println("Erro ao escrever em cotacao.txt:", err)
		return
	}

	fmt.Println("Cotação do dólar salva com sucesso!")
}
