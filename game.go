package main

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WSMessage struct {
	Type      string `json:"type"`
	Health    int    `json:"health,omitempty"`
	MaxHealth int    `json:"maxHealth,omitempty"`
	Damage    int    `json:"damage,omitempty"`
}

type Game struct {
	mu               sync.Mutex
	players          map[string]*Player
	sessions         map[string]string
	bosses           []Boss
	shop             []ShopItem
	CurrentBossIndex int
	ticker           *time.Ticker
	stopTicker       chan bool
	clients          map[*websocket.Conn]bool
	broadcast        chan WSMessage
}

func NewGame() *Game {

	game := &Game{
		players:          make(map[string]*Player),
		sessions:         make(map[string]string),
		bosses:           initBosses(),
		shop:             initShopItems(),
		CurrentBossIndex: 0,
		stopTicker:       make(chan bool),
		clients:          make(map[*websocket.Conn]bool),
		broadcast:        make(chan WSMessage),
	}

	game.loadPlayers()
	game.startGameTicker()
	go game.handleBroadcasts()
	return game
}

func (g *Game) startGameTicker() {
	g.ticker = time.NewTicker(time.Duration(2+rand.Intn(3)) * time.Second)
	go func() {
		for {
			select {
			case <-g.ticker.C:
				g.processBossAttacks()
			case <-g.stopTicker:
				g.ticker.Stop()
				return
			}
		}
	}()
}

func (g *Game) handleBroadcasts() {
	for msg := range g.broadcast {
		g.mu.Lock()
		for client := range g.clients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(g.clients, client)
			}
		}
		g.mu.Unlock()
	}
}

func (g *Game) processBossAttacks() {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, username := range g.sessions {
		player := g.players[username]
		boss := &g.bosses[g.CurrentBossIndex]

		if boss.IsActive && !boss.Defeated && player.Health > 0 {
			if rand.Intn(100) < 60 {
				damageToPlayer := boss.Damage - player.Armor
				if damageToPlayer < 1 {
					damageToPlayer = 1
				}
				player.Health -= damageToPlayer

				if player.Health <= 0 {
					player.Health = 0
					player.IsDead = true

					g.broadcast <- WSMessage{
						Type:      "player_death", // Новый тип сообщения
						Health:    player.Health,
						MaxHealth: player.MaxHealth,
					}
				} else {
					g.broadcast <- WSMessage{
						Type:      "boss_attack",
						Health:    player.Health,
						MaxHealth: player.MaxHealth,
						Damage:    damageToPlayer,
					}
				}
			}
		}
	}
	g.savePlayers()
}

func (g *Game) websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Сохраняем соединение в контексте
	ctx := context.WithValue(r.Context(), "wsConn", conn)
	r = r.WithContext(ctx)

	g.mu.Lock()
	g.clients[conn] = true
	g.mu.Unlock()

	defer func() {
		g.mu.Lock()
		delete(g.clients, conn)
		g.mu.Unlock()
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (g *Game) loadPlayers() {
	data, err := os.ReadFile("players.json")
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic(err)
	}
	if len(data) > 0 {
		json.Unmarshal(data, &g.players)
	}
}

func (g *Game) savePlayers() {
	data, err := json.Marshal(g.players)
	if err != nil {
		panic(err)
	}
	os.WriteFile("players.json", data, 0644)
}

func initBosses() []Boss {
	return []Boss{
		{
			Name:      "Ликантроп",
			Health:    50,
			MaxHealth: 50,
			Damage:    5,
			Gold:      30,
			Image:     "/static/images/bosses/boss1.png",
			Lore:      "Мелкий, но злобный лидер волков",
		},
		{
			Name:      "Защитник Сомов",
			Health:    120,
			MaxHealth: 120,
			Damage:    10,
			Gold:      80,
			Image:     "/static/images/bosses/boss2.png",
			Lore:      "Когда в Краснознаменск приходит тьма, только его крик «Не пройдёте!» заставляет гасторбайтеров спотыкаться о собственные тени.",
		},
		{
			Name:      "Хельмут-Курбатов",
			Health:    200,
			MaxHealth: 200,
			Damage:    20,
			Gold:      150,
			Image:     "/static/images/bosses/boss3.png",
			Lore:      "Он видел рождение звёзд и падение империй. Его борода хранит секреты древних, а взгляд способен заморозить даже самого дерзкого ученика.",
		},
		{
			Name:      "Асылбек-Ассасино",
			Health:    400,
			MaxHealth: 400,
			Damage:    30,
			Gold:      300,
			Image:     "/static/images/bosses/boss4.png",
			Lore:      "Мастер скрытных убийств (потому что в открытом бою не проживет и раунда). Его главное оружие - чтобы жертва не догадалась дыхнуть в его сторону.",
		},
		{
			Name:      "Серега-Нурсултан",
			Health:    400,
			MaxHealth: 400,
			Damage:    40,
			Gold:      300,
			Image:     "/static/images/bosses/boss5.png",
			Lore:      "Бывший таксист из Шымкента, вознёсшийся до божественного статуса.",
		},
		{
			Name:      "Зинкин-Великодушный",
			Health:    4000000000,
			MaxHealth: 4000000000,
			Damage:    400,
			Gold:      1,
			Image:     "/static/images/bosses/boss6.png",
			Lore:      "Его конспекты способны воскрешать павших студентов, когда тьма безнадёги сгущается, он произносит: 'Ладно, пересдача будет', — и дарует ещё одну попытку. ",
		},
	}
}

func initShopItems() []ShopItem {
	return []ShopItem{
		{
			Name:        "Деревянный меч",
			Description: "Простой меч для начинающих",
			Cost:        50,
			Damage:      5,
			Image:       "/static/images/items/sword1.png",
		},
		{
			Name:        "Железный меч",
			Description: "Надежное железное оружие",
			Cost:        150,
			Damage:      15,
			Image:       "/static/images/items/sword2.png",
		},
		{
			Name:        "Кожаная броня",
			Description: "Легкая защита",
			Cost:        100,
			Armor:       5,
			Health:      10,
			Image:       "/static/images/items/armor1.png",
		},
		{
			Name:        "Зелье здоровья",
			Description: "+20 к максимальному здоровью",
			Cost:        200,
			Health:      20,
			Image:       "/static/images/items/potion1.png",
		},
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
