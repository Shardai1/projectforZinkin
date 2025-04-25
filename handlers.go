package main

import (
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
	username := g.sessions[session.Value]
	player := g.players[username]
	boss := &g.bosses[g.CurrentBossIndex]

	// Проверяем, нужно ли активировать босса
	if !boss.IsActive && !boss.Defeated {
		boss.IsActive = true
		boss.Health = boss.MaxHealth
	}
	g.mu.Unlock()

	g.renderTemplate(w, "index.html", map[string]interface{}{
		"Player": player,
		"Boss":   boss,
	})
}

func (g *Game) attackHandler(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	username := g.sessions[session.Value]
	player := g.players[username]
	boss := &g.bosses[g.CurrentBossIndex]

	// Игрок атакует
	boss.Health -= player.Damage

	// Проверяем смерть босса
	if boss.Health <= 0 {
		boss.Defeated = true
		boss.IsActive = false
		player.Gold += boss.Gold
		player.Level++
	}

	g.savePlayers()
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
		// Усиливаем боссов
		for i := range g.bosses {
			g.bosses[i].MaxHealth = int(float64(g.bosses[i].MaxHealth) * 1.5)
			g.bosses[i].Damage = int(float64(g.bosses[i].Damage) * 1.3)
			g.bosses[i].Gold = int(float64(g.bosses[i].Gold) * 1.2)
		}
		player.Gold += 500
		player.Level++
	}

	// Активируем нового босса
	g.bosses[g.CurrentBossIndex].IsActive = true
	g.bosses[g.CurrentBossIndex].Defeated = false
	g.bosses[g.CurrentBossIndex].Health = g.bosses[g.CurrentBossIndex].MaxHealth

	g.savePlayers()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
