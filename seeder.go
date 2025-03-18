package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Produto struct {
	ProductID  int      `json:"product_id"`
	Nome       string   `json:"nome"`
	Descricao  string   `json:"descricao"`
	Preco      float64  `json:"preco"`
	Imagens    []string `json:"imagens"`
	Tamanhos   []string `json:"tamanhos"`
	Marca      string   `json:"marca"`
	Categoria  string   `json:"categoria"`
	QtdEstoque int      `json:"qtd_estoque"`
}

func main() {
	produtos := []Produto{
		{
			Nome:       "Camiseta Street Urban",
			Descricao:  "Camiseta 100% algodão, estampa exclusiva de grafite urbano.",
			Preco:      89.90,
			Imagens:    []string{"https://cdn.exemplo.com/produtos/camiseta1.jpg"},
			Tamanhos:   []string{"P", "M", "G", "GG"},
			Marca:      "UrbanWear",
			Categoria:  "Roupas",
			QtdEstoque: 50,
		},
		{
			Nome:       "Tênis Runner X Pro",
			Descricao:  "Tênis leve e resistente, ideal para corridas e uso diário.",
			Preco:      249.99,
			Imagens:    []string{"https://cdn.exemplo.com/produtos/tenis1.jpg"},
			Tamanhos:   []string{"38", "39", "40", "41", "42", "43"},
			Marca:      "SpeedMax",
			Categoria:  "Calçados",
			QtdEstoque: 35,
		},
		{
			Nome:       "Boné Snapback Preto",
			Descricao:  "Boné estilo snapback com ajuste traseiro e bordado frontal minimalista.",
			Preco:      59.90,
			Imagens:    []string{"https://cdn.exemplo.com/produtos/bone1.jpg"},
			Tamanhos:   []string{"Único"},
			Marca:      "FlexFit",
			Categoria:  "Acessórios",
			QtdEstoque: 80,
		},
		{
			Nome:       "Mochila TechGear 25L",
			Descricao:  "Mochila resistente à água com compartimento para notebook até 15 polegadas.",
			Preco:      199.90,
			Imagens:    []string{"https://cdn.exemplo.com/produtos/mochila1.jpg"},
			Tamanhos:   []string{"Único"},
			Marca:      "TechGear",
			Categoria:  "Bolsas e Mochilas",
			QtdEstoque: 20,
		},
		{
			Nome:       "Relógio Digital Pulse",
			Descricao:  "Relógio digital com cronômetro, alarme e iluminação noturna.",
			Preco:      149.90,
			Imagens:    []string{"https://cdn.exemplo.com/produtos/relogio1.jpg"},
			Tamanhos:   []string{"Único"},
			Marca:      "PulseTech",
			Categoria:  "Relógios",
			QtdEstoque: 60,
		},
	}

	url := "http://localhost:8080/produtos" // Atualize se necessário

	for _, p := range produtos {
		jsonData, err := json.Marshal(p)
		if err != nil {
			fmt.Println("Erro ao converter produto para JSON:", err)
			continue
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Erro ao fazer POST do produto:", err)
			continue
		}
		defer resp.Body.Close()

		fmt.Printf("Produto '%s' enviado! Status: %s\n", p.Nome, resp.Status)
	}
}
