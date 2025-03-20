package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ==================== MODELOS =====================

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

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ItemCarrinho struct {
	ProductID  int
	Nome       string
	Preco      float64
	Quantidade int
	Tamanho    string
}

type Login struct {
	HashedPassword string
	SessionToken   string
	CSRFToken      string
}

// ==================== VARIÁVEIS GLOBAIS =====================

var (
	produtos  []Produto
	nextID    = 1
	carrinhos = map[string][]ItemCarrinho{} // session_token -> itens
	users     = map[string]Login{}
)

// ==================== MAIN =====================

func main() {
	r := gin.Default()

	// Rotas de Autenticação
	r.POST("/register", register)
	r.POST("/login", login)
	r.POST("/logout", logout)

	// Rota protegida (exemplo)
	r.GET("/protected", protected)

	// Rotas CRUD de Produtos
	r.GET("/produtos", listProdutos)
	r.GET("/produtos/:id", getProduto)
	r.POST("/produtos", createProduto)
	r.PUT("/produtos/:id", updateProduto)
	r.DELETE("/produtos/:id", deleteProduto)

	// Carrinho
	r.POST("/cart/add", addToCart)
	r.GET("/cart/view", viewCart)
	r.POST("/cart/remove", removeFromCart)
	r.POST("/cart/checkout", checkout)

	r.Run(":8080")
}

// ==================== AUTH CONTROLLERS =====================

func register(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if len(username) < 8 || len(password) < 8 {
		c.JSON(http.StatusNotAcceptable, gin.H{"message": "Usuário ou senha inválidos."})
		return
	}

	if _, exists := users[username]; exists {
		c.JSON(http.StatusConflict, gin.H{"message": "Usuário já existente."})
		return
	}

	hashedPassword, _ := hashPassword(password)
	users[username] = Login{
		HashedPassword: hashedPassword,
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Usuário registrado com sucesso!"})
}

func login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	user, exists := users[username]
	if !exists || !checkPasswordHash(password, user.HashedPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Usuário ou senha incorretos."})
		return
	}

	sessionToken := generateToken(32)
	csrfToken := generateToken(32)

	// Cookies
	c.SetCookie("session_token", sessionToken, 86400, "/", "", true, true)
	c.SetCookie("csrf_token", csrfToken, 86400, "/", "", false, true)
	c.SetCookie("username", username, 86400, "/", "", false, true)

	// Atualizar user tokens
	user.SessionToken = sessionToken
	user.CSRFToken = csrfToken
	users[username] = user

	c.JSON(http.StatusOK, gin.H{"message": "Login realizado com sucesso!"})
}

func logout(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil || sessionToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Sem sessão ativa."})
		return
	}

	// Limpar sessão
	for username, user := range users {
		if user.SessionToken == sessionToken {
			user.SessionToken = ""
			user.CSRFToken = ""
			users[username] = user
			break
		}
	}

	// Expira cookies
	c.SetCookie("session_token", "", -1, "/", "", true, true)
	c.SetCookie("csrf_token", "", -1, "/", "", false, true)
	c.SetCookie("username", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Logout realizado com sucesso!"})
}

func protected(c *gin.Context) {
	if err := Authorize(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conteúdo protegido acessado com sucesso!"})
}

// ==================== PRODUTOS CONTROLLERS =====================

func listProdutos(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Status:  "success",
		Message: "Lista de produtos recuperada com sucesso",
		Data:    produtos,
	})
}

func getProduto(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Status: "error", Message: "ID inválido."})
		return
	}

	for _, p := range produtos {
		if p.ProductID == id {
			c.JSON(http.StatusOK, Response{Status: "success", Data: p})
			return
		}
	}

	c.JSON(http.StatusNotFound, Response{Status: "error", Message: "Produto não encontrado."})
}

func createProduto(c *gin.Context) {
	var novo Produto
	if err := c.ShouldBindJSON(&novo); err != nil {
		c.JSON(http.StatusBadRequest, Response{Status: "error", Message: err.Error()})
		return
	}

	// Validação adicional
	if novo.Preco <= 0 {
		c.JSON(http.StatusBadRequest, Response{Status: "error", Message: "O preço deve ser positivo."})
		return
	}
	if novo.QtdEstoque < 0 {
		c.JSON(http.StatusBadRequest, Response{Status: "error", Message: "A quantidade em estoque não pode ser negativa."})
		return
	}

	novo.ProductID = nextID
	novo.CriadoEm = time.Now()
	nextID++
	produtos = append(produtos, novo)

	c.JSON(http.StatusCreated, Response{Status: "success", Data: novo})
}

