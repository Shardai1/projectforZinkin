package main

import (
	"encoding/json"
	"html/template"
	"net/http"
)

var templateFuncs = template.FuncMap{
	"firstChar": func(s string) string {
		if len(s) > 0 {
			return string(s[0])
		}
		return "?"
	},
	"percent": func(current, max int) int {
		if max == 0 {
			return 0
		}
		return int(float64(current) / float64(max) * 100)
	},
}

func (g *Game) renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.New(tmpl).Funcs(templateFuncs).ParseFiles("templates/" + tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (g *Game) homeHandler(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	username, ok := g.sessions[session.Value]
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	player, ok := g.players[username]
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	boss := &g.bosses[g.CurrentBossIndex]

	if !boss.IsActive && !boss.Defeated {
		boss.IsActive = true
		boss.Health = boss.MaxHealth
	}

	g.renderTemplate(w, "index.html", map[string]interface{}{
		"Player": player,
		"Boss":   boss,
	})
}

func (g *Game) attackHandler(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	username := g.sessions[session.Value]
	player := g.players[username]

	if player.Health <= 0 {
		http.Error(w, "Player is dead", http.StatusForbidden)
		return
	}

	boss := &g.bosses[g.CurrentBossIndex]
	boss.Health -= player.Damage

	bossDefeated := false
	if boss.Health <= 0 {
		boss.Defeated = true
		boss.IsActive = false
		bossDefeated = true
		player.Gold += boss.Gold
		player.Level++
	}

	g.savePlayers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"bossHealth":      boss.Health,
		"bossMaxHealth":   boss.MaxHealth,
		"playerGold":      player.Gold,
		"bossDefeated":    bossDefeated,
		"playerHealth":    player.Health,
		"playerMaxHealth": player.MaxHealth,
	})
}

func (g *Game) respawnHandler(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	username := g.sessions[session.Value]
	player, exists := g.players[username]
	if !exists {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	if player.IsDead {
		player.Health = player.MaxHealth / 2
		player.IsDead = false
		g.savePlayers()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"health":    player.Health,
			"maxHealth": player.MaxHealth,
		})
		return
	}

	http.Error(w, "Player is not dead", http.StatusBadRequest)
}

func (g *Game) nextBossHandler(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	username := g.sessions[session.Value]
	player := g.players[username]

	g.CurrentBossIndex++
	if g.CurrentBossIndex >= len(g.bosses) {
		g.CurrentBossIndex = 0
		for i := range g.bosses {
			g.bosses[i].MaxHealth = int(float64(g.bosses[i].MaxHealth) * 1.5)
			g.bosses[i].Damage = int(float64(g.bosses[i].Damage) * 1.3)
			g.bosses[i].Gold = int(float64(g.bosses[i].Gold) * 1.2)
		}
		player.Gold += 500
		player.Level++
	}

	g.bosses[g.CurrentBossIndex].IsActive = true
	g.bosses[g.CurrentBossIndex].Defeated = false
	g.bosses[g.CurrentBossIndex].Health = g.bosses[g.CurrentBossIndex].MaxHealth

	g.savePlayers()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (g *Game) playerDataHandler(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	username := g.sessions[session.Value]
	player, exists := g.players[username]
	if !exists {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}
