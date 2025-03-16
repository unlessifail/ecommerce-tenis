package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Struct Produto ajustada com campo opcional CreatedAt
type Produto struct {
	ProductID  int       `json:"product_id"`
	Nome       string    `json:"nome"`
	Descrição  string    `json:"descricao"`
	Preço      float64   `json:"preco"`
	Imagens    []string  `json:"imagens"`
	Tamanhos   []string  `json:"tamanhos"`
	Marca      string    `json:"marca"`
	Categoria  string    `json:"categoria"`
	QtdEstoque int       `json:"qtd_estoque"`
	CreatedAt  time.Time `json:"created_at,omitempty"` // Opcional, aparece só se preenchido
}

// Struct para respostas padronizadas
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Banco de dados em memória
var produtos []Produto
var nextID = 1

func main() {
	r := gin.Default()

	// Rotas CRUD
	r.GET("/produtos", listProdutos)
	r.GET("/produtos/:id", getProduto)
	r.POST("/produtos", createProduto)
	r.PUT("/produtos/:id", updateProduto)
	r.DELETE("/produtos/:id", deleteProduto)

	r.Run(":8080")
}

// Listar todos os produtos
func listProdutos(c *gin.Context) {
	response := Response{
		Status:  "success",
		Message: "Lista de produtos recuperada com sucesso",
		Data:    produtos,
	}
	c.JSON(http.StatusOK, response)
}

// Obter um produto por ID
func getProduto(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response := Response{
			Status:  "error",
			Message: "ID inválido. Por favor, forneça um número inteiro.",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	for _, p := range produtos {
		if p.ProductID == id {
			response := Response{
				Status:  "success",
				Message: "Produto encontrado",
				Data:    p,
			}
			c.JSON(http.StatusOK, response)
			return
		}
	}

	response := Response{
		Status:  "error",
		Message: "Produto não encontrado com o ID fornecido.",
	}
	c.JSON(http.StatusNotFound, response)
}

// Criar um novo produto
func createProduto(c *gin.Context) {
	var novoProduto Produto
	if err := c.ShouldBindJSON(&novoProduto); err != nil {
		response := Response{
			Status:  "error",
			Message: "Erro ao processar os dados: " + err.Error(),
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	novoProduto.ProductID = nextID
	novoProduto.CreatedAt = time.Now() // Adiciona timestamp de criação
	nextID++
	produtos = append(produtos, novoProduto)

	response := Response{
		Status:  "success",
		Message: "Produto criado com sucesso",
		Data:    novoProduto,
	}
	c.JSON(http.StatusCreated, response)
}

// Atualizar um produto
func updateProduto(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response := Response{
			Status:  "error",
			Message: "ID inválido. Por favor, forneça um número inteiro.",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var produtoAtualizado Produto
	if err := c.ShouldBindJSON(&produtoAtualizado); err != nil {
		response := Response{
			Status:  "error",
			Message: "Erro ao processar os dados: " + err.Error(),
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	for i, p := range produtos {
		if p.ProductID == id {
			produtoAtualizado.ProductID = id
			produtoAtualizado.CreatedAt = p.CreatedAt // Mantém o timestamp original
			produtos[i] = produtoAtualizado
			response := Response{
				Status:  "success",
				Message: "Produto atualizado com sucesso",
				Data:    produtoAtualizado,
			}
			c.JSON(http.StatusOK, response)
			return
		}
	}

	response := Response{
		Status:  "error",
		Message: "Produto não encontrado com o ID fornecido.",
	}
	c.JSON(http.StatusNotFound, response)
}

// Deletar um produto
func deleteProduto(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response := Response{
			Status:  "error",
			Message: "ID inválido. Por favor, forneça um número inteiro.",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	for i, p := range produtos {
		if p.ProductID == id {
			produtos = append(produtos[:i], produtos[i+1:]...)
			response := Response{
				Status:  "success",
				Message: "Produto deletado com sucesso",
				Data:    nil,
			}
			c.JSON(http.StatusOK, response)
			return
		}
	}

	response := Response{
		Status:  "error",
		Message: "Produto não encontrado com o ID fornecido.",
	}
	c.JSON(http.StatusNotFound, response)
}
