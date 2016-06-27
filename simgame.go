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
	StartTime      time.Time
	EndTime        time.Time
}

type playerRequest struct {
	ID  uint64
	ret chan *Player
}

type matchResult struct {
	m    *Match
	done chan struct{}
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
	es.StartTime = time.Now()
}

func (es *EloSim) Stop() {
	es.EndTime = time.Now()
}

// newPlayerID gernerate a new player ID
func (es *EloSim) newPlayerID() uint64 {
	return uint64(len(es.Players))
}

// newMatchID generate a new match ID
func (es *EloSim) newMatchID() uint64 {
	rand.Seed(time.Now().UnixNano() * rand.Int63() / 2)
	dt := time.Now().UnixNano() * rand.Int63()

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
	pa = uint64(rand.Int63n(int64(len(es.Players) - 1)))

	rand.Seed(time.Now().UnixNano() * rand.Int63() * 2)
	pb = uint64(rand.Int63n(int64(len(es.Players) - 1)))

	return &PendingMatch{TeamA: []uint64{pa}, TeamB: []uint64{pb}}
}

func (es *EloSim) SimMatch(pm *PendingMatch) {
	es.wg.Add(1)
	match := &Match{ID: es.newMatchID(), TeamA: pm.TeamA, TeamB: pm.TeamB}
	match.Start()

	if len(match.TeamA) == 1 && len(match.TeamB) == 1 {
		win := false

		pa := es.Players[match.TeamA[0]]
		pb := es.Players[match.TeamB[0]]

		if pa == nil || pb == nil {
			println("Match error", pa, pb, match.TeamA[0], match.TeamB[0])
			return
		}

		// add unique match check
		u := fmt.Sprintf("%X%X", match.TeamA[0], match.TeamB[0])
		es.UniqueMatches[u] = es.UniqueMatches[u] + 1

		if rand.Float64() > calcEloWinChance(pa.Elo, pb.Elo) {
			win = true
			match.TeamAWin()
			pa.Wins = pa.Wins + 1
			pb.Losses = pb.Losses + 1
		} else {
			match.TeamBWin()
			pb.Wins = pb.Wins + 1
			pb.Losses = pa.Losses + 1
		}
		pa.Elo, pb.Elo = eloBattle(pa.Elo, pb.Elo, win)
		// add match history
		es.Players[pa.ID] = pa
		es.Players[pb.ID] = pb

	}

	match.Stop()
	es.MatchHistory[match.ID] = match
}

type EloSimReport struct {
	StartTime        time.Time
	EndTime          time.Time
	TotalPlayers     int
	UniqueMatches    int
	TotalMatches     int
	HighestElo       int
	HighestEloPlayer uint64
	LowestElo        int
	LowestEloPlayer  uint64
	AverageElo       int
	EloBrackets      []int
}

func (es *EloSim) FinalReport() *EloSimReport {
	esr := &EloSimReport{EloBrackets: []int{0, 0, 0}}

	esr.StartTime = es.StartTime
	esr.EndTime = es.EndTime
	esr.TotalPlayers = len(es.Players)
	esr.TotalMatches = len(es.MatchHistory)
	esr.UniqueMatches = len(es.UniqueMatches)

	// determine highest and lowest elo
	for _, v := range es.Players {
		esr.AverageElo = esr.AverageElo + v.Elo

		// check highest elo
		if v.Elo > esr.HighestElo {
			esr.HighestElo = v.Elo
			esr.HighestEloPlayer = v.ID
		}

		if esr.LowestElo > v.Elo || esr.LowestElo == 0 {
			esr.LowestElo = v.Elo
			esr.LowestEloPlayer = v.ID
		}

	}
	esr.AverageElo = esr.AverageElo / len(es.Players)

	// bracket players

	for _, v := range es.Players {
		if v.Elo <= esr.AverageElo {
			esr.EloBrackets[0] = esr.EloBrackets[0] + 1
		} else if v.Elo > esr.AverageElo && v.Elo <= esr.AverageElo+((esr.HighestElo-esr.AverageElo)/3) {
			esr.EloBrackets[1] = esr.EloBrackets[1] + 1
		} else if v.Elo > esr.AverageElo+((esr.HighestElo-esr.AverageElo)/3) {
			esr.EloBrackets[2] = esr.EloBrackets[2] + 1
		}
	}

	return esr
}
