package main

import (
	"time"

	"github.com/jinzhu/gorm"
)

// PendingMatch a match that is pending between two parties
type PendingMatch struct {
	// TeamA the first team
	TeamA []uint64
	// TeamB the second team
	TeamB []uint64
}

// Match a record of a match between two teams
type Match struct {
	ID        uint64    `json:"id,string" gorm:"primary_key;not null;AUTO_INCREMENT;unique"`
	TeamA     []uint64  `json:"teama"`
	TeamB     []uint64  `json:"teamb"`
	Winner    int       `json:"winner"`
	Loser     int       `json:"loser"`
	StartTime time.Time `json:"start"`
	EndTime   time.Time `json:"end"`
	gorm.Model
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
