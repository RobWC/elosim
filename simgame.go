package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	BasePlayerID = 90000
	IDIncrement  = 32
)

// EloSim an elo simulation system
type EloSim struct {
	Players        map[uint64]*Player
	MatchHistory   map[uint64]*Match
	UniqueMatches  map[string]int
	BaseElo        int
	PendingMatches []*PendingMatch
	matchMaker     func() *PendingMatch
	wg             sync.WaitGroup
}

// Create a new sim and set the base elo with be
func NewEloSim(be int) *EloSim {
	return &EloSim{Players: make(map[uint64]*Player),
		MatchHistory:  make(map[uint64]*Match),
		UniqueMatches: make(map[string]int),
		BaseElo:       be}
}

// Start set elo sim background tasks
func (es *EloSim) Start() {

	es.wg.Add(1)
	go func() {
		defer es.wg.Done()
	}()
}

// newPlayerID gernerate a new player ID
func (es *EloSim) newPlayerID() uint64 {
	return uint64(len(es.Players) + 1)
}

// newMatchID generate a new match ID
func (es *EloSim) newMatchID() uint64 {
	dt := time.Now().UnixNano()

	return uint64(dt)
}

// AddPlayer add an eligible player to the simulation
func (es *EloSim) AddPlayer() uint64 {
	newID := es.newPlayerID()
	p := &Player{}
	p.CreatedAt = time.Now()
	p.Elo = es.BaseElo
	p.ID = newID
	if _, ok := es.Players[newID]; !ok {
		es.Players[newID] = p
	} else {
		es.AddPlayer()
	}
	return newID
}

// SetMatchMaking set the match making method to generate a PendingMatch
func (es *EloSim) SetMatchMaking(mm func() *PendingMatch) {
	es.matchMaker = mm
}

// GenerateMatch crate a new PendingMatch within the simulation
func (es *EloSim) GenerateMatch() *PendingMatch {
	return es.matchMaker()
}

func (es *EloSim) RandomSelectPlayers() *PendingMatch {
	pa, pb := uint64(0), uint64(0)
	rand.Seed(time.Now().UnixNano() * rand.Int63() / 2)
	pa = uint64(rand.Int63n(int64(len(es.Players))))
	rand.Seed(time.Now().UnixNano() * rand.Int63() * 2)
	pb = uint64(rand.Int63n(int64(len(es.Players))))

	return &PendingMatch{TeamA: []uint64{pa}, TeamB: []uint64{pb}}
}

func (es *EloSim) SimMatch(pm *PendingMatch) {
	match := &Match{ID: es.newMatchID(), TeamA: pm.TeamA, TeamB: pm.TeamB}
	match.Start()

	if len(match.TeamA) == 1 && len(match.TeamB) == 1 {
		//		win := false
		pa := es.Players[match.TeamA[0]]
		pb := es.Players[match.TeamB[0]]

		// add unique match check
		u := fmt.Sprintf("%X%X", match.TeamA[0], match.TeamB[0])
		println(u)
		es.UniqueMatches[u] = es.UniqueMatches[u] + 1

		if rand.Float64() > calcEloWinChance(pa.Elo, pb.Elo) {
			//			win = true
			match.TeamAWin()
			//			es.Players[pa.ID].Wins = es.Players[pa.ID].Wins + 1
			//			es.Players[pb.ID].Losses = es.Players[pb.ID].Losses + 1
		} else {
			match.TeamBWin()
			//			es.Players[pb.ID].Wins = es.Players[pb.ID].Wins + 1
			//			es.Players[pb.ID].Losses = es.Players[pa.ID].Losses + 1
		}
		//		es.Players[pa.ID].Elo, es.Players[pb.ID].Elo = eloBattle(es.Players[pa.ID].Elo, es.Players[pb.ID].Elo, win)
	}

	// add match history
	match.Stop()
	//	es.MatchHistory[match.ID] = match
}
