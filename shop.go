package main

import "net/http"

func (g *Game) shopHandler(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	g.mu.Lock()
	username := g.sessions[session.Value]
	player := g.players[username]
	g.mu.Unlock()

	g.renderTemplate(w, "shop.html", map[string]interface{}{
		"Player":    player,
		"ShopItems": g.shop,
	})
}

func (g *Game) buyHandler(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	itemName := r.FormValue("item")

	g.mu.Lock()
	defer g.mu.Unlock()

	username := g.sessions[session.Value]
	player := g.players[username]

	for _, item := range g.shop {
		if item.Name == itemName && player.Gold >= item.Cost {
			player.Gold -= item.Cost
			player.Damage += item.Damage
			player.Armor += item.Armor
			player.MaxHealth += item.Health
			player.Health += item.Health
			if item.Damage > 0 {
				player.Weapon = item.Name
			}
			g.savePlayers()
			break
		}
	}

	http.Redirect(w, r, "/shop", http.StatusSeeOther)
}
