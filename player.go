package main

import (
	"sync"

	"github.com/jinzhu/gorm"
)

type Player struct {
	ID     uint64 `json:"id" gorm:"not null;primary_key;AUTO_INCREMENT;unique"`
	Elo    int    `json:"elo"`
	Wins   uint   `json:"wins"`
	Losses uint   `json:"losses"`
	InGame bool   `json:"ingame"`
	sync.Mutex
	gorm.Model
}

func (p *Player) StartGame() {
	p.InGame = true
}

func (p *Player) EndGame() {
	p.InGame = false
}
