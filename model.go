package main

type Player struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Health    int    `json:"health"`
	MaxHealth int    `json:"maxHealth"`
	Damage    int    `json:"damage"`
	Armor     int    `json:"armor"`
	Gold      int    `json:"gold"`
	Weapon    string `json:"weapon"`
	Level     int    `json:"level"`
}

type Boss struct {
	Name      string `json:"name"`
	Health    int    `json:"health"`
	MaxHealth int    `json:"maxHealth"`
	Damage    int    `json:"damage"`
	Gold      int    `json:"gold"`
	Image     string `json:"image"`
	Lore      string `json:"lore"`
	IsActive  bool   `json:"isActive"`
	Defeated  bool   `json:"defeated"`
}

type ShopItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Cost        int    `json:"cost"`
	Damage      int    `json:"damage"`
	Armor       int    `json:"armor"`
	Health      int    `json:"health"`
	Image       string `json:"image"`
}
