package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Struct Produto ajustada com campo opcional CriadoEm
type Produto struct {
	ProductID  int       `json:"product_id"`
	Nome       string    `json:"nome"`
	Descricao  string    `json:"descricao"`
	Preco      float64   `json:"preco"`
	Imagens    []string  `json:"imagens"`
	Tamanhos   []string  `json:"tamanhos"`
	Marca      string    `json:"marca"`
	Categoria  string    `json:"categoria"`
	QtdEstoque int       `json:"qtd_estoque"`
	CriadoEm   time.Time `json:"created_at,omitempty"`
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

	http.HandleFunc("/cart/add", addToCart)
	http.HandleFunc("/cart/view", viewCart)
	http.HandleFunc("/cart/remove", removeFromCart)
	http.HandleFunc("/cart/checkout", checkout)

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
	novoProduto.CriadoEm = time.Now() // Adiciona timestamp de criação
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
			produtoAtualizado.CriadoEm = p.CriadoEm // Mantém o timestamp original
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

// Carrinho de Compras
type ItemCarrinho struct {
	ProductID  int
	Nome       string
	Preço      float64
	Quantidade int
	Tamanho    string
}

var carrinhos = map[string][]ItemCarrinho{} // Chaveada pelo session_token

// Adicionar ao Carrinho
func addToCart(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_token")
	if err != nil || sessionCookie.Value == "" {
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	sessionToken := sessionCookie.Value

	// Recebendo dados do produto
	productIDStr := r.FormValue("product_id")
	quantityStr := r.FormValue("quantity")
	tamanho := r.FormValue("tamanho")

	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "ID do produto inválido", http.StatusBadRequest)
		return
	}

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil || quantity <= 0 {
		http.Error(w, "Quantidade inválida", http.StatusBadRequest)
		return
	}

	// Buscar o produto no slice de produtos
	var produto Produto
	encontrado := false
	for _, p := range produtos { // produtos o slice retornado do /produtos
		if p.ProductID == productID {
			produto = p
			encontrado = true
			break
		}
	}

	if !encontrado {
		http.Error(w, "Produto não encontrado", http.StatusNotFound)
		return
	}

	item := ItemCarrinho{
		ProductID:  produto.ProductID,
		Nome:       produto.Nome,
		Preço:      produto.Preco,
		Quantidade: quantity,
		Tamanho:    tamanho,
	}

	// Adiciona ao carrinho do usuário
	carrinhos[sessionToken] = append(carrinhos[sessionToken], item)

	fmt.Fprintf(w, "Produto %s adicionado ao carrinho!\n", produto.Nome)
}

// Ver carrinho
func viewCart(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_token")
	if err != nil || sessionCookie.Value == "" {
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	sessionToken := sessionCookie.Value

	cart, ok := carrinhos[sessionToken]
	if !ok || len(cart) == 0 {
		fmt.Fprintln(w, "Seu carrinho está vazio.")
		return
	}

	total := 0.0
	for _, item := range cart {
		fmt.Fprintf(w, "Produto: %s | Quantidade: %d | Tamanho: %s | Preço Unitário: %.2f\n", item.Nome, item.Quantidade, item.Tamanho, item.Preço)
		total += item.Preço * float64(item.Quantidade)
	}

	fmt.Fprintf(w, "\nTotal: R$ %.2f", total)
}

// Remover produto do carrinho
func removeFromCart(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_token")
	if err != nil || sessionCookie.Value == "" {
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	sessionToken := sessionCookie.Value

	productIDStr := r.FormValue("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "ID do produto inválido", http.StatusBadRequest)
		return
	}

	cart, ok := carrinhos[sessionToken]
	if !ok || len(cart) == 0 {
		http.Error(w, "Carrinho vazio", http.StatusBadRequest)
		return
	}

	newCart := []ItemCarrinho{}
	for _, item := range cart {
		if item.ProductID != productID {
			newCart = append(newCart, item)
		}
	}

	carrinhos[sessionToken] = newCart

	fmt.Fprintln(w, "Produto removido do carrinho.")
}

// Finalizar compra
func checkout(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_token")
	if err != nil || sessionCookie.Value == "" {
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	sessionToken := sessionCookie.Value
	cart, ok := carrinhos[sessionToken]
	if !ok || len(cart) == 0 {
		http.Error(w, "Carrinho vazio", http.StatusBadRequest)
		return
	}

	total := 0.0
	for _, item := range cart {
		total += item.Preço * float64(item.Quantidade)
	}

	// Aqui faremos a lógica de pagamento
	delete(carrinhos, sessionToken) // Limpa o carrinho

	fmt.Fprintf(w, "Compra finalizada com sucesso! Total: R$ %.2f\n", total)
}
