package main

import (
	"encoding/json"
	"time"
)

type PendingMatch struct {
	TeamA []uint64
	TeamB []uint64
}

type Match struct {
	ID        uint64    `json:"id,string"`
	TeamA     []uint64  `json:"teama"`
	TeamB     []uint64  `json:"teamb"`
	Winner    int       `json:"winner"`
	Loser     int       `json:"loser"`
	StartTime time.Time `json:"start"`
	EndTime   time.Time `json:"end"`
}

func (m *Match) AddPlayers(teamA uint64, teamB uint64) {
	m.TeamA = append(m.TeamA, teamA)
	m.TeamB = append(m.TeamB, teamB)
}

func (m *Match) TeamAWin() {
	m.Winner = 0
	m.Loser = 1
}

func (m *Match) TeamBWin() {
	m.Winner = 1
	m.Loser = 0
}

func (m *Match) Start() {
	m.StartTime = time.Now()
}

func (m *Match) Stop() {
	m.EndTime = time.Now()
}

func (m *Match) GobEncode() ([]byte, error) {
	mj, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return []byte(mj), nil
}

func (m *Match) GobDecode(data []byte) error {
	err := json.Unmarshal(data, m)
	if err != nil {
		return err
	}
	return nil
}