func updateProduto(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Status: "error", Message: "ID inválido."})
		return
	}

	var atualizado Produto
	if err := c.ShouldBindJSON(&atualizado); err != nil {
		c.JSON(http.StatusBadRequest, Response{Status: "error", Message: err.Error()})
		return
	}

	for i, p := range produtos {
		if p.ProductID == id {
			atualizado.ProductID = id
			atualizado.CriadoEm = p.CriadoEm
			produtos[i] = atualizado
			c.JSON(http.StatusOK, Response{Status: "success", Data: atualizado})
			return
		}
	}

	c.JSON(http.StatusNotFound, Response{Status: "error", Message: "Produto não encontrado."})
}

func deleteProduto(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Status: "error", Message: "ID inválido."})
		return
	}

	for i, p := range produtos {
		if p.ProductID == id {
			produtos = append(produtos[:i], produtos[i+1:]...)
			c.JSON(http.StatusOK, Response{Status: "success", Message: "Produto deletado."})
			return
		}
	}

	c.JSON(http.StatusNotFound, Response{Status: "error", Message: "Produto não encontrado."})
}

// ==================== CARRINHO CONTROLLERS =====================

func addToCart(c *gin.Context) {
	if err := Authorize(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	sessionToken, _ := c.Cookie("session_token")

	productID, _ := strconv.Atoi(c.PostForm("product_id"))
	quantity, _ := strconv.Atoi(c.PostForm("quantity"))
	tamanho := c.PostForm("tamanho")

	if quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Quantidade inválida."})
		return
	}

	var produto Produto
	encontrado := false
	for _, p := range produtos {
		if p.ProductID == productID {
			produto = p
			encontrado = true
			break
		}
	}

	if !encontrado {
		c.JSON(http.StatusNotFound, gin.H{"message": "Produto não encontrado."})
		return
	}

	item := ItemCarrinho{
		ProductID:  produto.ProductID,
		Nome:       produto.Nome,
		Preco:      produto.Preco,
		Quantidade: quantity,
		Tamanho:    tamanho,
	}

	carrinhos[sessionToken] = append(carrinhos[sessionToken], item)

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Produto %s adicionado ao carrinho!", produto.Nome)})
}

func viewCart(c *gin.Context) {
	if err := Authorize(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	sessionToken, _ := c.Cookie("session_token")
	cart := carrinhos[sessionToken]

	if len(cart) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Carrinho vazio."})
		return
	}

	total := 0.0
	for _, item := range cart {
		total += item.Preco * float64(item.Quantidade)
	}

	c.JSON(http.StatusOK, gin.H{
		"itens": cart,
		"total": fmt.Sprintf("R$ %.2f", total),
	})
}

func removeFromCart(c *gin.Context) {
	if err := Authorize(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	sessionToken, _ := c.Cookie("session_token")

	productID, _ := strconv.Atoi(c.PostForm("product_id"))
	cart := carrinhos[sessionToken]

	newCart := []ItemCarrinho{}
	for _, item := range cart {
		if item.ProductID != productID {
			newCart = append(newCart, item)
		}
	}

	carrinhos[sessionToken] = newCart

	c.JSON(http.StatusOK, gin.H{"message": "Produto removido do carrinho."})
}

func checkout(c *gin.Context) {
	if err := Authorize(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	sessionToken, _ := c.Cookie("session_token")
	cart := carrinhos[sessionToken]

	if len(cart) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Carrinho vazio."})
		return
	}

	total := 0.0
	for _, item := range cart {
		total += item.Preco * float64(item.Quantidade)
	}

	delete(carrinhos, sessionToken)

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Compra finalizada! Total: R$ %.2f", total)})
}

// ==================== MIDDLEWARES E UTILS =====================

var AuthError = errors.New("Não autorizado.")

func Authorize(c *gin.Context) error {
	username, err := c.Cookie("username")
	if err != nil {
		return AuthError
	}

	user, ok := users[username]
	if !ok {
		return AuthError
	}

	st, err := c.Cookie("session_token")
	if err != nil || st != user.SessionToken {
		return AuthError
	}

	csrf := c.GetHeader("X-CSRF-Token")
	if csrf != user.CSRFToken || csrf == "" {
		return AuthError
	}

	return nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Erro ao gerar token: %v", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
