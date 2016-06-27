package main

import (
	"sync"
	"time"
)

type Player struct {
	ID        uint64
	CreatedAt time.Time
	Elo       int
	Wins      uint
	Losses    uint
	InGame    bool
	sync.Mutex
}

func (p *Player) StartGame() {
	p.InGame = true
}

func (p *Player) EndGame() {
	p.InGame = false
}
