package main

import "time"

type PendingMatch struct {
	TeamA []uint64
	TeamB []uint64
}

type Match struct {
	ID        uint64
	TeamA     []uint64
	TeamB     []uint64
	Winner    int
	Loser     int
	StartTime time.Time
	EndTime   time.Time
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
