package main

import (
	"fmt"
	"time"
)

// Produto
type Produto struct {
	ProductID  int      `json:"product_id"`
	Nome       string   `json:"nome"`
	Descrição  string   `json:"descricao"`
	Preço      float64  `json:"preco"`
	Imagens    [][]byte `json:"imagens"`
	Tamanhos   []string `json:"tamanhos"`
	Marca      string   `json:"marca"`
	Categoria  string   `json:"categoria"`
	QtdEstoque int      `json:"qtd_estoque"`
}

// Usuario
type Usuario struct {
	UserID           int    `json:"user_id"`
	Nome             string `json:"nome"`
	Email            string `json:"email"`
	Senha            string `json:"senha"`
	Endereco         string `json:"endereco"`
	HistoricoPedidos string `json:"historico_pedidos"`
}

// Pedido
type Pedido struct {
	ID     int       `json:"id"`
	UserID int       `json:"user_id"`
	Itens  int       `json:"itens"`
	Total  float64   `json:"total"`
	Status string    `json:"status"`
	Data   time.Time `json:"data"`
}

func main() {

	// Criando uma instância de Produto
	produto := Produto{
		ProductID:  1,
		Nome:       "Tênis Esportivo",
		Descrição:  "Tênis de corrida confortável e durável.",
		Preço:      299.99,
		Imagens:    [][]byte{[]byte("imagem1.jpg"), []byte("imagem2.jpg")},
		Tamanhos:   []string{"P", "M", "G"},
		Marca:      "Nike",
		Categoria:  "Esportes",
		QtdEstoque: 50,
	}

	// Criando uma instância de Usuario
	usuario := Usuario{
		UserID:           1,
		Nome:             "Jaci Ribeiro",
		Email:            "jaci.r@gmail.com",
		Senha:            "deutchsland",
		Endereco:         "Rua Germânia, 44",
		HistoricoPedidos: "Pedido 1, Pedido 2, Pedido 3",
	}

	// Criando uma instância de Pedido
	pedido := Pedido{
		ID:     1,
		UserID: 1,
		Itens:  3,
		Total:  150.75,
		Status: "Em andamento",
		Data:   time.Now(),
	}

	// Exibindo os dados
	fmt.Println("Produto:", produto)
	fmt.Println("Usuário:", usuario)
}
