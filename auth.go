package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func (g *Game) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Показываем форму входа
		http.ServeFile(w, r, "templates/login.html")
		return
	}

	// Обработка POST запроса
	username := r.FormValue("username")
	password := r.FormValue("password")

	g.mu.Lock()
	defer g.mu.Unlock()

	player, exists := g.players[username]
	if !exists || !checkPasswordHash(password, player.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Создаем новую сессию
	sessionID := generateSessionID()
	g.sessions[sessionID] = username

	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   sessionID,
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour),
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (g *Game) registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		g.renderTemplate(w, "register.html", nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if password != confirmPassword {
		g.renderTemplate(w, "register.html", map[string]interface{}{
			"Error": "Пароли не совпадают",
		})
		return
	}

	if len(password) < 6 {
		g.renderTemplate(w, "register.html", map[string]interface{}{
			"Error": "Пароль должен содержать минимум 6 символов",
		})
		return
	}

	if len(username) < 3 || len(username) > 20 {
		g.renderTemplate(w, "register.html", map[string]interface{}{
			"Error": "Имя пользователя должно быть от 3 до 20 символов",
		})
		return
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.players[username]; exists {
		g.renderTemplate(w, "register.html", map[string]interface{}{
			"Error": "Имя пользователя уже занято",
		})
		return
	}

	g.players[username] = &Player{
		Username:  username,
		Password:  hashedPassword,
		Health:    100,
		MaxHealth: 100,
		Damage:    1,
		Armor:     0,
		Gold:      10,
		Weapon:    "Кулаки",
		Level:     1,
	}
	g.savePlayers()

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (g *Game) logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
