package main

import "time"

const (
	BasePlayerID = 90000
	IDIncrement  = 32
)

type EloSim struct {
	Players map[uint64]*Player
	BaseElo int
}

func NewEloSim(be int) *EloSim {
	return &EloSim{Players: make(map[uint64]*Player), BaseElo: be}
}

func (es *EloSim) newPlayerID() uint64 {
	dt := time.Now().UnixNano()

	return uint64(dt)
}

func (es *EloSim) AddPlayer(p *Player) uint64 {
	newID := es.newPlayerID()
	p.CreatedAt = time.Now()
	p.Elo = es.BaseElo
	p.ID = newID
	if _, ok := es.Players[newID]; !ok {
		es.Players[newID] = p
	} else {
		es.AddPlayer(p)
	}
	return newID
}
