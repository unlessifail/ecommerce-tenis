package main

import (
	"fmt"
	"net/http"
	"time"
)

type Login struct {
	HashedPassword string
	SessionToken   string
	CSRFToken      string
}

// Usuário

var users = map[string]Login{}

func main() {
	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/protected", protected)
	http.ListenAndServe(":8080", nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Método Inválido", er)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	if len(username) < 8 || len(password) < 8 {
		er := http.StatusNotAcceptable
		http.Error(w, "Usuário ou senha inválidos", er)
		return
	}

	if _, ok := users[username]; ok {
		er := http.StatusConflict
		http.Error(w, "Usuário existente", er)
		return
	}

	hashedPassword, _ := hashPassword(password)
	users[username] = Login{
		HashedPassword: hashedPassword,
	}

	fmt.Fprintln(w, "Usuário registrado com sucesso!")

}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Método de solicitação inválido.", er)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, ok := users[username]
	if !ok || !checkPasswordHash(password, user.HashedPassword) {
		er := http.StatusUnauthorized
		http.Error(w, "Usuário ou senha incorretos.", er)
		return
	}

	sessionToken := generateToken(32)
	csrfToken := generateToken(32)

	// Define cookie de sessão
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	// Define Token CSRF em um cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	})

	// Armazena tokens no banco de dados
	user.SessionToken = sessionToken
	users[username] = user

	fmt.Fprintln(w, "Login realizado com sucesso!")

}

func protected(w http.ResponseWriter, r *http.Request) {
	// Verifique o token de sessão
	sessionCookie, err := r.Cookie("session_token")
	if err != nil || sessionCookie == nil {
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	// Validar o token da sessão...
	fmt.Fprintln(w, "Conteúdo protegido acessado com sucesso!")
}

func logout(w http.ResponseWriter, r *http.Request) {

	// Adquire o cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Sem sessão ativa.", http.StatusUnauthorized)
		return
	}

	// Busca usuário pelo session_token
	for username, user := range users {
		if user.SessionToken == cookie.Value {

			// Remove session token
			user.SessionToken = ""
			users[username] = user
			break
		}
	}

	// Expira o cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	// Expira o cookie csrf_token também
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	})

	fmt.Fprintln(w, "Logout realizado com sucesso.")
}
