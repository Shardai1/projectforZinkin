package main

import (
	"net/http"
)

func main() {
	game := NewGame()
	defer func() {
		game.stopTicker <- true
	}()

	http.HandleFunc("/", game.homeHandler)
	http.HandleFunc("/login", game.loginHandler)
	http.HandleFunc("/register", game.registerHandler)
	http.HandleFunc("/logout", game.logoutHandler)
	http.HandleFunc("/attack", game.attackHandler)
	http.HandleFunc("/shop", game.shopHandler)
	http.HandleFunc("/buy", game.buyHandler)
	http.HandleFunc("/next", game.nextBossHandler)
	http.HandleFunc("/ws", game.websocketHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.ListenAndServe(":8080", nil)
}
