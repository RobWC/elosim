package main

import (
	"encoding/json"
	"sync"
	"time"
)

type Player struct {
	ID        uint64    `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at"`
	Elo       int       `json:"elo"`
	Wins      uint      `json:"wins"`
	Losses    uint      `json:"losses"`
	InGame    bool      `json:"ingame"`
	sync.Mutex
}

func (p *Player) StartGame() {
	p.InGame = true
}

func (p *Player) EndGame() {
	p.InGame = false
}

func (p *Player) GobEncode() ([]byte, error) {
	pj, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return []byte(pj), nil
}

func (p *Player) GobDecode(data []byte) error {
	err := json.Unmarshal(data, p)
	if err != nil {
		return err
	}
	return nil
}
