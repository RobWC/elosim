package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"sync"
	"time"
)

type Player struct {
	ID        uint64    `json:"id"`
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
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	pj, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	err = enc.Encode(pj)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Player) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&p)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf.Bytes(), p)
}
